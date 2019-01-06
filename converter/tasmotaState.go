package converter

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"
)

// examples:
// {
//   "Time":"2018-12-16T23:05:14","Uptime":"1T11:32:21","Vcc":3.177,"POWER":"OFF",
//   "Wifi":{"AP":1,"SSId":"piegn-iot","BSSId":"04:F0:21:33:40:99","Channel":1,"RSSI":66}
// }
// {
//   "Time":"2018-12-16T23:06:09","Uptime":"8T03:08:26","Vcc":3.112,
//   "POWER1":"ON","POWER2":"OFF","POWER3":"OFF","POWER4":"OFF",
//   "Wifi":{"AP":1,"SSId":"piegn-iot","BSSId":"04:F0:21:2F:B7:CC","Channel":1,"RSSI":100}
// }
type StateMessage struct {
	Time   string  // save to timeValue
	Uptime string  // save to timeValue
	Vcc    float64 // save to floatValues
	Power  string  // save to boolValues
	Power1 string  // save to boolValues
	Power2 string  // save to boolValues
	Power3 string  // save to boolValues
	Power4 string  // save to boolValues
	Wifi   Wifi
}

type Wifi struct {
	AP      int
	SSId    string
	BSSId   string
	Channel int
	RSSI    int
}

type stateWifiOutputMessage struct {
	timeStamp time.Time
	device    string
	wifi      Wifi
}

type stateTimeOutputMessage struct {
	timeStamp time.Time
	device    string
	value     time.Time
}

type stateFloatOutputMessage struct {
	timeStamp time.Time
	device    string
	field     string
	unit      string
	value     float64
}

type stateBoolOutputMessage struct {
	timeStamp time.Time
	device    string
	field     string
	value     bool
}

const tasmotaStateTopicRegexp = "^([^/]*/)*tele/(.*)/STATE$"

var tasmotaStateTopicMatcher = regexp.MustCompile(tasmotaStateTopicRegexp)

func init() {
	registerHandler("tasmota-state", tasmotaStateHandler)
}

func tasmotaStateHandler(c Config, input Input, outputFunc OutputFunc) {
	// parse topic
	matches := tasmotaStateTopicMatcher.FindStringSubmatch(input.Topic())
	if len(matches) < 3 {
		log.Printf("tasmota-state[%s]: cannot extract device from topic='%s', regex='%s'",
			c.Name(), input.Topic(), tasmotaStateTopicRegexp,
		)
		return
	}
	device := matches[2]

	// parse payload
	var message StateMessage
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("tasmota-state[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	// setup timestamps
	now := time.Now()
	timeStamp, err := parseTime(message.Time)
	if err != nil {
		timeStamp = time.Now()
	}

	// save time given vs time now
	outputFunc(stateTimeOutputMessage{
		timeStamp: now,
		device:    device,
		value:     timeStamp,
	})

	// save uptime
	if upTime, err := parseUpTime(message.Uptime); err == nil {
		outputFunc(stateFloatOutputMessage{
			timeStamp: timeStamp,
			device:    device,
			field:     "UpTime",
			unit:      "s",
			value:     float64(upTime),
		})
	} else {
		log.Printf("tasmota-state[%s]: cannot parse uptime='%s': %s", c.Name(), message.Uptime, err)
	}

	// Vcc
	outputFunc(stateFloatOutputMessage{
		timeStamp: timeStamp,
		device:    device,
		field:     "Vcc",
		unit:      "V",
		value:     message.Vcc,
	})

	// Power[1,2,3,4]?
	powerToBoolean := func(power string) (res, ok bool) {
		power = strings.ToUpper(power)
		switch power {
		case "":
			return false, false
		case "ON":
			return true, true
		case "OFF":
			return false, true
		default:
			log.Printf("tasmota-state[%s]: cannot parse POWER='%s': only ON/OFF case-insentive known",
				c.Name(), power,
			)
			return false, false
		}
	}
	outputPower := func(field string, power string) {
		if value, ok := powerToBoolean(message.Power); ok {
			outputFunc(stateBoolOutputMessage{
				timeStamp: timeStamp,
				device:    device,
				field:     field,
				value:     value,
			})
		}
	}
	outputPower("Power", message.Power)
	outputPower("Power1", message.Power1)
	outputPower("Power2", message.Power2)
	outputPower("Power3", message.Power3)
	outputPower("Power4", message.Power4)

	// wifi value
	outputFunc(stateWifiOutputMessage{
		timeStamp: timeStamp,
		device:    device,
		wifi:      message.Wifi,
	})
}

func (m stateWifiOutputMessage) Measurement() string {
	return "wifi"
}

func (m stateWifiOutputMessage) Tags() map[string]string {
	return map[string]string{
		"device": m.device,
		"SSId":   m.wifi.SSId,
		"BSSId":  m.wifi.BSSId,
	}
}

func (m stateWifiOutputMessage) Fields() map[string]interface{} {
	return map[string]interface{}{
		"AP":      m.wifi.AP,
		"Channel": m.wifi.Channel,
		"RSSI":    m.wifi.RSSI,
	}
}

func (m stateWifiOutputMessage) Time() time.Time {
	return m.timeStamp
}

func (m stateTimeOutputMessage) Measurement() string {
	return "timeValue"
}

func (m stateTimeOutputMessage) Tags() map[string]string {
	return map[string]string{
		"device": m.device,
	}
}

func (m stateTimeOutputMessage) Fields() map[string]interface{} {
	return map[string]interface{}{
		"value": m.value,
	}
}

func (m stateTimeOutputMessage) Time() time.Time {
	return m.timeStamp
}

func (m stateFloatOutputMessage) Measurement() string {
	return "floatValue"
}

func (m stateFloatOutputMessage) Tags() map[string]string {
	return map[string]string{
		"device": m.device,
		"field":  m.field,
		"unit":   m.unit,
	}
}

func (m stateFloatOutputMessage) Fields() map[string]interface{} {
	return map[string]interface{}{
		"value": m.value,
	}
}

func (m stateFloatOutputMessage) Time() time.Time {
	return m.timeStamp
}

func (m stateBoolOutputMessage) Measurement() string {
	return "boolValue"
}

func (m stateBoolOutputMessage) Tags() map[string]string {
	return map[string]string{
		"device": m.device,
		"field":  m.field,
	}
}

func (m stateBoolOutputMessage) Fields() map[string]interface{} {
	return map[string]interface{}{
		"value": m.value,
	}
}

func (m stateBoolOutputMessage) Time() time.Time {
	return m.timeStamp
}
