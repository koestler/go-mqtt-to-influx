package mqttClient

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/config"
	"log"
	"os"
	"strings"
)

var client mqtt.Client

func Run(config *config.MqttClientConfig) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().AddBroker(config.Broker).SetClientID(config.ClientId)
	if len(config.User) > 0 {
		opts.SetUsername(config.User)
	}
	if len(config.Password) > 0 {
		opts.SetPassword(config.Password)
	}

	availableTopic := replaceTemplate(config.AvailableTopic, config)

	opts.SetWill(availableTopic, "Offline", config.Qos, true)

	mqtt.ERROR = log.New(os.Stdout, "", 0)
	if config.DebugLog {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
	}

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("mqttClient connect failed", token.Error())
	}
	log.Printf("mqttClient: connected to %v", config.Broker)

	client.Publish(availableTopic, config.Qos, true, "Online")
}

func replaceTemplate(template string, config *config.MqttClientConfig) (r string){
	r = strings.Replace(template, "%Prefix%", config.TopicPrefix, 1)
	r = strings.Replace(r, "%ClientId%", config.ClientId, 1)
	return
}
