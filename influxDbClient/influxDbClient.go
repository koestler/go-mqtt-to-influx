package influxDbClient

import (
	influxClient "github.com/influxdata/influxdb/client/v2"
	"log"
	"time"
)

type Client struct {
	config Config
	client influxClient.Client

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

func RunClient(config Config) (client *Client) {
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
		config,
		dbClient,

		make(chan *influxClient.Point, 64),
		getBatch(config),
	}

	// start to send out points
	go client.pointsSender(config.WriteInterval())

	return
}

func (ic *Client) Stop() {
	// Close client resources
	if err := ic.client.Close(); err != nil {
		log.Fatal(err)
	}
}

func (ic *Client) Name() string {
	return ic.config.Name()
}

func (ic *Client) pointsSender(writeInterval time.Duration) {
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

	if ic.config.LogLineProtocol() {
		points := ic.currentBatch.Points()
		if len(points) == 1 {
			log.Printf("influxDbClient[%s]: %s", ic.Name(), points[0].String())
		} else {
			log.Printf("influxDbClient[%s]: write batch of %d points", ic.Name(), len(points))
			for _, p := range points {
				log.Printf("influxDbClient[%s]:   %s", ic.Name(), p.String())
			}
		}
	}

	if err := ic.client.Write(ic.currentBatch); err != nil {
		log.Printf("influxDbClient[%s]: cannot write to db: %s", ic.Name(), err)
		return
	}
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
