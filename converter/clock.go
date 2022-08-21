package converter

import "time"

type stateClockOutputMessage struct {
	timeStamp time.Time
	device    string
	value     time.Time
}

func (m stateClockOutputMessage) Measurement() string {
	return "clock"
}

func (m stateClockOutputMessage) Tags() map[string]string {
	return map[string]string{
		"device": m.device,
	}
}

func (m stateClockOutputMessage) Fields() map[string]interface{} {
	return map[string]interface{}{
		"timeValue": m.value,
	}
}

func (m stateClockOutputMessage) Time() time.Time {
	return m.timeStamp
}
