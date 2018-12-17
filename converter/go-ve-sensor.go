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

func goVeSensorHandler(converter Converter, msg mqtt.Message) {
	// parse topic
	strings := topicMatcher.FindStringSubmatch(msg.Topic())
	if len(strings) < 3 {
		log.Printf("go-ve-sensor: cannot extract device from topic='%s", msg.Topic())
		return
	}
	device := strings[2]

	// parse payload
	var message TelemetryMessage
	if err := json.Unmarshal(msg.Payload(), &message); err != nil {
		log.Printf("go-ve-sensor: cannot json decode: %s", err)
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
		log.Printf("go-ve-sensor: TimeZone='%s' but only UTC is supported", message.TimeZone)
	}

	timeStamp, err := parseTime(message.Time)
	if err != nil {
		timeStamp = time.Now()
	}

	converter.influxDbClientPoolInstance.WritePoints(
		converter.config.TargetMeasurement,
		points,
		timeStamp,
		converter.config.InfluxDbClients,
	)
}
