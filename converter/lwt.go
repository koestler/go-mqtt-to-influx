package converter

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"log"
	"regexp"
	"time"
)

var lwtTopicMatcher = regexp.MustCompile("^([^/]*/)*tele/(.*)/LWT$")

func lwtHandler(c *Converter, msg mqtt.Message) {
	// parse topic
	matches := lwtTopicMatcher.FindStringSubmatch(msg.Topic())
	if len(matches) < 3 {
		log.Printf("lwt[%s]: cannot extract device from topic='%s", c.GetName(), msg.Topic())
		return
	}
	device := matches[2]

	// parse payload
	var value bool
	switch string(msg.Payload()) {
	case "Online":
		value = true
	case "Offline":
		value = false
	default:
		log.Printf("lwt[%s]: unknown LWT value='%s'", c.GetName(), msg.Payload())
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

	c.influxDbClientPoolInstance.WritePoints(
		c.config.TargetMeasurement,
		points,
		time.Now(),
		c.config.InfluxDbClients,
	)
}
