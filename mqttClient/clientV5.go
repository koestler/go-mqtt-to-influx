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

type ClientV5 struct {
	ClientStruct

	cliCfg autopaho.ClientConfig
	cm     *autopaho.ConnectionManager
	router *paho.StandardRouter

	ctx    context.Context
	cancel context.CancelFunc
}

func CreateV5(cfg Config, statistics Statistics) (client *ClientV5) {
	ctx, cancel := context.WithCancel(context.Background())

	client = &ClientV5{
		ClientStruct: ClientStruct{
			cfg:        cfg,
			statistics: statistics,
			shutdown:   make(chan struct{}),
		},
		router: paho.NewStandardRouter(),
		ctx:    ctx,
		cancel: cancel,
	}

	client.cliCfg = autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{cfg.Broker()},
		KeepAlive:         uint16(cfg.KeepAlive().Seconds()),
		ConnectRetryDelay: cfg.ConnectRetryDelay(),
		ConnectTimeout:    cfg.ConnectTimeout(),
		OnConnectionUp:    client.onConnectionUp(),
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
	if client.AvailabilityEnabled() {
		client.cliCfg.SetWillMessage(client.GetAvailabilityTopic(), []byte("offline"), cfg.Qos(), true)
	}

	return
}

func (c *ClientV5) onConnectionUp() func(*autopaho.ConnectionManager, *paho.Connack) {
	return func(cm *autopaho.ConnectionManager, conack *paho.Connack) {
		log.Printf("mqttClient[%s]: connection is up", c.cfg.Name())
		// publish online
		if c.AvailabilityEnabled() {
			go func() {
				_, err := cm.Publish(c.ctx, c.availabilityMsg("online"))
				if err != nil {
					log.Printf("mqttClient[%s]: error during publish: %s", c.cfg.Name(), err)
				}
			}()
		}
		// subscribe topics
		if _, err := cm.Subscribe(c.ctx, &paho.Subscribe{
			Subscriptions: func() (ret map[string]paho.SubscribeOptions) {
				ret = make(map[string]paho.SubscribeOptions, len(c.subscriptions))
				for _, s := range c.subscriptions {
					ret[s.subscribeTopic] = paho.SubscribeOptions{QoS: c.cfg.Qos()}
				}
				return
			}(),
		}); err != nil {
			log.Printf("mqttClient[%s]: failed to subscribe: %s", c.cfg.Name(), err)
		}
	}
}

func (c *ClientV5) Run() {
	// add routes to router
	for _, s := range c.subscriptions {
		sub := s
		c.router.RegisterHandler(sub.subscribeTopic, func(p *paho.Publish) {
			sub.messageHandler(Message{
				topic:   p.Topic,
				payload: p.Payload,
			})
		})
	}

	// start connection manager
	var err error
	c.cm, err = autopaho.NewConnection(c.ctx, c.cliCfg)
	if err != nil {
		panic(err) // never happens
	}
}

func (c *ClientV5) Shutdown() {
	close(c.shutdown)

	// publish availability offline
	if c.AvailabilityEnabled() {
		ctx, cancel := context.WithTimeout(c.ctx, time.Second)
		defer cancel()
		_, err := c.cm.Publish(ctx, c.availabilityMsg("offline"))
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

func (c *ClientV5) availabilityMsg(payload string) *paho.Publish {
	return &paho.Publish{
		QoS:     c.cfg.Qos(),
		Topic:   c.GetAvailabilityTopic(),
		Payload: []byte(payload),
		Retain:  true,
	}
}
