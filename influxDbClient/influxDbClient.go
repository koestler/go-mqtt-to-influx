package influxDbClient

import (
	influxClient "github.com/influxdata/influxdb/client/v2"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"log"
	"time"
)

var dbClient influxClient.Client
var database string

func Run(config *config.InfluxDbConfig) {

	// Create a new HTTPClient
	var err error
	dbClient, err = influxClient.NewHTTPClient(influxClient.HTTPConfig{
		Addr:     config.Addr,
		Username: config.User,
		Password: config.Password,
	})
	if err != nil {
		log.Fatal(err)
	}
	database = config.Database
}

func Stop() {
	// Close client resources
	if err := dbClient.Close(); err != nil {
		log.Fatal(err)
	}
}

func WritePoint(name string, tags map[string]string, fields map[string]interface{}) {
	if dbClient == nil {
		// connection not ready yet / ignore
		return
	}

	// Create a new point batch
	bp, err := influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	pt, err := influxClient.NewPoint(name, tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	log.Printf("influxDb: write point %v", pt)

	// Write the batch
	if err := dbClient.Write(bp); err != nil {
		log.Fatal(err)
	}
}