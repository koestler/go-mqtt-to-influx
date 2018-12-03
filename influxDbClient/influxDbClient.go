package influxDbClient

import (
	influxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"log"
	"time"
)

type Point struct {
	Tags   map[string]string
	Fields map[string]interface{}
}

type InfluxDbClient struct {
	config *config.InfluxDbConfig
	client influxClient.Client

	pointToSendChannel chan *influxClient.Point
	currentBatch       influxClient.BatchPoints
}

func Run(config *config.InfluxDbConfig) (influxDbClient *InfluxDbClient) {
	// Create a new HTTPClient
	dbClient, err := influxClient.NewHTTPClient(influxClient.HTTPConfig{
		Addr:     config.Addr,
		Username: config.User,
		Password: config.Password,
	})
	if err != nil {
		log.Fatal(err)
	}

	influxDbClient = &InfluxDbClient{
		config,
		dbClient,
		make(chan *influxClient.Point, 64),
		getBatch(config.Database),
	}

	// start to send out points
	go influxDbClient.pointsSender(config.WriteInterval)

	return
}

func (ic *InfluxDbClient) Stop() {
	// Close client resources
	if err := ic.client.Close(); err != nil {
		log.Fatal(err)
	}
}

func (ic *InfluxDbClient) pointsSender(writeInterval time.Duration) {
	writeTick := time.Tick(writeInterval)

	for {
		select {
		case point := <-ic.pointToSendChannel:
			ic.currentBatch.AddPoint(point)

			// if interval = 0 -> send immediately
			if (writeInterval == 0) {
				ic.sendBatch()
			}
		case <-writeTick:
			ic.sendBatch()
		}
	}
}

func getBatch(database string) (bp influxClient.BatchPoints) {
	bp, err := influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Database:  database,
		Precision: "1s",
	})
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (ic *InfluxDbClient) sendBatch() {
	if len(ic.currentBatch.Points()) < 1 {
		// nothing to send
		return
	}

	if err := ic.client.Write(ic.currentBatch); err != nil {
		log.Printf("cannot write to db, err=%v", err)
		return
	}
	ic.currentBatch = getBatch(ic.config.Database)
}

func (ic *InfluxDbClient) WritePoints(
	measurement string,
	points []Point,
	time time.Time,
) {

	for _, point := range points {
		pt, err := influxClient.NewPoint(measurement, point.Tags, point.Fields, time)
		if err != nil {
			log.Printf("influxDbClient: error=%v", err)
			continue
		}
		ic.pointToSendChannel <- pt
	}
}
