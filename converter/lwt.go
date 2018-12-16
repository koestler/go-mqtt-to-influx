package converter

import (
	"github.com/eclipse/paho.mqtt.golang"
	"log"
)

func lwtHandler(converter Converter, msg mqtt.Message) {
	log.Printf("lwt: topic='%s' payload='%s'", msg.Topic(), msg.Payload())
}
