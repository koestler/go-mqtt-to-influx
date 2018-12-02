package converter

import (
	"github.com/eclipse/paho.mqtt.golang"
	"log"
)

func tasmotaHandler(converter Converter, msg mqtt.Message) {
	log.Printf("tasmota-converter: %s", msg.Payload())
}