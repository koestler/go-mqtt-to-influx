package main

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influx/config"
	"github.com/koestler/go-mqtt-to-influx/converter"
	"github.com/koestler/go-mqtt-to-influx/influxClient"
	"github.com/koestler/go-mqtt-to-influx/mqttClient"
	"github.com/koestler/go-mqtt-to-influx/statistics"
	"github.com/pkg/errors"
	"log"
)

func connectConverters(
	cfg *config.Config,
	statisticsInstance statistics.Statistics,
	mqttClientInstances map[string]*mqttClient.MqttClient,
	influxClientPoolInstance *influxClient.ClientPool,
	initiateShutdown chan<- error,
) {
	countStarted := 0

	for _, converterConfig := range cfg.Converters {
		handleFunc, err := converter.GetHandler(converterConfig.Implementation())
		messageHandler := getMqttMessageHandler(
			converterConfig, handleFunc, statisticsInstance, influxClientPoolInstance,
		)

		if err != nil {
			log.Printf("converter[%s]: cannot start: %s", converterConfig.Name(), err)
			continue
		}

		for _, mqttClientInstance := range getMqttClient(mqttClientInstances, converterConfig.MqttClients()) {
			if cfg.LogWorkerStart {
				log.Printf(
					"converter[%s]: start: Implementation='%s', MqttClient='%s', InfluxClients=%v",
					converterConfig.Name(),
					converterConfig.Implementation(),
					mqttClientInstance.Name(),
					converterConfig.InfluxClients(),
				)
			}

			for _, mqttTopic := range converterConfig.MqttTopics() {
				if err := mqttClientInstance.Subscribe(mqttTopic.Topic(), messageHandler); err != nil {
					log.Printf("converter[%s]: error while subscribing: %s", converterConfig.Name(), err)
				} else {
					countStarted += 1
				}
			}
		}
	}

	if countStarted < 1 {
		initiateShutdown <- errors.New("no converter was started")
	}
}

func getMqttMessageHandler(
	config converter.Config,
	handleFunc converter.HandleFunc,
	statisticsInstance statistics.Statistics,
	influxClientPoolInstance *influxClient.ClientPool,
) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		if config.LogHandleOnce() {
			converter.LogTopicOnce(config.Name(), message)
		}
		statisticsInstance.IncrementOne("converter", config.Name(), message.Topic())
		handleFunc(
			config,
			message,
			func(output converter.Output) {
				influxClientPoolInstance.WritePoint(
					output,
					config.InfluxClients(),
				)
			},
		)
	}
}

func getMqttClient(mqttClientInstances map[string]*mqttClient.MqttClient, clientNames []string) (
	clients []*mqttClient.MqttClient) {
	if len(clientNames) < 1 {
		clients = make([]*mqttClient.MqttClient, len(mqttClientInstances))
		i := 0
		for _, c := range mqttClientInstances {
			clients[i] = c
			i++
		}
		return
	}

	for _, clientName := range clientNames {
		if receiver, ok := mqttClientInstances[clientName]; ok {
			clients = append(clients, receiver)
		}
	}

	return
}
