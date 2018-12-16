package converter

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"github.com/koestler/go-mqtt-to-influxdb/mqttClient"
)

type ConverterHandleFunc func(converter Converter, msg mqtt.Message)

type Converter struct {
	config                 *config.ConverterConfig
	influxDbClientInstance *influxDbClient.InfluxDbClient
}

func RunConverter(
	config *config.ConverterConfig,
	mqttClientInstance *mqttClient.MqttClient,
	influcDbClientInstance *influxDbClient.InfluxDbClient,
) (err error) {
	var handleFunc ConverterHandleFunc

	switch config.Implementation {
	case "go-ve-sensor":
		handleFunc = goVeSensorHandler
	case "tasmota":
		handleFunc = tasmotaHandler
	default:
		return fmt.Errorf("unknown implementation '%s'", config.Implementation)
	}

	converter := Converter{
		config:                 config,
		influxDbClientInstance: influcDbClientInstance,
	}

	for _, mqttTopic := range config.MqttTopics {
		if err := mqttClientInstance.Subscribe(mqttTopic, getMqttMessageHandler(converter, handleFunc)); err != nil {
			return err
		}
	}

	return nil
}

func getMqttMessageHandler(converter Converter, handleFunc ConverterHandleFunc) (mqtt.MessageHandler) {
	return func(client mqtt.Client, message mqtt.Message) {
		handleFunc(converter, message)
	}
}
