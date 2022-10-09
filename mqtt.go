package main

import (
	"github.com/koestler/go-mqtt-to-influx/config"
	"github.com/koestler/go-mqtt-to-influx/mqttClient"
	"github.com/koestler/go-mqtt-to-influx/statistics"
	"log"
)

func runMqttClient(
	cfg *config.Config,
	statisticsInstance statistics.Statistics,
	initiateShutdown chan<- error,
) (mqttClientPoolInstance *mqttClient.ClientPool) {
	// run pool
	mqttClientPoolInstance = mqttClient.RunPool()

	for _, mqttClientConfig := range cfg.MqttClients() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"mqttClient[%s]: start: Broker='%s', ClientId='%s'",
				mqttClientConfig.Name(), mqttClientConfig.Broker(), mqttClientConfig.ClientId(),
			)
		}

		var client mqttClient.Client
		if mqttClientConfig.ProtocolVersion() == 3 {
			client = mqttClient.CreateV3(mqttClientConfig, statisticsInstance)
		} else {
			client = mqttClient.CreateV5(mqttClientConfig, statisticsInstance)
		}

		mqttClientPoolInstance.AddClient(client)
	}

	return
}
