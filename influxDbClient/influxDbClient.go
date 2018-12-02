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

func (ic *InfluxDbClient) WritePoint(
	measurement string,
	tags map[string]string,
	fields map[string]interface{},
	precision string,
	time time.Time,
) {
	bp, err := influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Database:  ic.config.Database,
		Precision: precision,
	})
	if err != nil {
		log.Fatal(err)
	}

	pt, err := influxClient.NewPoint(measurement, tags, fields, time)
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	log.Printf("influxDb: write point %v", pt)

	// Write the batch
	if err := ic.client.Write(bp); err != nil {
		log.Fatal(err)
	}
}
