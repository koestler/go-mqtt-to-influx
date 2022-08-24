package converter

import (
	"encoding/json"
	"log"
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
func goIotdeviceHandler(c Config, tm TopicMatcher, input Input, outputFunc OutputFunc) {
	// use our time
	timeStamp := time.Now()

	// parse topic
	device, err := tm.MatchDevice(input.Topic())
	if err != nil {
		log.Printf("go-iotdevice[%s]: cannot extract device from topic='%s err=%s", c.Name(), input.Topic(), err)
		return
	}

	// parse payload
	var message goIotdeviceTelemetryMessage
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("go-iotdevice[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	if sentClock, err := parseTimeWithZone(message.Time); err == nil {
		outputFunc(stateClockOutputMessage{
			timeStamp: timeStamp,
			device:    device,
			value:     sentClock,
		})
	} else {
		log.Printf("go-iotdevice[%s]: cannot parse time='%s': %s", c.Name(), message.Time, err)
	}

	// only use BMV-700, BlueSolar etc. as sensor variable
	sensor := strings.Split(message.Model, " ")[0]

	if sensor != message.Model {
		// and safe detailed model string as field
		outputFunc(telemetryOutputMessage{
			timeStamp:   timeStamp,
			device:      device,
			field:       "Model",
			unit:        nil,
			sensor:      sensor,
			stringValue: &message.Model,
		})
	}

	for field, value := range message.NumericValues {
		outputFunc(telemetryOutputMessage{
			timeStamp:  timeStamp,
			device:     device,
			field:      field,
			unit:       &value.Unit,
			sensor:     sensor,
			floatValue: &value.Value,
		})
	}

	for field, value := range message.TextValues {
		outputFunc(telemetryOutputMessage{
			timeStamp:   timeStamp,
			device:      device,
			field:       field,
			unit:        nil,
			sensor:      sensor,
			stringValue: &value.Value,
		})
	}
}
