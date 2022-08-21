package converter

import (
	"encoding/json"
	"log"
	"time"
)

type TemperatureHumidity struct {
	Temperature *float64
	Humidity    *float64
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
	timeStamp time.Time
	device    string
	field     string
	unit      *string
	sensor    string
	value     float64
}

func init() {
	registerHandler("tasmota-sensor", tasmotaSensorHandler)
}

func tasmotaSensorHandler(c Config, tm TopicMatcher, input Input, outputFunc OutputFunc) {
	// use our time
	timeStamp := time.Now()

	// parse topic
	device, err := tm.MatchDevice(input.Topic())
	if err != nil {
		log.Printf("tasmota-sensor[%s]: cannot extract device from topic='%s err=%s", c.Name(), input.Topic(), err)
		return
	}

	// parse payload
	var message SensorMessage
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("tasmota-sensor[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	// save clock
	if sentClock, err := parseTime(message.Time); err == nil {
		outputFunc(stateClockOutputMessage{
			timeStamp: timeStamp,
			device:    device,
			value:     sentClock,
		})
	} else {
		log.Printf("tasmota-sensor[%s]: cannot parse time='%s': %s", c.Name(), message.Time, err)
	}

	// send points
	count := 0

	output := func(field string, unit *string, sensor string, value float64) {
		count += 1
		outputFunc(tasmotaSensorOutputMessage{
			timeStamp: timeStamp,
			device:    device,
			field:     field,
			unit:      unit,
			sensor:    sensor,
			value:     value,
		})
	}

	percentStr := "%"

	if message.AM2301 != nil {
		if message.AM2301.Temperature != nil {
			output(
				"Temperature",
				&message.TempUnit,
				"AM2301",
				*message.AM2301.Temperature,
			)
		}
		if message.AM2301.Humidity != nil {
			output(
				"Humidity",
				&percentStr,
				"AM2301",
				*message.AM2301.Humidity,
			)
		}
	}

	if message.SI7021 != nil {
		if message.SI7021.Temperature != nil {
			output(
				"Temperature",
				&message.TempUnit,
				"SI7021",
				*message.SI7021.Temperature,
			)
		}
		if message.SI7021.Humidity != nil {
			output(
				"Humidity",
				&percentStr,
				"SI7021",
				*message.SI7021.Humidity,
			)
		}
	}

	if message.DS18B20 != nil {
		output(
			"Temperature",
			&message.TempUnit,
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
	return "telemetry"
}

func (m tasmotaSensorOutputMessage) Tags() map[string]string {
	ret := map[string]string{
		"sensor": m.sensor,
		"device": m.device,
		"field":  m.field,
	}

	if m.unit != nil {
		ret["unit"] = *m.unit
	}

	return ret
}

func (m tasmotaSensorOutputMessage) Fields() map[string]interface{} {
	return map[string]interface{}{
		"floatValue": m.value,
	}
}

func (m tasmotaSensorOutputMessage) Time() time.Time {
	return m.timeStamp
}
