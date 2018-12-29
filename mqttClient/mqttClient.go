package mqttClient

import (
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"strings"
	"time"
)

type MqttClient struct {
	config Config
	client mqtt.Client
}

type Config interface {
	Name() string
	Broker() string
	User() string
	Password() string
	ClientId() string
	Qos() byte
	AvailabilityTopic() string
	TopicPrefix() string
	LogMessages() bool
}

const (
	OfflinePayload string = "Offline"
	OnlinePayload  string = "Online"
)

func Run(config Config) (mqttClient *MqttClient) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().
		AddBroker(config.Broker()).
		SetClientID(config.ClientId())
	if len(config.User()) > 0 {
		opts.SetUsername(config.User())
	}
	if len(config.Password()) > 0 {
		opts.SetPassword(config.Password())
	}

	opts.MaxReconnectInterval = 30 * time.Second

	// setup availability topic using will
	availableTopic := replaceTemplate(config.AvailabilityTopic(), config)
	log.Printf("mqttClient[%s]: set will to topic='%s', payload='%s'",
		config.Name(), availableTopic, OfflinePayload,
	)
	opts.SetWill(availableTopic, OfflinePayload, config.Qos(), true)

	// start connection
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqttClient[%s]: connect failed: %s", config.Name(), token.Error())
	}
	log.Printf("mqttClient[%s]: connected to broker='%s'", config.Name(), config.Broker())

	// public availability
	log.Printf("mqttClient[%s]: set availability to topic='%s', payload='%s'",
		config.Name(), availableTopic, OnlinePayload,
	)
	client.Publish(availableTopic, config.Qos(), true, OnlinePayload)

	return &MqttClient{
		config: config,
		client: client,
	}
}

func replaceTemplate(template string, config Config) (r string) {
	r = strings.Replace(template, "%Prefix%", config.TopicPrefix(), 1)
	r = strings.Replace(r, "%clientId%", config.ClientId(), 1)
	return
}

func (mq *MqttClient) wrapCallBack(callback mqtt.MessageHandler) mqtt.MessageHandler {
	if !mq.config.LogMessages() {
		return callback
	}

	return func(client mqtt.Client, message mqtt.Message) {
		log.Printf(
			"mqttClient[%s]: received qos=%d: %s %s",
			mq.Name(), message.Qos(), message.Topic(), message.Payload(),
		)
		callback(client, message)
	}
}

func (mq *MqttClient) Subscribe(topic string, callback mqtt.MessageHandler) error {
	log.Printf("mqttClient[%s]: subscribe to topic='%s'", mq.Name(), topic)
	if token := mq.client.Subscribe(topic, mq.config.Qos(), mq.wrapCallBack(callback));
		token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mq *MqttClient) Name() string {
	return mq.config.Name()
}
