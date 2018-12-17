package converter

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"github.com/koestler/go-mqtt-to-influxdb/mqttClient"
)

type HandleFunc func(converter *Converter, msg mqtt.Message)

type Converter struct {
	config                     config.ConverterConfig
	influxDbClientPoolInstance *influxDbClient.ClientPool
}

var converterImplementations = map[string]HandleFunc{
	"go-ve-sensor":   goVeSensorHandler,
	"lwt":            lwtHandler,
	"tasmota-state":  tasmotaStateHandler,
	"tasmota-sensor": tasmotaSensorHandler,
}

func RunConverter(
	config config.ConverterConfig,
	mqttClientInstance *mqttClient.MqttClient,
	influxDbClientPoolInstance *influxDbClient.ClientPool,
) (err error) {
	var handleFunc HandleFunc

	handleFunc, ok := converterImplementations[config.Implementation]
	if !ok {
		return fmt.Errorf("unknown implementation='%s'", config.Implementation)
	}

	converter := Converter{
		config: config,
		influxDbClientPoolInstance: influxDbClientPoolInstance,
	}

	for _, mqttTopic := range config.MqttTopics {
		if err := mqttClientInstance.Subscribe(mqttTopic, getMqttMessageHandler(&converter, handleFunc)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Converter) GetName() string {
	return c.config.Name
}

func getMqttMessageHandler(converter *Converter, handleFunc HandleFunc) mqtt.MessageHandler {
	if converter.config.LogHandleOnce {
		return func(client mqtt.Client, message mqtt.Message) {
			logTopicOnce(converter.config.Name, message.Topic())
			handleFunc(converter, message)
		}
	}
	return func(client mqtt.Client, message mqtt.Message) {
		handleFunc(converter, message)
	}
}
