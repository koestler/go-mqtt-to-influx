package mqttClient

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"strings"
	"time"
)

type Client struct {
	config     Config
	client     mqtt.Client
	statistics Statistics
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

type Statistics interface {
	IncrementOne(module, name, field string)
}

const (
	OfflinePayload string = "offline"
	OnlinePayload  string = "online"
)

func Run(config Config, statistics Statistics) (*Client, error) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().
		AddBroker(config.Broker()).
		SetClientID(config.ClientId()).
		SetOrderMatters(false)

	if len(config.User()) > 0 {
		opts.SetUsername(config.User())
	}
	if len(config.Password()) > 0 {
		opts.SetPassword(config.Password())
	}

	opts.SetOrderMatters(false)
	opts.SetCleanSession(false)
	opts.MaxReconnectInterval = 10 * time.Second

	// setup availability topic using will
	availableTopic := getAvailableTopic(config)
	if len(availableTopic) > 0 {
		log.Printf("mqttClient[%s]: set will to topic='%s', payload='%s'",
			config.Name(), availableTopic, OfflinePayload,
		)
		opts.SetWill(availableTopic, OfflinePayload, config.Qos(), true)

		// public availability after each connect
		opts.SetOnConnectHandler(func(client mqtt.Client) {
			sendAvailableMsg(config, client)
		})
	}

	// start connection
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqttClient[%s]: connect failed: %s", config.Name(), token.Error())
	}
	log.Printf("mqttClient[%s]: connected to broker='%s'", config.Name(), config.Broker())

	return &Client{
		config:     config,
		client:     client,
		statistics: statistics,
	}, nil
}

func (mq *Client) Shutdown() {
	sendUnavailableMsg(mq.config, mq.client)
	mq.client.Disconnect(1000)
}

func (mq *Client) ReplaceTemplate(template string) string {
	return replaceTemplate(template, mq.config)
}

func replaceTemplate(template string, config Config) (r string) {
	r = strings.Replace(template, "%Prefix%", config.TopicPrefix(), 1)
	r = strings.Replace(r, "%ClientId%", config.ClientId(), 1)
	return
}

func (mq *Client) wrapCallBack(callback mqtt.MessageHandler, subscribeTopic string) mqtt.MessageHandler {
	if !mq.config.LogMessages() {
		return func(client mqtt.Client, message mqtt.Message) {
			mq.statistics.IncrementOne("mqtt", mq.Name(), subscribeTopic)
			callback(client, message)
		}
	}

	return func(client mqtt.Client, message mqtt.Message) {
		log.Printf(
			"mqttClient[%s]: received qos=%d: %s %s",
			mq.Name(), message.Qos(), message.Topic(), message.Payload(),
		)
		mq.statistics.IncrementOne("mqtt", mq.Name(), subscribeTopic)
		callback(client, message)
	}
}

func (mq *Client) Subscribe(topicWithPlaceholders string, callback mqtt.MessageHandler) (subscribeTopic string, err error) {
	subscribeTopic = replaceTemplate(topicWithPlaceholders, mq.config)
	if token := mq.client.Subscribe(
		subscribeTopic,
		mq.config.Qos(), mq.wrapCallBack(callback, subscribeTopic),
	); token.Wait() && token.Error() != nil {
		err = token.Error()
		return
	}
	return
}

func (mq *Client) Name() string {
	return mq.config.Name()
}

func getAvailableTopic(config Config) string {
	return replaceTemplate(config.AvailabilityTopic(), config)
}

func sendUnavailableMsg(config Config, client mqtt.Client) {
	availableTopic := getAvailableTopic(config)
	if len(availableTopic) < 1 {
		return
	}

	log.Printf("mqttClient[%s]: set availability to topic='%s', payload='%s'",
		config.Name(), availableTopic, OfflinePayload,
	)
	client.Publish(availableTopic, config.Qos(), true, OfflinePayload)
}

func sendAvailableMsg(config Config, client mqtt.Client) {
	availableTopic := getAvailableTopic(config)
	log.Printf("mqttClient[%s]: set availability to topic='%s', payload='%s'",
		config.Name(), availableTopic, OnlinePayload,
	)
	client.Publish(availableTopic, config.Qos(), true, OnlinePayload)
}
