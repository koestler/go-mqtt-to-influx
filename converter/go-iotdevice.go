package converter

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"
)

type goIotdeviceTelemetryMessage struct {
	Time                   string
	NextTelemetry          string
	Model                  string
	SecondsSinceLastUpdate float64
	NumericValues          map[string]goIotdeviceNumericTelemetryValue
	TextValues             map[string]goIotdeviceTextTelemetryValue
}

type goIotdeviceNumericTelemetryValue struct {
	Value float64
	Unit  string
}

type goIotdeviceTextTelemetryValue struct {
	Value string
}

type goVeSensorOutputMessage struct {
	timeStamp   time.Time
	measurement string
	device      string
	field       string
	unit        *string
	sensor      string
	stringValue *string
	floatValue  *float64
}

// example input: piegn/tele/24v-bmv/state
// -> use 24v-bmv as device identifier
var topicMatcher = regexp.MustCompile("/([^/]*)(/state)?$")

func init() {
	registerHandler("go-iotdevice", goIotdeviceHandler)
}

// parses messages generated by the go-iotdevices tool in the format given by goVeSensorTelemetryMessage
// and write one point per value to the influxdb
// example input:
// {
//  "Time": "2022-08-19T16:19:54Z",
//  "NextTelemetry": "2022-08-19T16:19:59Z",
//  "Model": "BMV-702",
//  "SecondsSinceLastUpdate": 0.631630042,
//  "NumericValues": {
//    "AmountOfChargedEnergy": {
//      "Value": 1883.52,
//      "Unit": "kWh"
//    },
//    "CurrentHighRes": {
//      "Value": -0.58,
//      "Unit": "A"
//    },
//    "NumberOfCycles": {
//      "Value": 241,
//      "Unit": ""
//    },
//    "ProductId": {
//      "Value": 4261544960,
//      "Unit": ""
//    },
//    "SOC": {
//      "Value": 58.16,
//      "Unit": "%"
//    },
//    "TTG": {
//      "Value": 5742,
//      "Unit": "min"
//    },
//    "Uptime": {
//      "Value": 17182790,
//      "Unit": "s"
//    }
//  },
//  "TextValues": {
//    "ModelName": {
//      "Value": "BMV-702"
//    },
//    "SerialNumber": {
//      "Value": "HQ15149CFQI,HQ1515RP6L7,"
//    },
//    "SynchronizationState": {
//      "Value": "true"
//    }
// }
func goIotdeviceHandler(c Config, input Input, outputFunc OutputFunc) {
	// parse topic
	matches := topicMatcher.FindStringSubmatch(input.Topic())
	if len(matches) < 2 {
		log.Printf("go-iotdevice[%s]: cannot extract device from topic='%s", c.Name(), input.Topic())
		return
	}
	device := matches[1]

	// parse payload
	var message goIotdeviceTelemetryMessage
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("go-iotdevice[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	timeStamp, err := parseTimeWithZone(message.Time)
	if err != nil {
		timeStamp = time.Now()
	}

	// only use BMV-700, BlueSolar etc. as sensor variable
	sensor := strings.Split(message.Model, " ")[0]

	if sensor != message.Model {
		// and safe detailed model string as field
		outputFunc(goVeSensorOutputMessage{
			timeStamp:   timeStamp,
			measurement: c.TargetMeasurement(),
			device:      device,
			field:       "Model",
			unit:        nil,
			sensor:      sensor,
			stringValue: &message.Model,
		})
	}

	for field, value := range message.NumericValues {
		outputFunc(goVeSensorOutputMessage{
			timeStamp:   timeStamp,
			measurement: c.TargetMeasurement(),
			device:      device,
			field:       field,
			unit:        &value.Unit,
			sensor:      sensor,
			floatValue:  &value.Value,
		})
	}

	for field, value := range message.TextValues {
		outputFunc(goVeSensorOutputMessage{
			timeStamp:   timeStamp,
			measurement: c.TargetMeasurement(),
			device:      device,
			field:       field,
			unit:        nil,
			sensor:      sensor,
			stringValue: &value.Value,
		})
	}
}

func (m goVeSensorOutputMessage) Measurement() string {
	return m.measurement
}

func (m goVeSensorOutputMessage) Tags() map[string]string {
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

func (m goVeSensorOutputMessage) Fields() (ret map[string]interface{}) {
	ret = make(map[string]interface{}, 2)

	if m.floatValue != nil {
		ret["floatValue"] = *m.floatValue
	}

	if m.stringValue != nil {
		ret["stringValue"] = *m.stringValue
	}

	return
}

func (m goVeSensorOutputMessage) Time() time.Time {
	return m.timeStamp
}
