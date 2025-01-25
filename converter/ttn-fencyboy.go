package converter

import (
	"log"
	"time"
)

type ttnFencyboyMessage struct {
	ReceivedAt    time.Time `json:"received_at"`
	UplinkMessage struct {
		DecodedPayload struct {
			ActiveMode        *bool    `json:"ACTIVE_MODE"`
			Impulses          *int64   `json:"IMPULSES"`
			FenceVoltage      *float64 `json:"FENCEVOLTAGE"`
			FenceVoltageStd   *float64 `json:"FENCE_VOLTAGE_STD"`
			FenceVoltageMin   *float64 `json:"FENCEVOLTAGEMIN"`
			FenceVoltageMax   *float64 `json:"FENCEVOLTAGEMAX"`
			BatteryVoltage    *float64 `json:"BATTERYVOLTAGE"`
			RemainingCapacity *float64 `json:"REMAINING_CAPACITY"`
			Temperature       *float64 `json:"TEMPERATURE"`
		} `json:"decoded_payload"`
	} `json:"uplink_message"`
}

func ttnFencyboyHandler(c Config, device, model string, input Input, outputFunc OutputFunc) {
	// parse payload
	var message ttnFencyboyMessage
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("ttn-fencyboy[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	// send points
	count := 0
	output := func(field string, unit string, stringValue *string, floatValue *float64, boolValue *bool, intValue *int64) {
		count += 1
		outputFunc(telemetryOutputMessage{
			timeStamp: message.ReceivedAt,
			device:    device,
			field:     field,
			unit: func(u string) *string {
				if len(u) > 0 {
					return &u
				}
				return nil
			}(unit),
			sensor:      model,
			stringValue: stringValue,
			floatValue:  floatValue,
			boolValue:   boolValue,
			intValue:    intValue,
		})
	}
	outputFloat := func(field string, unit string, floatValue *float64) {
		output(field, unit, nil, floatValue, nil, nil)
	}

	outputBool := func(field string, unit string, boolValue *bool) {
		output(field, unit, nil, nil, boolValue, nil)
	}
	outputInt := func(field string, unit string, intValue *int64) {
		output(field, unit, nil, nil, nil, intValue)
	}

	if message.UplinkMessage.DecodedPayload.ActiveMode != nil {
		outputBool(
			"ActiveMode",
			"",
			message.UplinkMessage.DecodedPayload.ActiveMode,
		)
	}
	if message.UplinkMessage.DecodedPayload.Impulses != nil {
		outputInt(
			"Impulses",
			"",
			message.UplinkMessage.DecodedPayload.Impulses,
		)
	}
	if message.UplinkMessage.DecodedPayload.FenceVoltage != nil {
		outputFloat(
			"FenceVoltage",
			"V",
			message.UplinkMessage.DecodedPayload.FenceVoltage,
		)
	}
	if message.UplinkMessage.DecodedPayload.FenceVoltageStd != nil {
		outputFloat(
			"FenceVoltageStd",
			"",
			message.UplinkMessage.DecodedPayload.FenceVoltageStd,
		)
	}
	if message.UplinkMessage.DecodedPayload.FenceVoltageMin != nil {
		outputFloat(
			"FenceVoltageMin",
			"V",
			message.UplinkMessage.DecodedPayload.FenceVoltageMin,
		)
	}
	if message.UplinkMessage.DecodedPayload.FenceVoltageMax != nil {
		outputFloat(
			"FenceVoltageMax",
			"V",
			message.UplinkMessage.DecodedPayload.FenceVoltageMax,
		)
	}
	if message.UplinkMessage.DecodedPayload.BatteryVoltage != nil {
		outputFloat(
			"BatteryVoltage",
			"V",
			message.UplinkMessage.DecodedPayload.BatteryVoltage,
		)
	}
	if message.UplinkMessage.DecodedPayload.RemainingCapacity != nil {
		outputFloat(
			"RemainingCapacity",
			"mAh",
			message.UplinkMessage.DecodedPayload.RemainingCapacity,
		)
	}
	if message.UplinkMessage.DecodedPayload.Temperature != nil {
		outputFloat(
			"Temperature",
			"°C",
			message.UplinkMessage.DecodedPayload.Temperature,
		)
	}

	// any points sent?
	if count < 1 && c.LogDebug() {
		log.Printf(
			"ttn-fencyboy[%s]: could not extract any sensor data; payload='%s'",
			c.Name(), input.Payload(),
		)
		return
	}
}
