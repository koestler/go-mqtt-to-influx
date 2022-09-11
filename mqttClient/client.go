package mqttClient

import (
	"log"
)

type ClientStruct struct {
	cfg           Config
	statistics    Statistics
	shutdown      chan struct{}
	subscriptions []subscription
}

type subscription struct {
	subscribeTopic string
	messageHandler MessageHandler
}

func (c *ClientStruct) AddRoute(subscribeTopic string, messageHandler MessageHandler) {
	log.Printf("mqttClient[%s]: add route for topic='%s'", c.cfg.Name(), subscribeTopic)

	s := subscription{subscribeTopic: subscribeTopic}
	if c.cfg.LogMessages() {
		s.messageHandler = func(message Message) {
			log.Printf("mqttClient[%s]: received: %s %s", c.cfg.Name(), message.Topic(), message.Payload())
			messageHandler(message)
		}
	} else {
		s.messageHandler = messageHandler
	}
	c.subscriptions = append(c.subscriptions, s)
}

func (c *ClientStruct) Name() string {
	return c.cfg.Name()
}
