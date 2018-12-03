package converter

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"log"
	"regexp"
	"time"
)

const timeFormat string = "2006-01-02T15:04:05"

type TelemetryMessage struct {
	Time     string
	NextTele string
	TimeZone string
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
			},
			Fields: map[string]interface{}{
				"value": value.Value,
			},
		}
		i += 1
	}

	timeStamp, err := time.Parse(timeFormat, message.Time)
	if err != nil {
		log.Printf("go-ve-sensor: cannot parse timeStamp, err=%v", err)
		return
	}

	converter.influxDbClientInstance.WritePoints(
		"ve-sensor-float",
		"1s",
		points,
		timeStamp,
	)
}
