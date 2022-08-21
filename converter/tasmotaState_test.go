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
				"clock,device=elektronik/control0 timeValue=\"2019-01-10T22:45:22Z\"",
				"telemetry,device=elektronik/control0,field=UpTime,sensor=tasmota,unit=s floatValue=811741",
				"telemetry,device=elektronik/control0,field=Vcc,sensor=tasmota,unit=V floatValue=3.108",
				"telemetry,device=elektronik/control0,field=Power1,sensor=tasmota boolValue=false",
				"telemetry,device=elektronik/control0,field=Power2,sensor=tasmota boolValue=true",
				"telemetry,device=elektronik/control0,field=Power3,sensor=tasmota boolValue=false",
				"telemetry,device=elektronik/control0,field=Power4,sensor=tasmota boolValue=false",
				"wifi,BSSId=04:F0:21:2F:B7:CC,SSId=piegn-iot,device=elektronik/control0 AP=1i,Channel=1i,RSSI=100i",
			},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic: "piegn/tele/mezzo/bridge0/STATE",
			Payload: `{
  "Time":"2019-01-10T22:45:24",
  "Uptime":"0T01:35:14","Vcc":3.248,
  "Wifi":{"AP":2,"SSId":"piegn-iot","BSSId":"04:F0:21:33:40:99","Channel":1,"RSSI":90}
}`,
			ExpectedLines: []string{
				"clock,device=mezzo/bridge0 timeValue=\"2019-01-10T22:45:24Z\"",
				"telemetry,device=mezzo/bridge0,field=UpTime,sensor=tasmota,unit=s floatValue=5714",
				"telemetry,device=mezzo/bridge0,field=Vcc,sensor=tasmota,unit=V floatValue=3.248",
				"wifi,BSSId=04:F0:21:33:40:99,SSId=piegn-iot,device=mezzo/bridge0 AP=2i,Channel=1i,RSSI=90i",
			},
			ExpectedTimeStamp: time.Now(),
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
				"clock,device=mezzo/zimmer-gross timeValue=\"2019-01-10T22:45:41Z\"",
				"telemetry,device=mezzo/zimmer-gross,field=UpTime,sensor=tasmota,unit=s floatValue=5745",
				"telemetry,device=mezzo/zimmer-gross,field=Vcc,sensor=tasmota,unit=V floatValue=3.177",
				"wifi,BSSId=04:F0:21:33:40:99,SSId=piegn-iot,device=mezzo/zimmer-gross AP=2i,Channel=1i,RSSI=100i",
			},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic: "piegn/tele/mezzo/zimmer-gross/STATE",
			Payload: `{
  "Time":"invalid2019-01-10T22:45:41",
  "Vcc":3.177,
  "POWER":"invalid",
  "Wifi":{"AP":2,"SSId":"piegn-iot","BSSId":"04:F0:21:33:40:99","Channel":1,"RSSI":100}
}`,
			ExpectedLines: []string{
				"telemetry,device=mezzo/zimmer-gross,field=Vcc,sensor=tasmota,unit=V floatValue=3.177",
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
