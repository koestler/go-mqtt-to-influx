package mqttClient

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"time"
)

type ClientV3 struct {
	ClientStruct

	cliOpts *mqtt.ClientOptions
	mc      mqtt.Client
}

func CreateV3(cfg Config, statistics Statistics) (client *ClientV3) {
	client = &ClientV3{
		ClientStruct: createClientStruct(cfg, statistics),
	}

	// configure mqtt library
	client.cliOpts = mqtt.NewClientOptions().
		AddBroker(cfg.Broker().String()).
		SetKeepAlive(cfg.KeepAlive()).
		SetConnectRetryInterval(time.Second).
		SetMaxReconnectInterval(cfg.ConnectRetryDelay()).
		SetConnectTimeout(cfg.ConnectTimeout()).
		SetOnConnectHandler(client.onConnectionUp()).
		SetClientID(cfg.ClientId()).
		SetOrderMatters(false).
		SetCleanSession(true)

	// set logging
	// this is done globally for all mqtt v3 connection since it cannot be done per client
	mqtt.ERROR = log.New(os.Stdout, "MqttClientV3 Error: ", log.LstdFlags)
	if cfg.LogDebug() {
		mqtt.DEBUG = log.New(os.Stdout, "MqttClientV3 Debug: ", log.LstdFlags)
	}

	// configure login
	if user := cfg.User(); len(user) > 0 {
		client.cliOpts.SetUsername(user)
		client.cliOpts.SetPassword(cfg.Password())
	}

	// setup availability topic using will
	if client.AvailabilityEnabled() {
		client.cliOpts.SetWill(client.GetAvailabilityTopic(), availabilityOffline, cfg.Qos(), availabilityRetain)
	}

	return
}

func (c *ClientV3) onConnectionUp() func(client mqtt.Client) {
	return func(client mqtt.Client) {
		log.Printf("mqttClientV3[%s]: connection is up", c.cfg.Name())
		// publish online
		if c.AvailabilityEnabled() {
			go func() {
				client.Publish(c.GetAvailabilityTopic(), c.cfg.Qos(), availabilityRetain, availabilityOnline)
			}()
		}

		// subscribe topics
		c.subscriptionsMutex.RLock()
		defer c.subscriptionsMutex.RUnlock()
		for _, s := range c.subscriptions {
			sub := s
			if token := c.mc.Subscribe(
				s.subscribeTopic,
				c.cfg.Qos(),
				func(c mqtt.Client, m mqtt.Message) {
					sub.messageHandler(Message{
						topic:   m.Topic(),
						payload: m.Payload(),
					})
				},
			); token.Wait() && token.Error() != nil {
				log.Printf("mqttClientV3[%s]: failed to subscribe: %s", c.cfg.Name(), token.Error())
			}
		}
	}
}

func (c *ClientV3) Run() {
	c.mc = mqtt.NewClient(c.cliOpts)
	c.mc.Connect()
}

func (c *ClientV3) Shutdown() {
	close(c.shutdown)

	// publish availability offline
	if c.AvailabilityEnabled() {
		token := c.mc.Publish(c.GetAvailabilityTopic(), c.cfg.Qos(), availabilityRetain, availabilityOffline)
		token.WaitTimeout(time.Second)
	}

	c.mc.Disconnect(1000)

	log.Printf("mqttClientV3[%s]: shutdown completed", c.cfg.Name())
}
