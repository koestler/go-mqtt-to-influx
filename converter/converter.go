package converter

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"github.com/koestler/go-mqtt-to-influxdb/mqttClient"
)

type HandleFunc func(c Config, oup Output, msg mqtt.Message)

type Converter struct {
	config                     Config
	influxDbClientPoolInstance *influxDbClient.ClientPool
}

type Config interface {
	Name() string
	Implementation() string
	TargetMeasurement() string
	MqttTopics() []string
	InfluxDbClients() []string
	LogHandleOnce() bool
}

type Output interface {
	WriteRawPoints(rawPoints []influxDbClient.RawPoint, receiverNames []string)
}

var converterImplementations = map[string]HandleFunc{
	"go-ve-sensor":   goVeSensorHandler,
	"lwt":            lwtHandler,
	"tasmota-state":  tasmotaStateHandler,
	"tasmota-sensor": tasmotaSensorHandler,
}

func RunConverter(
	config Config,
	mqttClientInstance *mqttClient.MqttClient,
	influxDbClientPoolInstance *influxDbClient.ClientPool,
) (err error) {
	var handleFunc HandleFunc

	handleFunc, ok := converterImplementations[config.Implementation()]
	if !ok {
		return fmt.Errorf("unknown implementation='%s'", config.Implementation())
	}

	converter := Converter{
		config: config,
		influxDbClientPoolInstance: influxDbClientPoolInstance,
	}

	for _, mqttTopic := range config.MqttTopics() {
		if err := mqttClientInstance.Subscribe(mqttTopic, getMqttMessageHandler(&converter, handleFunc)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Converter) Name() string {
	return c.config.Name()
}

func getMqttMessageHandler(converter *Converter, handleFunc HandleFunc) mqtt.MessageHandler {
	if converter.config.LogHandleOnce() {
		return func(client mqtt.Client, message mqtt.Message) {
			logTopicOnce(converter.Name(), message.Topic())
			handleFunc(converter.config, converter.influxDbClientPoolInstance, message)
		}
	}
	return func(client mqtt.Client, message mqtt.Message) {
		handleFunc(converter.config, converter.influxDbClientPoolInstance, message)
	}
}
