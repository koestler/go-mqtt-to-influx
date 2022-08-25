package converter

import "time"

type telemetryOutputMessage struct {
	timeStamp   time.Time
	device      string
	field       string
	unit        *string
	sensor      string
	stringValue *string
	floatValue  *float64
	boolValue   *bool
	intValue    *int64
}

func (m telemetryOutputMessage) Measurement() string {
	return "telemetry"
}

func (m telemetryOutputMessage) Tags() map[string]string {
	ret := map[string]string{
		"sensor": m.sensor,
		"device": m.device,
		"field":  m.field,
	}

	if m.unit != nil {
		ret["unit"] = *m.unit
	}

	return ret
}

func (m telemetryOutputMessage) Fields() (ret map[string]interface{}) {
	ret = make(map[string]interface{}, 2)

	if m.floatValue != nil {
		ret["floatValue"] = *m.floatValue
	}

	if m.stringValue != nil {
		ret["stringValue"] = *m.stringValue
	}

	if m.boolValue != nil {
		ret["boolValue"] = *m.boolValue
	}

	if m.intValue != nil {
		ret["intValue"] = *m.intValue
	}

	return
}

func (m telemetryOutputMessage) Time() time.Time {
	return m.timeStamp
}
