package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influx/converter/mock"
	"testing"
	"time"
)

func TestGoVeSensor(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := converter_mock.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Name().Return("test-converter").AnyTimes()
	mockConfig.EXPECT().TargetMeasurement().Return("telemetry").MinTimes(1)

	stimuli := TestStimuliResponse{
		{
			Topic: "piegn/tele/iot-device/24v-bmv/state",
			Payload: `{
  "Time": "2022-08-19T14:52:19Z",
  "NextTelemetry": "2022-08-19T14:52:24Z",
  "Model": "BMV-702",
  "SecondsSinceLastUpdate": 0.576219653,
  "NumericValues": {
    "AmountOfChargedEnergy": {
      "Value": 1883.52,
      "Unit": "kWh"
    },
    "CurrentHighRes": {
      "Value": -0.58,
      "Unit": "A"
    },
    "NumberOfCycles": {
      "Value": 241,
      "Unit": ""
    },
    "ProductId": {
      "Value": 4261544960,
      "Unit": ""
    },
    "SOC": {
      "Value": 58.16,
      "Unit": "%"
    },
    "TTG": {
      "Value": 5742,
      "Unit": "min"
    },
    "Uptime": {
      "Value": 17182790,
      "Unit": "s"
    }
  },
  "TextValues": {
    "ModelName": {
      "Value": "BMV-702"
    },
    "SerialNumber": {
      "Value": "HQ15149CFQI,HQ1515RP6L7,"
    },
    "SynchronizationState": {
      "Value": "true"
    }
  }
}
`,
			ExpectedLines: []string{
				"telemetry,device=24v-bmv,field=AmountOfChargedEnergy,sensor=BMV-702,unit=kWh floatValue=1883.52",
				"telemetry,device=24v-bmv,field=CurrentHighRes,sensor=BMV-702,unit=A floatValue=-0.58",
				"telemetry,device=24v-bmv,field=ModelName,sensor=BMV-702 stringValue=\"BMV-702\"",
				"telemetry,device=24v-bmv,field=NumberOfCycles,sensor=BMV-702,unit= floatValue=241",
				"telemetry,device=24v-bmv,field=ProductId,sensor=BMV-702,unit= floatValue=4.26154496e+09",
				"telemetry,device=24v-bmv,field=SOC,sensor=BMV-702,unit=% floatValue=58.16",
				"telemetry,device=24v-bmv,field=SerialNumber,sensor=BMV-702 stringValue=\"HQ15149CFQI,HQ1515RP6L7,\"",
				"telemetry,device=24v-bmv,field=SynchronizationState,sensor=BMV-702 stringValue=\"true\"",
				"telemetry,device=24v-bmv,field=TTG,sensor=BMV-702,unit=min floatValue=5742",
				"telemetry,device=24v-bmv,field=Uptime,sensor=BMV-702,unit=s floatValue=1.718279e+07",
			},
			ExpectedTimeStamp: time.Date(2022, time.August, 19, 14, 52, 19, 0, time.UTC),
		}, {
			Topic: "piegn/tele/ve/24v-bmv",
			Payload: `{
  "NextTelemetry": "2022-08-19T14:52:24Z",
  "Model":"SmartSolar MPPT VE.Can 250/100",
  "NumericValues":{
    "Power":{"Value":-18,"Unit":"W"}
  }
}`,
			ExpectedLines: []string{
				"telemetry,device=24v-bmv,field=Power,sensor=SmartSolar,unit=W floatValue=-18",
				"telemetry,device=24v-bmv,field=Model,sensor=SmartSolar stringValue=\"SmartSolar MPPT VE.Can 250/100\"",
			},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic:             "piegn/tele/ve/24v-bmv/state",
			Payload:           "invalid",
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic:             "invalid",
			Payload:           "{}",
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Now(),
		},
	}

	if h, err := GetHandler("go-iotdevice"); err != nil {
		t.Errorf("did not expect an error while getting handler: %s", err)
	} else {
		testStimuliResponse(t, mockCtrl, mockConfig, h, stimuli)
	}
}
