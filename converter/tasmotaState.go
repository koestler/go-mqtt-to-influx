package converter

import (
	"encoding/json"
	"log"
	"strings"
	"time"
)

// examples:
// {
//   "Time":"2018-12-16T23:05:14","Uptime":"1T11:32:21","Vcc":3.177,"POWER":"OFF",
//   "wifi":{"AP":1,"SSId":"piegn-iot","BSSId":"04:F0:21:33:40:99","Channel":1,"RSSI":66}
// }
// {
//   "Time":"2018-12-16T23:06:09","Uptime":"8T03:08:26","Vcc":3.112,
//   "POWER1":"ON","POWER2":"OFF","POWER3":"OFF","POWER4":"OFF",
//   "wifi":{"AP":1,"SSId":"piegn-iot","BSSId":"04:F0:21:2F:B7:CC","Channel":1,"RSSI":100}
// }
type tasmotaStateMessage struct {
	Time   string  // save to timeValue
	Uptime string  // save to timeValue
	Vcc    float64 // save to floatValues
	Power  string  // save to boolValues
	Power1 string  // save to boolValues
	Power2 string  // save to boolValues
	Power3 string  // save to boolValues
	Power4 string  // save to boolValues
	Wifi   tasmotaWifi
}

type tasmotaWifi struct {
	AP      int
	SSId    string
	BSSId   string
	Channel int
	RSSI    int
}

type stateWifiOutputMessage struct {
	timeStamp time.Time
	device    string
	wifi      tasmotaWifi
}

func init() {
	registerHandler("tasmota-state", tasmotaStateHandler)
}

func tasmotaStateHandler(c Config, tm TopicMatcher, input Input, outputFunc OutputFunc) {
	// use our time
	timeStamp := time.Now()

	// parse topic
	device, err := tm.MatchDevice(input.Topic())
	if err != nil {
		log.Printf("tasmota-state[%s]: cannot extract device from topic='%s err=%s", c.Name(), input.Topic(), err)
		return
	}

	// parse payload
	var message tasmotaStateMessage
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("tasmota-state[%s]: cannot json decode: %s", c.Name(), err)
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
		log.Printf("tasmota-state[%s]: cannot parse time='%s': %s", c.Name(), message.Time, err)
	}

	sensor := "tasmota"

	// save uptime
	if upTime, err := parseUpTime(message.Uptime); err != nil {
		log.Printf("tasmota-state[%s]: cannot parse uptime='%s': %s", c.Name(), message.Uptime, err)
	} else {
		unit := "s"
		value := float64(upTime)
		outputFunc(telemetryOutputMessage{
			timeStamp:  timeStamp,
			device:     device,
			field:      "UpTime",
			unit:       &unit,
			sensor:     sensor,
			floatValue: &value,
		})
	}

	// Vcc
	{
		unit := "V"
		outputFunc(telemetryOutputMessage{
			timeStamp:  timeStamp,
			device:     device,
			field:      "Vcc",
			unit:       &unit,
			sensor:     sensor,
			floatValue: &message.Vcc,
		})
	}

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
		if value, ok := powerToBoolean(power); ok {
			outputFunc(telemetryOutputMessage{
				timeStamp: timeStamp,
				device:    device,
				field:     field,
				sensor:    sensor,
				boolValue: &value,
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
