package converter

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"log"
	"regexp"
	"time"
)

var lwtTopicMatcher = regexp.MustCompile("^([^/]*/)*tele/(.*)/LWT$")

func lwtHandler(converter Converter, msg mqtt.Message) {
	// parse topic
	strings := lwtTopicMatcher.FindStringSubmatch(msg.Topic())
	if len(strings) < 3 {
		log.Printf("lwt: cannot extract device from topic='%s", msg.Topic())
		return
	}
	device := strings[2]

	// parse payload
	var value bool
	switch string(msg.Payload()) {
	case "Online":
		value = true;
	case "Offline":
		value = false;
	default:
		log.Printf("lwt: unknown LWT value='%s'", msg.Payload())
		return
	}

	// create points
	points := []influxDbClient.Point{
		{
			Tags: map[string]string{
				"device": device,
				"field":  "Available",
			},
			Fields: map[string]interface{}{
				"value": value,
			},
		},
	}

	converter.influxDbClientInstance.WritePoints(
		converter.config.TargetMeasurement,
		points,
		time.Now(),
	)
}
