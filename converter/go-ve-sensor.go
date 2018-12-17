package converter

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"log"
	"regexp"
	"time"
)

type TelemetryMessage struct {
	Time     string
	NextTele string
	TimeZone string
	Model    string
	Values   map[string]struct {
		Value float64
		Unit  string
	}
}

var topicMatcher = regexp.MustCompile("^(.*)/([^/]*)$")

func goVeSensorHandler(c *Converter, msg mqtt.Message) {
	// parse topic
	matches := topicMatcher.FindStringSubmatch(msg.Topic())
	if len(matches) < 3 {
		log.Printf("go-ve-sensor[%s]: cannot extract device from topic='%s", c.GetName(), msg.Topic())
		return
	}
	device := matches[2]

	// parse payload
	var message TelemetryMessage
	if err := json.Unmarshal(msg.Payload(), &message); err != nil {
		log.Printf("go-ve-sensor[%s]: cannot json decode: %s", c.GetName(), err)
		return
	}

	// map values to points
	points := make([]influxDbClient.Point, len(message.Values))
	i := 0
	for field, value := range message.Values {
		points[i] = influxDbClient.Point{
			Tags: map[string]string{
				"device": device,
				"field":  field,
				"unit":   value.Unit,
				"sensor": message.Model,
			},
			Fields: map[string]interface{}{
				"value": value.Value,
			},
		}
		i += 1
	}

	if message.TimeZone != "UTC" {
		log.Printf("go-ve-sensor[%s]: TimeZone='%s' but only 'UTC' is supported", c.GetName(), message.TimeZone)
	}

	timeStamp, err := parseTime(message.Time)
	if err != nil {
		timeStamp = time.Now()
	}

	c.influxDbClientPoolInstance.WritePoints(
		c.config.TargetMeasurement,
		points,
		timeStamp,
		c.config.InfluxDbClients,
	)
}
