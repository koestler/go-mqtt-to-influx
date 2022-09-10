package mqttClient

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"strings"
	"time"
)

type Client struct {
	cfg        Config
	mqttClient mqtt.Client
	statistics Statistics
	shutdown   chan struct{}
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

func Run(cfg Config, statistics Statistics) (*Client, error) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Broker()).
		SetClientID(cfg.ClientId()).
		SetOrderMatters(false).
		SetCleanSession(false). // use persistent session
		SetKeepAlive(10 * time.Second).
		SetMaxReconnectInterval(30 * time.Second)
	if user := cfg.User(); len(user) > 0 {
		opts.SetUsername(user)
	}
	if password := cfg.Password(); len(password) > 0 {
		opts.SetPassword(password)
	}

	// setup availability topic using will
	if availabilityTopic := getAvailabilityTopic(cfg); len(availabilityTopic) > 0 {
		opts.SetWill(availabilityTopic, "offline", cfg.Qos(), true)

		// publish availability after each connect
		opts.SetOnConnectHandler(func(client mqtt.Client) {
			client.Publish(availabilityTopic, cfg.Qos(), true, "online")
		})
	}

	// start connection
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("connect failed: %s", token.Error())
	}
	log.Printf("mqttClient[%s]: connected to broker='%s'", cfg.Name(), cfg.Broker())

	return &Client{
		cfg:        cfg,
		mqttClient: mqttClient,
		statistics: statistics,
		shutdown:   make(chan struct{}),
	}, nil
}

func (c *Client) Shutdown() {
	close(c.shutdown)

	// publish availability offline
	if availabilityTopic := getAvailabilityTopic(c.cfg); len(availabilityTopic) > 0 {
		c.mqttClient.Publish(availabilityTopic, c.cfg.Qos(), true, "offline")
	}

	c.mqttClient.Disconnect(1000)
	log.Printf("mqttClient[%s]: shutdown completed", c.cfg.Name())
}

func (c *Client) ReplaceTemplate(template string) string {
	return replaceTemplate(template, c.cfg)
}

func replaceTemplate(template string, cfg Config) (r string) {
	r = strings.Replace(template, "%Prefix%", cfg.TopicPrefix(), 1)
	r = strings.Replace(r, "%ClientId%", cfg.ClientId(), 1)
	return
}

func (c *Client) wrapCallBack(callback mqtt.MessageHandler, subscribeTopic string) mqtt.MessageHandler {
	if !c.cfg.LogMessages() {
		return func(client mqtt.Client, message mqtt.Message) {
			c.statistics.IncrementOne("mqtt", c.Name(), subscribeTopic)
			callback(client, message)
		}
	}

	return func(client mqtt.Client, message mqtt.Message) {
		log.Printf(
			"mqttClient[%s]: received qos=%d: %s %s",
			c.Name(), message.Qos(), message.Topic(), message.Payload(),
		)
		c.statistics.IncrementOne("mqtt", c.Name(), subscribeTopic)
		callback(client, message)
	}
}

func (c *Client) Subscribe(topicWithPlaceholders string, callback mqtt.MessageHandler) (subscribeTopic string, err error) {
	subscribeTopic = replaceTemplate(topicWithPlaceholders, c.cfg)
	if token := c.mqttClient.Subscribe(
		subscribeTopic,
		c.cfg.Qos(), c.wrapCallBack(callback, subscribeTopic),
	); token.Wait() && token.Error() != nil {
		err = token.Error()
		return
	}
	return
}

func (c *Client) Name() string {
	return c.cfg.Name()
}

func getAvailabilityTopic(config Config) string {
	return replaceTemplate(config.AvailabilityTopic(), config)
}
