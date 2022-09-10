package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influx/config"
	"github.com/koestler/go-mqtt-to-influx/mqttClient"
	"github.com/koestler/go-mqtt-to-influx/statistics"
	"github.com/pkg/errors"
	"log"
	"os"
)

func runMqttClient(
	cfg *config.Config,
	statisticsInstance statistics.Statistics,
	initiateShutdown chan<- error,
) (mqttClientPoolInstance *mqttClient.ClientPool) {
	// setup logging
	mqtt.ERROR = log.New(os.Stdout, "MqttDebugLog: ", log.LstdFlags)
	if cfg.LogMqttDebug {
		mqtt.DEBUG = log.New(os.Stdout, "MqttDebugLog: ", log.LstdFlags)
	}

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

		if client, err := mqttClient.Run(mqttClientConfig, statisticsInstance); err != nil {
			log.Printf("mqttClient[%s]: start failed: %s", mqttClientConfig.Name(), err)
		} else {
			mqttClientPoolInstance.AddClient(client)

			countStarted += 1
			if cfg.LogWorkerStart {
				log.Printf(
					"mqttClient[%s]: started", mqttClientConfig.Name(),
				)
			}
		}
	}

	if countStarted < 1 {
		initiateShutdown <- errors.New("no mqttClient was started")
	}

	return
}
