package main

import (
	"github.com/koestler/go-mqtt-to-influx/config"
	"github.com/koestler/go-mqtt-to-influx/influxClient"
	"github.com/koestler/go-mqtt-to-influx/statistics"
	"github.com/pkg/errors"
	"log"
)

func runInfluxClient(
	cfg *config.Config,
	statisticsInstance statistics.Statistics,
	initiateShutdown chan<- error,
) *influxClient.ClientPool {
	influxClientPoolInstance := influxClient.RunPool()

	countStarted := 0

	for _, influxClientConfig := range cfg.InfluxClients {
		if cfg.LogWorkerStart {
			log.Printf(
				"influxClient[%s]: start: url='%s', len(token)=%d, org='%s', bucket='%s'",
				influxClientConfig.Name(),
				influxClientConfig.Url(),
				len(influxClientConfig.Token()),
				influxClientConfig.Org(),
				influxClientConfig.Bucket(),
			)
		}

		client := influxClient.RunClient(influxClientConfig, statisticsInstance)

		influxClientPoolInstance.AddClient(client)
		countStarted += 1
		if cfg.LogWorkerStart {
			log.Printf(
				"influxClient[%s]: started", influxClientConfig.Name(),
			)
		}
	}

	if countStarted < 1 {
		initiateShutdown <- errors.New("no influxClient was started")
	}

	return influxClientPoolInstance
}
