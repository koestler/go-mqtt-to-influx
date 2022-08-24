package converter

import (
	"encoding/json"
	"log"
	"time"
)

type tasmotaSensorTemperatureHumidity struct {
	Temperature *float64
	Humidity    *float64
}

type tasmotaSensorMessage struct {
	Time    string
	AM2301  *tasmotaSensorTemperatureHumidity
	SI7021  *tasmotaSensorTemperatureHumidity
	DS18B20 *struct {
		Temperature float64
	}
	TempUnit string
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
	var message tasmotaSensorMessage
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
		outputFunc(telemetryOutputMessage{
			timeStamp:  timeStamp,
			device:     device,
			field:      field,
			unit:       unit,
			sensor:     sensor,
			floatValue: &value,
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
