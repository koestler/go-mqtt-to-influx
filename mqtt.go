package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"github.com/koestler/go-mqtt-to-influxdb/mqttClient"
	"github.com/koestler/go-mqtt-to-influxdb/statistics"
	"github.com/pkg/errors"
	"log"
	"os"
)

func runMqttClient(
	cfg *config.Config,
	statisticsInstance *statistics.Statistics,
	initiateShutdown chan<- error,
) map[string]*mqttClient.MqttClient {
	mqtt.ERROR = log.New(os.Stdout, "MqttDebugLog: ", log.LstdFlags)
	if cfg.LogMqttDebug {
		mqtt.DEBUG = log.New(os.Stdout, "MqttDebugLog: ", log.LstdFlags)
	}

	mqttClientInstances := make(map[string]*mqttClient.MqttClient)

	for _, mqttClientConfig := range cfg.MqttClients {
		if cfg.LogWorkerStart {
			log.Printf(
				"main: Mqtt[%s] start, Broker='%s', ClientId='%s'",
				mqttClientConfig.Name(), mqttClientConfig.Broker(), mqttClientConfig.ClientId(),
			)
		}

		if client, err := mqttClient.Run(mqttClientConfig, statisticsInstance); err == nil {
			mqttClientInstances[mqttClientConfig.Name()] = client

			log.Printf("main: Mqtt[%s] started", mqttClientConfig.Name())
		} else {
			log.Printf("main: Mqtt[%s] start failed: %s", mqttClientConfig.Name(), err)
		}
	}

	if len(mqttClientInstances) < 1 {
		initiateShutdown <- errors.New("no mqtt client was started")
	}

	return mqttClientInstances
}
