package converter

import (
	"encoding/json"
	"log"
	"time"
)

// Parse payload of SenseCAP S2120 8-in-1 LoRaWAN Weather Sensor
// Code is compatible with TTN decoder function shown here:
// https://github.com/Seeed-Solution/TTN-Payload-Decoder/blob/master/SenseCAP_S2120_Weather_Station_Decoder.js

type sensecapS2120Message struct {
	EndDeviceIds struct {
		DeviceId string `json:"device_id"`
		DevEui   string `json:"dev_eui"`
	} `json:"end_device_ids"`
	ReceivedAt    time.Time `json:"received_at"`
	UplinkMessage struct {
		DecodedPayload struct {
			Err      int `json:"err"`
			Messages []struct {
				MeasurementId    string  `json:"measurementId"`
				MeasurementValue float64 `json:"measurementValue"`
				Type             string  `json:"type"`
			} `json:"messages"`
			Valid bool `json:"valid"`
		} `json:"decoded_payload"`
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
			ReceivedAt  time.Time `json:"received_at"`
		} `json:"rx_metadata"`
		ConsumedAirtime string `json:"consumed_airtime"`
		VersionIds      struct {
			BrandId string `json:"brand_id"`
			ModelId string `json:"model_id"`
		} `json:"version_ids"`
	} `json:"uplink_message"`
}

func init() {
	registerHandler("ttn-sensecap-s2120", ttnSensecapS2120Handler)
}

func ttnSensecapS2120Handler(c Config, tm TopicMatcher, input Input, outputFunc OutputFunc) {
	// parse topic
	device, err := tm.MatchDevice(input.Topic())
	if err != nil {
		log.Printf("ttn-sensecap-s2120[%s]: cannot extract device from topic='%s err=%s", c.Name(), input.Topic(), err)
		return
	}

	// parse payload
	var message sensecapS2120Message
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("ttn-sensecap-s2120[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	sensor := message.UplinkMessage.VersionIds.ModelId

	// lora
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

	// send points
	count := 0
	for _, m := range message.UplinkMessage.DecodedPayload.Messages {
		count += 1
		outputFunc(telemetryOutputMessage{
			timeStamp: message.ReceivedAt,
			device:    device,
			field:     m.Type,
			unit: func() string {
				switch m.Type {
				case "Air Temperature":
					return "°C"
				case "Air Humidity":
					return "%"
				case "Light Intensity":
					return "lux"

				case "BarometricPressure":
					return "hPa"
				case "WindSpeed":
					return "m/s"
				case "WindDirection":
					return "°"
				case "Rainfall":
					return "mm"

				case "Battery":
					return "V"
				default:
					return ""
				}
			}(),
			sensor:     sensor,
			floatValue: &m.MeasurementValue,
		})
	}

	// any points sent?
	if count < 1 && c.LogDebug() {
		log.Printf("ttn-sensecap-s2120[%s]:could not extract any sensor data; payload='%s'",
			c.Name(), input.Payload(),
		)
		return
	}

}
