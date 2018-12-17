package influxDbClient

import (
	influxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"log"
	"time"
)

type InfluxDbClient struct {
	config config.InfluxDbClientConfig
	client influxClient.Client

	pointToSendChannel chan *influxClient.Point
	currentBatch       influxClient.BatchPoints
}

func RunClient(config config.InfluxDbClientConfig) (client *InfluxDbClient) {
	// Create a new HTTPClient
	dbClient, err := influxClient.NewHTTPClient(influxClient.HTTPConfig{
		Addr:     config.Address,
		Username: config.User,
		Password: config.Password,
	})
	if err != nil {
		log.Fatal(err)
	}

	client = &InfluxDbClient{
		config,
		dbClient,

		make(chan *influxClient.Point, 64),
		getBatch(config),
	}

	// start to send out points
	go client.pointsSender(config.WriteInterval)

	return
}

func (ic *InfluxDbClient) Stop() {
	// Close client resources
	if err := ic.client.Close(); err != nil {
		log.Fatal(err)
	}
}

func (ic *InfluxDbClient) GetName() string {
	return ic.config.Name
}

func (ic *InfluxDbClient) pointsSender(writeInterval time.Duration) {
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

func getBatch(config config.InfluxDbClientConfig) (bp influxClient.BatchPoints) {
	bp, err := influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Database:  config.Database,
		Precision: config.TimePrecision.String(),
	})
	if err != nil {
		log.Fatalf("influxDbClient[%s]: cannot create batch: %s", config.Name, err)
	}
	return
}

func (ic *InfluxDbClient) sendBatch() {
	if len(ic.currentBatch.Points()) < 1 {
		// nothing to send
		return
	}

	if ic.config.LogLineProtocol {
		points := ic.currentBatch.Points()
		if len(points) == 1 {
			log.Printf("influxDbClient[%s]: %s", ic.GetName(), points[0].String())
		} else {
			log.Printf("influxDbClient[%s]: write batch of %d points", ic.GetName(), len(points))
			for _, p := range points {
				log.Printf("influxDbClient[%s]:   %s", ic.GetName(), p.String())
			}
		}
	}

	if err := ic.client.Write(ic.currentBatch); err != nil {
		log.Printf("influxDbClient[%s]: cannot write to db: %s", ic.GetName(), err)
		return
	}
	ic.currentBatch = getBatch(ic.config)
}

func (ic *InfluxDbClient) WritePoints(
	measurement string,
	points Points,
	time time.Time,
) {
	ic.WriteRawPoints(points.ToRaw(measurement, time))
}

func (ic *InfluxDbClient) WriteRawPoints(rawPoints []RawPoint) {
	for _, point := range rawPoints {
		pt, err := influxClient.NewPoint(point.Measurement, point.Tags, point.Fields, point.Time)
		if err != nil {
			log.Printf("influxDbClient[%s]: cannot create point: %s", ic.GetName(), err)
			continue
		}
		ic.pointToSendChannel <- pt
	}
}
