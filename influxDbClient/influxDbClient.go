package influxDbClient

import (
	influxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"log"
	"time"
)

type InfluxDbClient struct {
	config *config.InfluxDbConfig
	client influxClient.Client
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

	return &InfluxDbClient{
		config,
		dbClient,
	}
}

func (ic *InfluxDbClient) Stop() {
	// Close client resources
	if err := ic.client.Close(); err != nil {
		log.Fatal(err)
	}
}

type Point struct {
	Tags   map[string]string
	Fields map[string]interface{}
}

func (ic *InfluxDbClient) WritePoints(
	measurement string,
	precision string,
	points []Point,
	time time.Time,
) {
	bp, err := influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Database:  ic.config.Database,
		Precision: precision,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, point := range points {
		pt, err := influxClient.NewPoint(measurement, point.Tags, point.Fields, time)
		if err != nil {
			log.Printf("influxDbClient: error=%v", err)
			continue
		}
		bp.AddPoint(pt)
		log.Printf("influxDb: write point %v", pt)
	}

	// Write the batch
	if err := ic.client.Write(bp); err != nil {
		log.Printf("influxDbClient: cannot write to db, err=%v",err)
	}
}
