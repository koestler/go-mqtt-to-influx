package mqttClient

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"log"
	"strings"
)

type MqttClient struct {
	config config.MqttClientConfig
	client mqtt.Client
}

const (
	OfflinePayload string = "Offline"
	OnlinePayloapd string = "Online"
)

func Run(config config.MqttClientConfig) (mqttClient *MqttClient) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().AddBroker(config.Broker).SetClientID(config.ClientId)
	if len(config.User) > 0 {
		opts.SetUsername(config.User)
	}
	if len(config.Password) > 0 {
		opts.SetPassword(config.Password)
	}

	// setup availability topic using will
	availableTopic := replaceTemplate(config.AvailabilityTopic, config)
	log.Printf("mqttClient[%s]: set will to topic='%s', payload='%s'",
		config.Name, availableTopic, OfflinePayload,
	)
	opts.SetWill(availableTopic, OfflinePayload, config.Qos, true)

	// start connection
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqttClient[%s]: connect failed: %s", config.Name, token.Error())
	}
	log.Printf("mqttClient[%s]: connected to broker='%s'", config.Name, config.Broker)

	// public availability
	log.Printf("mqttClient[%s]: set availability to topic='%s', payload='%s'",
		config.Name, availableTopic, OnlinePayloapd,
	)
	client.Publish(availableTopic, config.Qos, true, OnlinePayloapd)

	return &MqttClient{
		config: config,
		client: client,
	}
}

func replaceTemplate(template string, config config.MqttClientConfig) (r string) {
	r = strings.Replace(template, "%Prefix%", config.TopicPrefix, 1)
	r = strings.Replace(r, "%ClientId%", config.ClientId, 1)
	return
}

func (mq *MqttClient) wrapCallBack(callback mqtt.MessageHandler) (mqtt.MessageHandler) {
	if !mq.config.LogMessages {
		return callback
	}

	return func(client mqtt.Client, message mqtt.Message) {
		log.Printf(
			"mqttClient[%s]: received Qos=%d: %s %s",
			mq.GetName(), message.Qos(), message.Topic(), message.Payload(),
		)
		callback(client, message)
	}
}

func (mq *MqttClient) Subscribe(topic string, callback mqtt.MessageHandler) (error) {
	log.Printf("mqttClient[%s]: subscribe to topic='%s'", mq.GetName(), topic)
	if token := mq.client.Subscribe(topic, mq.config.Qos, mq.wrapCallBack(callback));
		token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (mq *MqttClient) GetName() string {
	return mq.config.Name
}
