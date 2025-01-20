package converter

import (
	"log"
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
	registerHandler("ttn", ttnHandler("auto"))

	// ttn-dragino is also accepted for backwards compatibility; it is deprecated because now there is a general
	// ttn implementation which automatically detects the sensor type
	registerHandler("ttn-dragino", ttnHandler("auto"))

	// ttn-fencyboy is needed because the fencyboy message does not contain the VersionIDs string
	registerHandler("ttn-fencyboy", ttnHandler("fencyboy"))
}

func ttnHandler(implementation string) func(Config, TopicMatcher, Input, OutputFunc) {
	return func(c Config, tm TopicMatcher, input Input, outputFunc OutputFunc) {
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

		switch implementation {
		case "fencyboy":
			ttnFencyboyHandler(c, device, input, outputFunc)
		case "auto":
			vs := message.UplinkMessage.VersionIds
			if vs == nil {
				log.Printf("ttn[%s]: VersionIds missing", c.Name())
				return
			}

			switch vs.BrandId {
			case "dragino":
				ttnDraginoHandler(c, device, vs.ModelId, input, outputFunc)
			case "sensecap":
				ttnSensecapHandler(c, device, vs.ModelId, input, outputFunc)
			default:
				if vs.BrandId != "" || vs.ModelId != "" {
					log.Printf("ttn[%s]: there is now decoder for brand='%s', model='%s'", c.Name(), vs.BrandId, vs.ModelId)
				}
			}
		}
	}
}
