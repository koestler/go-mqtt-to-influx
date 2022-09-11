package main

import (
	"github.com/koestler/go-mqtt-to-influx/config"
	"github.com/koestler/go-mqtt-to-influx/mqttClient"
	"github.com/koestler/go-mqtt-to-influx/statistics"
	"github.com/pkg/errors"
	"log"
)

func runMqttClient(
	cfg *config.Config,
	statisticsInstance statistics.Statistics,
	initiateShutdown chan<- error,
) (mqttClientPoolInstance *mqttClient.ClientPool) {
	// run pool
	mqttClientPoolInstance = mqttClient.RunPool()

	countStarted := 0
	for _, mqttClientConfig := range cfg.MqttClients {
		if cfg.LogWorkerStart {
			log.Printf(
				"mqttClient[%s]: start: Broker='%s', ClientId='%s'",
				mqttClientConfig.Name(), mqttClientConfig.Broker(), mqttClientConfig.ClientId(),
			)
		}

		client := mqttClient.Create(mqttClientConfig, statisticsInstance)
		mqttClientPoolInstance.AddClient(client)

		client.Run()
		countStarted += 1
	}

	if countStarted < 1 {
		initiateShutdown <- errors.New("no mqttClient was started")
	}

	return
}
