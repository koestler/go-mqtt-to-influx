package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influx/converter/mock"
	"testing"
	"time"
)

func TestLwt(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := converter_mock.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().Name().Return("test-converter").AnyTimes()

	now := time.Now()

	stimuli := TestStimuliResponse{
		{
			Topic:             "piegn/tele/software/srv1-go-iotdevice/LWT",
			Payload:           "Online",
			ExpectedLines:     []string{"availability,device=software/srv1-go-iotdevice boolValue=true"},
			ExpectedTimeStamp: now,
		}, {
			Topic:             "piegn/tele/mezzo/stube-licht1/LWT",
			Payload:           "Online",
			ExpectedLines:     []string{"availability,device=mezzo/stube-licht1 boolValue=true"},
			ExpectedTimeStamp: now,
		}, {
			Topic:             "piegn/tele/software/srv1-go-iotdevice/LWT",
			Payload:           "Offline",
			ExpectedLines:     []string{"availability,device=software/srv1-go-iotdevice boolValue=false"},
			ExpectedTimeStamp: now,
		}, {
			Topic:             "piegn/tele/software/srv1-go-iotdevice/LWT",
			Payload:           "invalid",
			ExpectedLines:     []string{},
			ExpectedTimeStamp: now,
		}, {
			Topic:             "piegn/tele/software/srv1-go-iotdevice/LWT-invalid-topic",
			Payload:           "Online",
			ExpectedLines:     []string{},
			ExpectedTimeStamp: now,
		},
	}

	if h, err := GetHandler("availability"); err != nil {
		t.Errorf("did not expect an error while getting handler: %s", err)
	} else {
		testStimuliResponse(t, mockCtrl, mockConfig, h, stimuli)
	}
}
