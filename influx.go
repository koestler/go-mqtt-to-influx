package main

import (
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"github.com/koestler/go-mqtt-to-influxdb/statistics"
	"github.com/pkg/errors"
	"log"
)

func runInfluxClient(
	cfg *config.Config,
	statisticsInstance statistics.Statistics,
	initiateShutdown chan<- error,
) *influxDbClient.ClientPool {
	influxDbClientPoolInstance := influxDbClient.RunPool()

	countStarted := 0

	for _, influxDbClientConfig := range cfg.InfluxDbClients {
		if cfg.LogWorkerStart {
			log.Printf(
				"influxClient[%s]: start: address='%s'",
				influxDbClientConfig.Name(),
				influxDbClientConfig.Address(),
			)
		}

		if client, err := influxDbClient.RunClient(influxDbClientConfig, statisticsInstance); err != nil {
			log.Printf("influxClient[%s]: start failed: %s", influxDbClientConfig.Name(), err)
		} else {
			influxDbClientPoolInstance.AddClient(client)
			countStarted += 1
			if cfg.LogWorkerStart {
				log.Printf(
					"influxClient[%s]: started; serverVersion='%s'",
					influxDbClientConfig.Name(), client.ServerVersion(),
				)
			}
		}
	}

	if countStarted < 1 {
		initiateShutdown <- errors.New("no influxClient was started")
	}

	return influxDbClientPoolInstance
}
