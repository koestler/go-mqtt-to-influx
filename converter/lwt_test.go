package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influxdb/converter/mock"
	"testing"
)

func Test(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := converter_mock.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Name().Return("test-converter").AnyTimes()
	mockConfig.EXPECT().TargetMeasurement().Return("boolValue").MinTimes(1)

	stimuli := TestStimuliResponse{
		{
			Topic:         "piegn/tele/software/srv1-go-ve-sensor/LWT",
			Payload:       "Online",
			ExpectedLines: []string{"boolValue,device=software/srv1-go-ve-sensor,field=Available value=true"},
		}, {
			Topic:         "piegn/tele/software/srv1-go-ve-sensor/LWT",
			Payload:       "Offline",
			ExpectedLines: []string{"boolValue,device=software/srv1-go-ve-sensor,field=Available value=false"},
		}, {
			Topic:         "piegn/tele/software/srv1-go-ve-sensor/LWT",
			Payload:       "invalid",
			ExpectedLines: []string{},
		}, {
			Topic:         "piegn/tele/software/srv1-go-ve-sensor/LWT-invalid-topic",
			Payload:       "Online",
			ExpectedLines: []string{},
		},
	}

	testStimuliResponse(t, mockCtrl, mockConfig, lwtHandler, stimuli)
}
