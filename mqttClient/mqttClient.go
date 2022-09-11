package mqttClient

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"log"
	"net/url"
	"time"
)

type Client struct {
	cfg    Config
	cliCfg autopaho.ClientConfig
	cm     *autopaho.ConnectionManager

	statistics Statistics

	ctx      context.Context
	cancel   context.CancelFunc
	shutdown chan struct{}
}

type Config interface {
	Name() string
	Broker() *url.URL
	User() string
	Password() string
	ClientId() string
	Qos() byte
	AvailabilityTopic() string
	TopicPrefix() string
	LogDebug() bool
	LogMessages() bool
}

type Statistics interface {
	IncrementOne(module, name, field string)
}

func Create(cfg Config, statistics Statistics) (client *Client) {
	ctx, cancel := context.WithCancel(context.Background())

	client = &Client{
		cfg:        cfg,
		statistics: statistics,
		ctx:        ctx,
		cancel:     cancel,
		shutdown:   make(chan struct{}),
	}

	availabilityTopic := getAvailabilityTopic(cfg)

	client.cliCfg = autopaho.ClientConfig{
		BrokerUrls: []*url.URL{cfg.Broker()},
		KeepAlive:  5,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, conack *paho.Connack) {
			log.Printf("mqttClient[%s]: connection is up", cfg.Name())
			if len(availabilityTopic) > 0 {
				go func() {
					_, err := cm.Publish(ctx, &paho.Publish{
						QoS:     cfg.Qos(),
						Topic:   availabilityTopic,
						Payload: []byte("online"),
					})
					if err != nil {
						log.Printf("mqttClient[%s]: error during publish: %s", cfg.Name(), err)
					}
				}()
			}
		},
		OnConnectError: func(err error) {
			log.Printf("mqttClient[%s]: connection error: %s", cfg.Name(), err)
		},

		ClientConfig: paho.ClientConfig{
			ClientID:      cfg.ClientId(),
			OnClientError: func(err error) { fmt.Printf("server requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("mqttClient[%s]: server requested disconnect: %s\n", cfg.Name(), d.Properties.ReasonString)
				} else {
					fmt.Printf("mqttClient[%s]: server requested disconnect; reason code: %d\n", cfg.Name(), d.ReasonCode)
				}
			},
		},
	}

	if cfg.LogDebug() {
		prefix := fmt.Sprintf("mqttClient[%s]: ", cfg.Name())
		client.cliCfg.Debug = logger{prefix: prefix + ": autoPaho: "}
		client.cliCfg.PahoDebug = logger{prefix: prefix + "paho: "}
	}

	if user := cfg.User(); len(user) > 0 {
		client.cliCfg.SetUsernamePassword(user, []byte(cfg.Password()))
	}

	// setup availability topic using will
	if len(availabilityTopic) > 0 {
		client.cliCfg.SetWillMessage(availabilityTopic, []byte("offline"), cfg.Qos(), true)
	}

	return
}

func (c *Client) Run() {
	var err error
	c.cm, err = autopaho.NewConnection(c.ctx, c.cliCfg)
	if err != nil {
		panic(err) // never happens
	}
}

func (c *Client) Publish() {

}

func (c *Client) Shutdown() {
	close(c.shutdown)

	// publish availability offline
	if availabilityTopic := getAvailabilityTopic(c.cfg); len(availabilityTopic) > 0 {
		ctx, cancel := context.WithTimeout(c.ctx, time.Second)
		defer cancel()
		_, err := c.cm.Publish(ctx, &paho.Publish{
			QoS:     c.cfg.Qos(),
			Topic:   availabilityTopic,
			Payload: []byte("offline"),
		})
		if err != nil {
			log.Printf("mqttClient[%s]: error during publish: %s", c.cfg.Name(), err)
		}
	}

	ctx, cancel := context.WithTimeout(c.ctx, time.Second)
	defer cancel()
	if err := c.cm.Disconnect(ctx); err != nil {
		log.Printf("mqttClient[%s]: error during disconnect: %s", c.cfg.Name(), err)
	}

	// cancel main context
	c.cancel()

	log.Printf("mqttClient[%s]: shutdown completed", c.cfg.Name())
}

func (c *Client) Subscribe(topicWithPlaceholders string) (subscribeTopic string) {
	subscribeTopic = replaceTemplate(topicWithPlaceholders, c.cfg)

	return
}

func (c *Client) Name() string {
	return c.cfg.Name()
}
