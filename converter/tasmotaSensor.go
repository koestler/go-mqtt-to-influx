package converter

import (
	"encoding/json"
	"log"
	"regexp"
	"time"
)

type TemperatureHumidity struct {
	Temperature float64
	Humidity    float64
}

type SensorMessage struct {
	Time    string
	AM2301  *TemperatureHumidity
	SI7021  *TemperatureHumidity
	DS18B20 *struct {
		Temperature float64
	}
	TempUnit string
}

type tasmotaSensorOutputMessage struct {
	timeStamp   time.Time
	measurement string
	device      string
	field       string
	unit        string
	sensor      string
	value       float64
}

const tasmotaSensorTopicRegex = "^([^/]*/)*tele/(.*)/SENSOR$"

var tasmotaSensorTopicMatcher = regexp.MustCompile(tasmotaSensorTopicRegex)

func init() {
	registerHandler("tasmota-sensor", tasmotaSensorHandler)
}

func tasmotaSensorHandler(c Config, input Input, outputFunc OutputFunc) {
	// parse topic
	matches := tasmotaSensorTopicMatcher.FindStringSubmatch(input.Topic())
	if len(matches) < 3 {
		log.Printf("tasmota-sensor[%s]: cannot extract device from topic='%s', regex='%s'",
			c.Name(), input.Topic(), tasmotaSensorTopicRegex,
		)
		return
	}
	device := matches[2]

	// parse payload
	var message SensorMessage
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("tasmota-sensor[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	// get timestamp
	timeStamp, err := parseTime(message.Time)
	if err != nil {
		timeStamp = time.Now()
	}

	// send points
	count := 0

	output := func(field, unit, sensor string, value float64) {
		count += 1
		outputFunc(tasmotaSensorOutputMessage{
			timeStamp:   timeStamp,
			measurement: c.TargetMeasurement(),
			device:      device,
			field:       field,
			unit:        unit,
			sensor:      sensor,
			value:       value,
		})
	}

	if message.AM2301 != nil {
		output(
			"Temperature",
			message.TempUnit,
			"AM2301",
			message.AM2301.Temperature,
		)
		output(
			"Humidity",
			"%",
			"AM2301",
			message.AM2301.Humidity,
		)
	}

	if message.SI7021 != nil {
		output(
			"Temperature",
			message.TempUnit,
			"SI7021",
			message.SI7021.Temperature,
		)
		output(
			"Humidity",
			"%",
			"SI7021",
			message.SI7021.Humidity,
		)
	}

	if message.DS18B20 != nil {
		output(
			"Temperature",
			message.TempUnit,
			"DS18B20",
			message.DS18B20.Temperature,
		)
	}

	// any points sent?
	if count < 1 {
		log.Printf(
			"tasmota-sensor[%s]: could not extract any sensor data; "+
				"sensor type is probably unknown; known sensors are AM2301, SI7021, DS18B20; payload='%s'",
			c.Name(), input.Payload(),
		)
		return
	}
}

func (m tasmotaSensorOutputMessage) Measurement() string {
	return m.measurement
}

func (m tasmotaSensorOutputMessage) Tags() map[string]string {
	return map[string]string{
		"device": m.device,
		"field":  m.field,
		"unit":   m.unit,
		"sensor": m.sensor,
	}
}

func (m tasmotaSensorOutputMessage) Fields() map[string]interface{} {
	return map[string]interface{}{
		"value": m.value,
	}
}

func (m tasmotaSensorOutputMessage) Time() time.Time {
	return m.timeStamp
}
