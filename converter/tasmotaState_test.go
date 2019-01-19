package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influx/converter/mock"
	"testing"
	"time"
)

func TestTasmotaState(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := converter_mock.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Name().Return("test-converter").AnyTimes()

	stimuli := TestStimuliResponse{
		{
			Topic: "piegn/tele/elektronik/control0/STATE",
			Payload: `{
  "Time":"2019-01-10T22:45:22",
  "Uptime":"9T09:29:01",
  "Vcc":3.108,
  "POWER1":"OFF",
  "POWER2":"ON",
  "POWER3":"OFF",
  "POWER4":"OFF",
  "Wifi":{"AP":1,"SSId":"piegn-iot","BSSId":"04:F0:21:2F:B7:CC","Channel":1,"RSSI":100}
}`,
			ExpectedLines: []string{
				"timeValue,device=elektronik/control0 value=\"2019-01-10 22:45:22 +0000 UTC\"",
				"floatValue,device=elektronik/control0,field=UpTime,unit=s value=811741",
				"floatValue,device=elektronik/control0,field=Vcc,unit=V value=3.108",
				"boolValue,device=elektronik/control0,field=Power1 value=false",
				"boolValue,device=elektronik/control0,field=Power2 value=true",
				"boolValue,device=elektronik/control0,field=Power3 value=false",
				"boolValue,device=elektronik/control0,field=Power4 value=false",
				"wifi,BSSId=04:F0:21:2F:B7:CC,SSId=piegn-iot,device=elektronik/control0 AP=1i,Channel=1i,RSSI=100i",
			},
			ExpectedTimeStamp: time.Date(2019, time.January, 10, 22, 45, 22, 0, time.UTC),
		}, {
			Topic: "piegn/tele/mezzo/bridge0/STATE",
			Payload: `{
  "Time":"2019-01-10T22:45:24",
  "Uptime":"0T01:35:14","Vcc":3.248,
  "Wifi":{"AP":2,"SSId":"piegn-iot","BSSId":"04:F0:21:33:40:99","Channel":1,"RSSI":90}
}`,
			ExpectedLines: []string{
				"timeValue,device=mezzo/bridge0 value=\"2019-01-10 22:45:24 +0000 UTC\"",
				"floatValue,device=mezzo/bridge0,field=UpTime,unit=s value=5714",
				"floatValue,device=mezzo/bridge0,field=Vcc,unit=V value=3.248",
				"wifi,BSSId=04:F0:21:33:40:99,SSId=piegn-iot,device=mezzo/bridge0 AP=2i,Channel=1i,RSSI=90i",
			},
			ExpectedTimeStamp: time.Date(2019, time.January, 10, 22, 45, 24, 0, time.UTC),
		}, {
			Topic: "piegn/tele/mezzo/zimmer-gross/STATE",
			Payload: `{
  "Time":"2019-01-10T22:45:41",
  "Uptime":"0T01:35:45",
  "Vcc":3.177,
  "POWER":"invalid",
  "Wifi":{"AP":2,"SSId":"piegn-iot","BSSId":"04:F0:21:33:40:99","Channel":1,"RSSI":100}
}`,
			ExpectedLines: []string{
				"timeValue,device=mezzo/zimmer-gross value=\"2019-01-10 22:45:41 +0000 UTC\"",
				"floatValue,device=mezzo/zimmer-gross,field=UpTime,unit=s value=5745",
				"floatValue,device=mezzo/zimmer-gross,field=Vcc,unit=V value=3.177",
				"wifi,BSSId=04:F0:21:33:40:99,SSId=piegn-iot,device=mezzo/zimmer-gross AP=2i,Channel=1i,RSSI=100i",
			},
			ExpectedTimeStamp: time.Date(2019, time.January, 10, 22, 45, 41, 0, time.UTC),
		}, {
			Topic: "piegn/tele/mezzo/zimmer-gross/STATE",
			Payload: `{
  "Time":"invalid2019-01-10T22:45:41",
  "Vcc":3.177,
  "POWER":"invalid",
  "Wifi":{"AP":2,"SSId":"piegn-iot","BSSId":"04:F0:21:33:40:99","Channel":1,"RSSI":100}
}`,
			ExpectedLines: []string{
				"floatValue,device=mezzo/zimmer-gross,field=Vcc,unit=V value=3.177",
				"wifi,BSSId=04:F0:21:33:40:99,SSId=piegn-iot,device=mezzo/zimmer-gross AP=2i,Channel=1i,RSSI=100i",
			},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic:             "piegn/tele/mezzo/bridge0/STATE",
			Payload:           "invalid",
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic:             "invalid",
			Payload:           ``,
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Now(),
		},
	}

	if h, err := GetHandler("tasmota-state"); err != nil {
		t.Errorf("did not expect an error while getting handler: %s", err)
	} else {
		testStimuliResponse(t, mockCtrl, mockConfig, h, stimuli)
	}
}
