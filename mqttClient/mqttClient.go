package mqttClient

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"log"
	"os"
	"strings"
)

type MqttClient struct {
	config *config.MqttClientConfig
	client mqtt.Client
}

func Run(config *config.MqttClientConfig) (mqttClient *MqttClient) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().AddBroker(config.Broker).SetClientID(config.ClientId)
	if len(config.User) > 0 {
		opts.SetUsername(config.User)
	}
	if len(config.Password) > 0 {
		opts.SetPassword(config.Password)
	}

	mqtt.ERROR = log.New(os.Stdout, "", log.LstdFlags)
	if config.DebugLog {
		mqtt.DEBUG = log.New(os.Stdout, "", log.LstdFlags)
	}

	// setup availability topic using will
	availableTopic := replaceTemplate(config.AvailabilityTopic, config)
	opts.SetWill(availableTopic, "Offline", config.Qos, true)

	// start connection
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqttClient: connect failed: %v", token.Error())
	}
	log.Printf("mqttClient: connected to %v", config.Broker)

	// public availability
	client.Publish(availableTopic, config.Qos, true, "Online")

	return &MqttClient{
		config: config,
		client: client,
	}
}

func replaceTemplate(template string, config *config.MqttClientConfig) (r string) {
	r = strings.Replace(template, "%Prefix%", config.TopicPrefix, 1)
	r = strings.Replace(r, "%ClientId%", config.ClientId, 1)
	return
}

func (mq *MqttClient) Subscribe(topic string, callback mqtt.MessageHandler) (error) {
	log.Printf("mqttClient: subscribe to %s", topic)
	if token := mq.client.Subscribe(topic, mq.config.Qos, callback);
		token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
