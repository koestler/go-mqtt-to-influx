package converter

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"github.com/koestler/go-mqtt-to-influxdb/mqttClient"
	"time"
)

type Converter struct {
	config                     Config
	influxDbClientPoolInstance *influxDbClient.ClientPool
	statistics                 Statistics
}

type Config interface {
	Name() string
	Implementation() string
	TargetMeasurement() string
	MqttTopics() []string
	InfluxDbClients() []string
	LogHandleOnce() bool
}

type Statistics interface {
	IncrementOne(module, name, field string)
}

type Input interface {
	Topic() string
	Payload() []byte
}

type Output interface {
	Measurement() string
	Tags() map[string]string
	Fields() map[string]interface{}
	Time() time.Time
}

type OutputFunc func(output Output)
type HandleFunc func(c Config, input Input, outputFunc OutputFunc)

func RunConverter(
	config Config,
	statistics Statistics,
	mqttClientInstance *mqttClient.MqttClient,
	influxDbClientPoolInstance *influxDbClient.ClientPool,
) (err error) {
	handleFunc, err := getHandler(config.Implementation())
	if err != nil {
		return err
	}

	// create new converter object
	converter := Converter{
		config: config,
		influxDbClientPoolInstance: influxDbClientPoolInstance,
		statistics:                 statistics,
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
	return func(client mqtt.Client, message mqtt.Message) {
		if converter.config.LogHandleOnce() {
			logTopicOnce(converter.Name(), message.Topic())
		}
		converter.statistics.IncrementOne("converter", converter.Name(), message.Topic())
		handleFunc(
			converter.config,
			message,
			func(output Output) {
				converter.influxDbClientPoolInstance.WritePoint(
					output,
					converter.config.InfluxDbClients(),
				)
			},
		)
	}
}
