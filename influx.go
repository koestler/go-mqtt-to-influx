package main

import (
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"github.com/koestler/go-mqtt-to-influxdb/influxClient"
	"github.com/koestler/go-mqtt-to-influxdb/statistics"
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
				"influxClient[%s]: start: address='%s'",
				influxClientConfig.Name(),
				influxClientConfig.Address(),
			)
		}

		if client, err := influxClient.RunClient(influxClientConfig, statisticsInstance); err != nil {
			log.Printf("influxClient[%s]: start failed: %s", influxClientConfig.Name(), err)
		} else {
			influxClientPoolInstance.AddClient(client)
			countStarted += 1
			if cfg.LogWorkerStart {
				log.Printf(
					"influxClient[%s]: started; serverVersion='%s'",
					influxClientConfig.Name(), client.ServerVersion(),
				)
			}
		}
	}

	if countStarted < 1 {
		initiateShutdown <- errors.New("no influxClient was started")
	}

	return influxClientPoolInstance
}
