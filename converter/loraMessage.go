package converter

import "time"

type loraOutputMessage struct {
	timeStamp       time.Time
	device          string
	devEui          string
	gatewayId       string
	gatewayEui      string
	rssi            int64
	channelRssi     int64
	snr             float64
	consumedAirtime time.Duration
	gatewayIdx      int
}

func (m loraOutputMessage) Measurement() string {
	return "lora"
}

func (m loraOutputMessage) Tags() map[string]string {
	ret := map[string]string{
		"device":     m.device,
		"devEui":     m.devEui,
		"gatewayId":  m.gatewayId,
		"gatewayEui": m.gatewayEui,
	}

	return ret
}

func (m loraOutputMessage) Fields() (ret map[string]interface{}) {
	return map[string]interface{}{
		"rssi":              m.rssi,
		"channelRssi":       m.channelRssi,
		"snr":               m.snr,
		"consumedAirtimeUs": m.consumedAirtime.Microseconds(),
		"gatewayIdx":        m.gatewayIdx,
	}
}

func (m loraOutputMessage) Time() time.Time {
	return m.timeStamp
}
