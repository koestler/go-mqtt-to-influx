package converter

import (
	"github.com/eclipse/paho.mqtt.golang"
	"log"
)

func goVeSensorHandler(converter Converter, msg mqtt.Message) {
	log.Printf("go-ve-sensor-converter: %s", msg.Payload())

	/*
	log.Printf("mqtt: %s %s\n", msg.Topic(), msg.Payload())

	// parse topic
	strings := topicMatcher.FindStringSubmatch(msg.Topic())
	if len(strings) < 4 {
		return
	}
	device := strings[2]
	name := strings[3]

	// parse payload
	var measurement Measurement
	if err := json.Unmarshal(msg.Payload(), &measurement); err != nil {
		log.Printf("cannot json decode: %s", err)
		return
	}

	// write to db
	influxDbClient.WritePoint(
		name,
		map[string]string{"device": device},
		measurement.toFields(),
	)
	*/
}

/*
var topicMatcher *regexp.Regexp = regexp.MustCompile("^(.*)/([^/]*)/([^/]*)$")

type Measurement struct {
	Value float64
	Unit  string
}

func (m Measurement) toFields() (map[string]interface{}) {
	return map[string]interface{}{
		"Value": m.Value,
		"Unit":  m.Unit,
	}
}
*/