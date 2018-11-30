package mqttClient

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"log"
	"os"
	"strings"
)

var client mqtt.Client

func Run(config *config.MqttClientConfig, messageHandler func(client mqtt.Client, msg mqtt.Message)) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().AddBroker(config.Broker).SetClientID(config.ClientId)
	if len(config.User) > 0 {
		opts.SetUsername(config.User)
	}
	if len(config.Password) > 0 {
		opts.SetPassword(config.Password)
	}
	opts.SetDefaultPublishHandler(messageHandler)

	mqtt.ERROR = log.New(os.Stdout, "", 0)
	if config.DebugLog {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
	}

	// setup availability topic using will
	availableTopic := replaceTemplate(config.AvailableTopic, config)
	opts.SetWill(availableTopic, "Offline", config.Qos, true)

	// start connection
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("mqttClient connect failed", token.Error())
	}
	log.Printf("mqttClient: connected to %v", config.Broker)

	// public availability
	client.Publish(availableTopic, config.Qos, true, "Online")

	// Subscribe to configured topics
	if token := client.Subscribe("piegn/go-ve-sensor/#", config.Qos, nil);
		token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
}

func replaceTemplate(template string, config *config.MqttClientConfig) (r string) {
	r = strings.Replace(template, "%Prefix%", config.TopicPrefix, 1)
	r = strings.Replace(r, "%ClientId%", config.ClientId, 1)
	return
}
