package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influxdb/converter/mock"
	"testing"
	"time"
)

func TestTasmotaSensor(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := converter_mock.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Name().Return("test-converter").AnyTimes()
	mockConfig.EXPECT().TargetMeasurement().Return("floatValue").MinTimes(1)

	stimuli := TestStimuliResponse{
		{
			Topic:   "piegn/tele/elektronik/control0/SENSOR",
			Payload: `{"Time":"2019-01-10T22:15:52","SI7021":{"Temperature":5.4,"Humidity":27.7},"TempUnit":"C"}`,
			ExpectedLines: []string{
				"floatValue,device=elektronik/control0,field=Temperature,sensor=SI7021,unit=C value=5.4",
				"floatValue,device=elektronik/control0,field=Humidity,sensor=SI7021,unit=% value=27.7"},
			ExpectedTimeStamp: time.Date(2019, time.January, 10, 22, 15, 52, 0, time.UTC),
		}, {
			Topic:   "piegn/tele/elektronik/mezzo-haupt/SENSOR",
			Payload: `{"Time":"2019-01-10T22:16:03","AM2301":{"Temperature":5.2,"Humidity":30.8},"TempUnit":"C"}`,
			ExpectedLines: []string{
				"floatValue,device=elektronik/mezzo-haupt,field=Temperature,sensor=AM2301,unit=C value=5.2",
				"floatValue,device=elektronik/mezzo-haupt,field=Humidity,sensor=AM2301,unit=% value=30.8",
			},
			ExpectedTimeStamp: time.Date(2019, time.January, 10, 22, 16, 03, 0, time.UTC),
		}, {
			Topic:   "piegn/tele/mezzo/kuehlschrank/SENSOR",
			Payload: `{"Time":"2019-01-10T22:16:04","DS18B20":{"Temperature":3.0},"TempUnit":"C"}`,
			ExpectedLines: []string{
				"floatValue,device=mezzo/kuehlschrank,field=Temperature,sensor=DS18B20,unit=C value=3",
			},
			ExpectedTimeStamp: time.Date(2019, time.January, 10, 22, 16, 04, 0, time.UTC),
		}, {
			Topic:             "piegn/tele/mezzo/no-sensor/SENSOR",
			Payload:           `{"Time":"2019-01-10T22:16:04","unkown":{"Temperature":3.0},"TempUnit":"C"}`,
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Date(2019, time.January, 10, 22, 16, 04, 0, time.UTC),
		}, {
			Topic:             "piegn/tele/mezzo/invalidTime/SENSOR",
			Payload:           `{"Time":"2019-01-10T22:16:04qq","unkown":{"Temperature":3.0},"TempUnit":"C"}`,
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Date(2019, time.January, 10, 22, 16, 04, 0, time.UTC),
		}, {
			Topic:             "piegn/tele/mezzo/kuehlschrank/SENSOR",
			Payload:           "invalid",
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic:             "invalid",
			Payload:           `{"Time":"2019-01-10T22:16:04","DS18B20":{"Temperature":3.0},"TempUnit":"C"}`,
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Now(),
		},
	}

	testStimuliResponse(t, mockCtrl, mockConfig, tasmotaSensorHandler, stimuli)
}
