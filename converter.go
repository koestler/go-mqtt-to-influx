package main

import (
	"github.com/koestler/go-mqtt-to-influx/config"
	"github.com/koestler/go-mqtt-to-influx/converter"
	"github.com/koestler/go-mqtt-to-influx/influxClient"
	"github.com/koestler/go-mqtt-to-influx/mqttClient"
	"github.com/koestler/go-mqtt-to-influx/statistics"
	"github.com/pkg/errors"
	"log"
)

func createConverters(
	cfg *config.Config,
	statisticsInstance statistics.Statistics,
	mqttClientPoolInstance *mqttClient.ClientPool,
	influxClientPoolInstance *influxClient.ClientPool,
	initiateShutdown chan<- error,
) {
	countCreated := 0

	if cfg.LogWorkerStart {
		log.Print("converter: create converters")
	}

	// iterate through all converters
	for _, converterConfig := range cfg.Converters {
		handleFunc, err := converter.GetHandler(converterConfig.Implementation())
		if err != nil {
			log.Printf("converter[%s]: cannot create: %s", converterConfig.Name(), err)
			continue
		}

		// iterate through all connected mqtt clients
		for _, mqttClientInstance := range mqttClientPoolInstance.GetClientsByNames(converterConfig.MqttClients()) {
			// iterate through all topics
			for _, mqttTopic := range converterConfig.MqttTopics() {
				topicMatcher, err := converter.CreateTopicMatcher(
					mqttTopic.ApplyTopicReplace(mqttClientInstance.ReplaceTemplate),
				)

				if err != nil {
					log.Printf("converter[%s]mqtt[%s]: error: %s", converterConfig.Name(), mqttClientInstance.Name(), err)
					continue
				}

				mqttClientInstance.AddRoute(
					topicMatcher.GetSubscribeTopic(),
					getMqttMessageHandler(
						converterConfig, topicMatcher, handleFunc, statisticsInstance, influxClientPoolInstance,
					),
				)

				if cfg.LogWorkerStart {
					log.Printf(
						"converter[%s]: Implementation='%s', MqttClient='%s', InfluxClients=%v, SubscribeTopic='%s'",
						converterConfig.Name(),
						converterConfig.Implementation(),
						mqttClientInstance.Name(),
						influxClientPoolInstance.GetReceiverClientsNames(converterConfig.InfluxClients()),
						topicMatcher.GetSubscribeTopic(),
					)
				}

				countCreated += 1
			}
		}
	}

	if countCreated < 1 {
		initiateShutdown <- errors.New("no converter was started")
	} else if cfg.LogWorkerStart {
		log.Printf("converter: %d converters created", countCreated)
	}
}

func getMqttMessageHandler(
	config converter.Config,
	topicMatcher converter.TopicMatcher,
	handleFunc converter.HandleFunc,
	statisticsInstance statistics.Statistics,
	influxClientPoolInstance *influxClient.ClientPool,
) mqttClient.MessageHandler {
	return func(message mqttClient.Message) {
		if config.LogHandleOnce() {
			converter.LogTopicOnce(config.Name(), message)
		}
		statisticsInstance.IncrementOne("converter", config.Name(), message.Topic())
		handleFunc(
			config,
			topicMatcher,
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
