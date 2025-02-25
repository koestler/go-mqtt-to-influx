package converter

import (
	"log"
	"strings"
	"time"
)

type ttnMessage struct {
	EndDeviceIds struct {
		DeviceId string `json:"device_id"`
		DevEui   string `json:"dev_eui"`
	} `json:"end_device_ids"`
	ReceivedAt    time.Time `json:"received_at"`
	UplinkMessage struct {
		RxMetadata []struct {
			GatewayIds struct {
				GatewayId string `json:"gateway_id"`
				Eui       string `json:"eui"`
			} `json:"gateway_ids"`
			Time        time.Time `json:"time"`
			Timestamp   int64     `json:"timestamp"`
			Rssi        int64     `json:"rssi"`
			ChannelRssi int64     `json:"channel_rssi"`
			Snr         float64   `json:"snr"`
			UplinkToken string    `json:"uplink_token"`
			GpsTime     time.Time `json:"gps_time"`
			ReceivedAt  time.Time `json:"received_at"`
		} `json:"rx_metadata"`
		ConsumedAirtime string `json:"consumed_airtime"`
		VersionIds      *struct {
			BrandId string `json:"brand_id"`
			ModelId string `json:"model_id"`
		} `json:"version_ids"`
	} `json:"uplink_message"`
}

func init() {
	registerHandler("ttn", ttnHandler)

	// ttn-dragino is also accepted for backwards compatibility; it is depraecated because now there is a general
	// ttn implementation which automatically detects the sensor type
	registerHandler("ttn-dragino", ttnHandler)
}

func ttnHandler(c Config, tm TopicMatcher, input Input, outputFunc OutputFunc) {
	// parse topic
	device, err := tm.MatchDevice(input.Topic())
	if err != nil {
		log.Printf("ttn[%s]: cannot extract device from topic='%s err=%s", c.Name(), input.Topic(), err)
		return
	}

	// parse payload
	var message ttnMessage
	payload := input.Payload()
	if err := json.Unmarshal(payload, &message); err != nil {
		log.Printf("ttn[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	// save general lora data
	for gatewayIdx, rx := range message.UplinkMessage.RxMetadata {
		airtime, err := time.ParseDuration(message.UplinkMessage.ConsumedAirtime)
		if err != nil {
			airtime = 0
		}

		outputFunc(loraOutputMessage{
			timeStamp:       message.ReceivedAt,
			device:          device,
			devEui:          message.EndDeviceIds.DevEui,
			gatewayId:       rx.GatewayIds.GatewayId,
			gatewayEui:      rx.GatewayIds.Eui,
			rssi:            rx.Rssi,
			channelRssi:     rx.ChannelRssi,
			snr:             rx.Snr,
			consumedAirtime: airtime,
			gatewayIdx:      gatewayIdx,
		})
	}

	if vids := message.UplinkMessage.VersionIds; vids != nil {
		switch vids.BrandId {
		case "dragino":
			ttnDraginoHandler(c, device, vids.ModelId, input, outputFunc)
		case "sensecap":
			ttnSensecapHandler(c, device, vids.ModelId, input, outputFunc)
		case "fencyboy":
			ttnFencyboyHandler(c, device, vids.ModelId, input, outputFunc)
		default:
			if vids.BrandId != "" || vids.ModelId != "" {
				log.Printf("ttn[%s]: VersionIds present, but no decoder for brand='%s', model='%s'", c.Name(),
					vids.BrandId, vids.ModelId,
				)
			}
		}
	} else {
		deviceId := message.EndDeviceIds.DeviceId
		if strings.Contains(deviceId, "dragino") ||
			strings.Contains(deviceId, "d20s") { // temporary solution until sensor is in the device registry
			ttnDraginoHandler(c, device, "dragino", input, outputFunc)
		} else if strings.Contains(deviceId, "sensecap") {
			ttnSensecapHandler(c, device, "sensecap", input, outputFunc)
		} else if strings.Contains(deviceId, "fencyboy") {
			ttnFencyboyHandler(c, device, "fencyboy", input, outputFunc)
		} else {
			log.Printf("ttn[%s]: fallback to device_id, but no match for device_id='%s'", c.Name(), deviceId)
		}
	}
}
