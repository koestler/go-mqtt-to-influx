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
	cfg               Config
	cliCfg            autopaho.ClientConfig
	cm                *autopaho.ConnectionManager
	router            *paho.StandardRouter
	subscribeTopics   []string
	availabilityTopic string

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
		cfg:               cfg,
		router:            paho.NewStandardRouter(),
		availabilityTopic: getAvailabilityTopic(cfg),
		statistics:        statistics,
		ctx:               ctx,
		cancel:            cancel,
		shutdown:          make(chan struct{}),
	}

	client.cliCfg = autopaho.ClientConfig{
		BrokerUrls:     []*url.URL{cfg.Broker()},
		KeepAlive:      5,
		OnConnectionUp: client.onConnectionUp(),
		OnConnectError: func(err error) {
			log.Printf("mqttClient[%s]: connection error: %s", cfg.Name(), err)
		},
		ClientConfig: paho.ClientConfig{
			ClientID: cfg.ClientId(),
			Router:   client.router,
		},
	}

	if cfg.LogDebug() {
		prefix := fmt.Sprintf("mqttClient[%s]: ", cfg.Name())
		client.cliCfg.Debug = logger{prefix: prefix + "autoPaho: "}
		client.cliCfg.PahoDebug = logger{prefix: prefix + "paho: "}
	}

	if user := cfg.User(); len(user) > 0 {
		client.cliCfg.SetUsernamePassword(user, []byte(cfg.Password()))
	}

	// setup availability topic using will
	if len(client.availabilityTopic) > 0 {
		client.cliCfg.SetWillMessage(client.availabilityTopic, []byte("offline"), cfg.Qos(), true)
	}

	return
}

func (c *Client) onConnectionUp() func(*autopaho.ConnectionManager, *paho.Connack) {
	return func(cm *autopaho.ConnectionManager, conack *paho.Connack) {
		log.Printf("mqttClient[%s]: connection is up", c.cfg.Name())
		// publish online
		if len(c.availabilityTopic) > 0 {
			go func() {
				_, err := cm.Publish(c.ctx, &paho.Publish{
					QoS:     c.cfg.Qos(),
					Topic:   c.availabilityTopic,
					Payload: []byte("online"),
				})
				if err != nil {
					log.Printf("mqttClient[%s]: error during publish: %s", c.cfg.Name(), err)
				}
			}()
		}
		// subscribe topics
		if _, err := cm.Subscribe(c.ctx, &paho.Subscribe{
			Subscriptions: func() (ret map[string]paho.SubscribeOptions) {
				ret = make(map[string]paho.SubscribeOptions, len(c.subscribeTopics))
				for _, t := range c.subscribeTopics {
					ret[t] = paho.SubscribeOptions{QoS: c.cfg.Qos()}
				}
				return
			}(),
		}); err != nil {
			log.Printf("mqttClient[%s]: failed to subscribe: %s", err)
		}
	}
}

func (c *Client) AddRoute(subscribeTopic string, messageHandler MessageHandler) {
	log.Printf("mqttClient[%s]: add route for topic='%s'", c.cfg.Name(), subscribeTopic)

	var handler func(*paho.Publish)
	if c.cfg.LogMessages() {
		handler = func(p *paho.Publish) {
			if c.cfg.LogMessages() {
				log.Printf("mqttClient[%s]: received: %v", c.cfg.Name(), p)
			}
			messageHandler(Message{
				topic:   p.Topic,
				payload: p.Payload,
			})
		}
	} else {
		handler = func(p *paho.Publish) {
			messageHandler(Message{
				topic:   p.Topic,
				payload: p.Payload,
			})
		}
	}

	c.router.RegisterHandler(subscribeTopic, handler)
	c.subscribeTopics = append(c.subscribeTopics, subscribeTopic)
}

func (c *Client) Run() {
	var err error
	c.cm, err = autopaho.NewConnection(c.ctx, c.cliCfg)
	if err != nil {
		panic(err) // never happens
	}
}

func (c *Client) Shutdown() {
	close(c.shutdown)

	// publish availability offline
	if len(c.availabilityTopic) > 0 {
		ctx, cancel := context.WithTimeout(c.ctx, time.Second)
		defer cancel()
		_, err := c.cm.Publish(ctx, &paho.Publish{
			QoS:     c.cfg.Qos(),
			Topic:   c.availabilityTopic,
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

func (c *Client) Name() string {
	return c.cfg.Name()
}
