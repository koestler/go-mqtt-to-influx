package converter

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
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
	Vcc    float32 // save to floatValues
	Power  string  // save to boolValues
	Power1 string  // save to boolValues
	Power2 string  // save to boolValues
	Power3 string  // save to boolValues
	Power4 string  // save to boolValues
	Wifi   struct {
		// -> wifi
		AP      int
		SSId    string
		BSSId   string
		Channel int
		RSSI    int
	}
}

var tasmotaStateTopicMatcher = regexp.MustCompile("^([^/]*/)*tele/(.*)/STATE$")

func tasmotaStateHandler(c *Converter, msg mqtt.Message) {
	// parse topic
	matches := tasmotaStateTopicMatcher.FindStringSubmatch(msg.Topic())
	if len(matches) < 3 {
		log.Printf("tasmota-state[%s]: cannot extract device from topic='%s", c.GetName(), msg.Topic())
		return
	}
	device := matches[2]

	// parse payload
	var message StateMessage
	if err := json.Unmarshal(msg.Payload(), &message); err != nil {
		log.Printf("tasmota-state[%s]: cannot json decode: %s", c.GetName(), err)
		return
	}

	// create points
	rawPoints := message.toPoints(c.GetName(), device)
	c.influxDbClientPoolInstance.WriteRawPoints(
		rawPoints,
		c.config.InfluxDbClients,
	)
}

func (v StateMessage) toPoints(converterName, device string) []influxDbClient.RawPoint {
	ret := make([]influxDbClient.RawPoint, 0, 16)

	// setup timestamps
	now := time.Now()
	timeStamp, err := parseTime(v.Time)
	if err != nil {
		timeStamp = time.Now()
	}

	// save time given vs time now
	ret = append(ret, influxDbClient.RawPoint{
		Measurement: "timeValue",
		Tags: map[string]string{
			"device": device,
		},
		Fields: map[string]interface{}{
			"value": timeStamp,
		},
		Time: now,
	})

	// save uptime
	if upTime, err := parseUpTime(v.Uptime); err == nil {
		ret = append(ret, influxDbClient.RawPoint{
			Measurement: "floatValue",
			Tags: map[string]string{
				"device": device,
				"field":  "UpTime",
				"unit":   "s",
			},
			Fields: map[string]interface{}{
				"value": float64(upTime),
			},
			Time: timeStamp,
		})
	} else {
		log.Printf("tasmota-state[%s]: cannot parse uptime='%s': %s", converterName, v.Uptime, err)
	}

	// Vcc
	ret = append(ret, influxDbClient.RawPoint{
		Measurement: "floatValue",
		Tags: map[string]string{
			"device": device,
			"field":  "Vcc",
			"unit":   "V",
		},
		Fields: map[string]interface{}{
			"value": v.Vcc,
		},
		Time: timeStamp,
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
				converterName, power,
			)
			return false, false
		}
	}

	if value, ok := powerToBoolean(v.Power); ok {
		ret = append(ret, powerPoint(device, value, timeStamp))
	}
	if value, ok := powerToBoolean(v.Power1); ok {
		ret = append(ret, powerPoint(device, value, timeStamp))
	}
	if value, ok := powerToBoolean(v.Power2); ok {
		ret = append(ret, powerPoint(device, value, timeStamp))
	}
	if value, ok := powerToBoolean(v.Power3); ok {
		ret = append(ret, powerPoint(device, value, timeStamp))
	}
	if value, ok := powerToBoolean(v.Power4); ok {
		ret = append(ret, powerPoint(device, value, timeStamp))
	}

	// wifi value
	ret = append(ret, influxDbClient.RawPoint{
		Measurement: "wifi",
		Tags: map[string]string{
			"device": device,
			"SSId":   v.Wifi.SSId,
			"BSSId":  v.Wifi.BSSId,
		},
		Fields: map[string]interface{}{
			"AP":      v.Wifi.AP,
			"Channel": v.Wifi.Channel,
			"RSSI":    v.Wifi.RSSI,
		},
		Time: timeStamp,
	})

	return ret
}

func powerPoint(device string, value bool, timeStamp time.Time) influxDbClient.RawPoint {
	return influxDbClient.RawPoint{
		Measurement: "boolValue",
		Tags: map[string]string{
			"device": device,
			"field":  "Power",
		},
		Fields: map[string]interface{}{
			"value": value,
		},
		Time: timeStamp,
	}
}
