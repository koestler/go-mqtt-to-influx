package converter

import (
	"log"
	"time"
)

type ttnDraginoMessage struct {
	ReceivedAt    time.Time `json:"received_at"`
	UplinkMessage struct {
		DecodedPayload struct {
			AdcCh0         *float64 `json:"ADC_CH0V"`
			BatV           *float64 `json:"BatV"`
			BatMV          *float64 `json:"Bat_mV"`
			DigitalIStatus *string  `json:"Digital_IStatus"`
			DoorStatus     *string  `json:"Door_status"`
			EXTITrigger    *string  `json:"EXTI_Trigger"`
			HumSHT         *float64 `json:"Hum_SHT"`
			WorkMode       *string  `json:"Work_mode"`
			AlarmStatus    *string  `json:"ALARM_status"`
			TempBlack      *float64 `json:"Temp_Black"`
			TempRed        *float64 `json:"Temp_Red"`
			TempWhite      *float64 `json:"Temp_White"`
			TempC1         *float64 `json:"TempC1"`
			TempC2         *float64 `json:"TempC2"`
			TempC3         *float64 `json:"TempC3"`
			TempCDs        *float64 `json:"TempC_DS"`
			TempCSht       *float64 `json:"TempC_SHT"`
			Ext            *int64   `json:"Ext"`
			Systimestamp   *int64   `json:"Systimestamp"`
		} `json:"decoded_payload"`
	} `json:"uplink_message"`
}

func ttnDraginoHandler(c Config, device, model string, input Input, outputFunc OutputFunc) {
	// parse payload
	var message ttnDraginoMessage
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("ttn-dragino[%s]: cannot json decode: %s", c.Name(), err)
		return
	}

	// save clock
	if message.UplinkMessage.DecodedPayload.Systimestamp != nil {
		outputFunc(stateClockOutputMessage{
			timeStamp: message.ReceivedAt,
			device:    device,
			value:     time.Unix(*message.UplinkMessage.DecodedPayload.Systimestamp, 0).UTC(),
		})
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
	outputString := func(field string, unit string, stringValue *string) {
		output(field, unit, stringValue, nil, nil, nil)
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

	if message.UplinkMessage.DecodedPayload.AdcCh0 != nil {
		outputFloat(
			"AdcCh0",
			"V",
			message.UplinkMessage.DecodedPayload.AdcCh0,
		)
	}
	if message.UplinkMessage.DecodedPayload.BatV != nil {
		outputFloat(
			"BatV",
			"V",
			message.UplinkMessage.DecodedPayload.BatV,
		)
	}
	if message.UplinkMessage.DecodedPayload.BatMV != nil {
		value := *message.UplinkMessage.DecodedPayload.BatMV / 1000
		outputFloat(
			"BatV",
			"V",
			&value,
		)
	}
	if message.UplinkMessage.DecodedPayload.DigitalIStatus != nil {
		outputString(
			"DigitalIStatus",
			"",
			message.UplinkMessage.DecodedPayload.DigitalIStatus,
		)
	}
	if message.UplinkMessage.DecodedPayload.DoorStatus != nil {
		outputString(
			"DoorStatus",
			"",
			message.UplinkMessage.DecodedPayload.DoorStatus,
		)
	}
	if message.UplinkMessage.DecodedPayload.EXTITrigger != nil {
		value := false
		if *message.UplinkMessage.DecodedPayload.EXTITrigger == "TRUE" {
			value = true
		}
		outputBool(
			"EXTITrigger",
			"",
			&value,
		)
	}
	if message.UplinkMessage.DecodedPayload.HumSHT != nil {
		outputFloat(
			"HumSHT",
			"%",
			message.UplinkMessage.DecodedPayload.HumSHT,
		)
	}
	if message.UplinkMessage.DecodedPayload.WorkMode != nil {
		outputString(
			"WorkMode",
			"",
			message.UplinkMessage.DecodedPayload.WorkMode,
		)
	}
	if message.UplinkMessage.DecodedPayload.AlarmStatus != nil {
		alarmStatus := false
		if *message.UplinkMessage.DecodedPayload.AlarmStatus == "TRUE" {
			alarmStatus = true
		}
		outputBool(
			"AlarmStatus",
			"",
			&alarmStatus,
		)
	}
	if message.UplinkMessage.DecodedPayload.TempBlack != nil {
		outputFloat(
			"TempBlack",
			"°C",
			message.UplinkMessage.DecodedPayload.TempBlack,
		)
	}
	if message.UplinkMessage.DecodedPayload.TempRed != nil {
		outputFloat(
			"TempRed",
			"°C",
			message.UplinkMessage.DecodedPayload.TempRed,
		)
	}
	if message.UplinkMessage.DecodedPayload.TempWhite != nil {
		outputFloat(
			"TempWhite",
			"°C",
			message.UplinkMessage.DecodedPayload.TempWhite,
		)
	}
	if message.UplinkMessage.DecodedPayload.TempC1 != nil {
		outputFloat(
			"TempC1",
			"°C",
			message.UplinkMessage.DecodedPayload.TempC1,
		)
	}
	if message.UplinkMessage.DecodedPayload.TempC2 != nil {
		outputFloat(
			"TempC2",
			"°C",
			message.UplinkMessage.DecodedPayload.TempC2,
		)
	}
	if message.UplinkMessage.DecodedPayload.TempC3 != nil {
		outputFloat(
			"TempC3",
			"°C",
			message.UplinkMessage.DecodedPayload.TempC3,
		)
	}
	if message.UplinkMessage.DecodedPayload.TempCDs != nil {
		outputFloat(
			"TempCDs",
			"V",
			message.UplinkMessage.DecodedPayload.TempCDs,
		)
	}
	if message.UplinkMessage.DecodedPayload.TempCSht != nil {
		outputFloat(
			"TempCSht",
			"°C",
			message.UplinkMessage.DecodedPayload.TempCSht,
		)
	}
	if message.UplinkMessage.DecodedPayload.Ext != nil {
		outputInt(
			"Ext",
			"",
			message.UplinkMessage.DecodedPayload.Ext,
		)
	}
	if message.UplinkMessage.DecodedPayload.Systimestamp != nil {
		outputInt(
			"Systimestamp",
			"ms",
			message.UplinkMessage.DecodedPayload.Systimestamp,
		)
	}

	// any points sent?
	if count < 1 && c.LogDebug() {
		log.Printf(
			"ttn-dragino[%s]: could not extract any sensor data; payload='%s'",
			c.Name(), input.Payload(),
		)
		return
	}
}
