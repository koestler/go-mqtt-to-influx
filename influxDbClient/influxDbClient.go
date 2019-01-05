package influxDbClient

import (
	influxClient "github.com/influxdata/influxdb/client/v2"
	"log"
	"strings"
	"time"
)

const ErrorDelayMin = time.Second
const ErrorDelayMax = time.Minute

type Client struct {
	config     Config
	client     influxClient.Client
	statistics Statistics

	lastTransmission time.Time
	errorRetryDelay  time.Duration

	shutdown           chan struct{}
	closed             chan struct{}
	pointToSendChannel chan *influxClient.Point
	currentBatch       influxClient.BatchPoints
}

type Config interface {
	Name() string
	Address() string
	User() string
	Password() string
	Database() string
	WriteInterval() time.Duration
	TimePrecision() time.Duration
	LogLineProtocol() bool
}

type Statistics interface {
	IncrementOne(module, name, field string)
}

func RunClient(config Config, statistics Statistics) (client *Client) {
	// Create a new HTTPClient
	dbClient, err := influxClient.NewHTTPClient(influxClient.HTTPConfig{
		Addr:     config.Address(),
		Username: config.User(),
		Password: config.Password(),
	})
	if err != nil {
		log.Fatal(err)
	}

	client = &Client{
		config:     config,
		client:     dbClient,
		statistics: statistics,

		shutdown:           make(chan struct{}),
		closed:             make(chan struct{}),
		pointToSendChannel: make(chan *influxClient.Point, 64),
		currentBatch:       getBatch(config),
	}

	// start to send out points
	go client.pointsSender(config.WriteInterval())

	return
}

func (ic *Client) Shutdown() {
	// send remaining points
	close(ic.shutdown)
	// wait for pointsSender to shut down
	<-ic.closed

	// shutdown client ressources
	if err := ic.client.Close(); err != nil {
		log.Fatal(err)
	}
}

func (ic *Client) Name() string {
	return ic.config.Name()
}

func (ic *Client) pointsSender(writeInterval time.Duration) {
	defer close(ic.closed)

	writeTick := time.Tick(writeInterval)

	for {
		select {
		case point := <-ic.pointToSendChannel:
			ic.currentBatch.AddPoint(point)

			// if interval = 0 -> send immediately
			if writeInterval == 0 {
				ic.sendBatch()
			}
		case <-writeTick:
			ic.sendBatch()
		case <-ic.shutdown:
			ic.sendBatch()
			return // shutdown
		}

	}
}

func getBatch(config Config) (bp influxClient.BatchPoints) {
	bp, err := influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Database:  config.Database(),
		Precision: config.TimePrecision().String(),
	})
	if err != nil {
		log.Fatalf("influxDbClient[%s]: cannot create batch: %s", config.Name(), err)
	}
	return
}

func (ic *Client) sendBatch() {
	if len(ic.currentBatch.Points()) < 1 {
		// nothing to send
		return
	}

	if time.Now().Before(ic.lastTransmission.Add(ic.errorRetryDelay)) {
		// in retransmission backoff: keep data & return
		return
	}

	points := ic.currentBatch.Points()
	if ic.config.LogLineProtocol() {
		if len(points) == 1 {
			log.Printf("influxDbClient[%s]: %s", ic.Name(), points[0].String())
		} else {
			log.Printf("influxDbClient[%s]: write batch of %d points", ic.Name(), len(points))
			for _, p := range points {
				log.Printf("influxDbClient[%s]:   %s", ic.Name(), p.String())
			}
		}
	}
	// update statistics
	for _, p := range points {
		ic.statistics.IncrementOne("influxDb", ic.Name(), strings.Split(p.String(), " ")[0])
	}

	ic.lastTransmission = time.Now()
	if err := ic.client.Write(ic.currentBatch); err == nil {
		// all ok
		ic.errorRetryDelay = 0
	} else {
		// retry with exponential backoff
		if ic.errorRetryDelay < 1 {
			ic.errorRetryDelay = ErrorDelayMin
		} else {
			ic.errorRetryDelay *= 2
		}
		if ic.errorRetryDelay > ErrorDelayMax {
			ic.errorRetryDelay = ErrorDelayMax
		}

		log.Printf("influxDbClient[%s]: cannot write to db; retry no ealier than %s; err: %s",
			ic.Name(),
			ic.lastTransmission.Add(ic.errorRetryDelay).Format(time.UnixDate),
			err,
		)

		// keep current batch for retransmission
		return
	}

	// flush batch / start a new one
	ic.currentBatch = getBatch(ic.config)
}

func (ic *Client) WritePoints(
	measurement string,
	points Points,
	time time.Time,
) {
	ic.WriteRawPoints(points.ToRaw(measurement, time))
}

func (ic *Client) WriteRawPoints(rawPoints []RawPoint) {
	for _, point := range rawPoints {
		pt, err := influxClient.NewPoint(point.Measurement, point.Tags, point.Fields, point.Time)
		if err != nil {
			log.Printf("influxDbClient[%s]: cannot create point: %s", ic.Name(), err)
			continue
		}
		ic.pointToSendChannel <- pt
	}
}
