package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influxdb/converter/mock"
	"testing"
	"time"
)

func TestGoVeSensor(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := converter_mock.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Name().Return("test-converter").AnyTimes()
	mockConfig.EXPECT().TargetMeasurement().Return("floatValue").MinTimes(1)

	stimuli := TestStimuliResponse{
		{
			Topic: "piegn/tele/ve/24v-bmv",
			Payload: `{
  "Time":"2019-01-06T23:40:03",
  "NextTele":"2019-01-06T23:40:13",
  "TimeZone":"UTC",
  "Model":"bmv700",
  "Values":{
    "AmountOfChargedEnergy":{"Value":756.6,"Unit":"kWh"},
    "AmountOfDischargedEnergy":{"Value":363.1,"Unit":"kWh"},
    "Consumed":{"Value":-7.2,"Unit":"Ah"},
    "Current":{"Value":-0.7,"Unit":"A"},
    "StateOfCharge":{"Value":99,"Unit":"%"},
    "Power":{"Value":-18,"Unit":"W"},
    "TimeToGo":{"Value":14400,"Unit":"min"}
  }
}`,
			ExpectedLines: []string{
				"floatValue,device=24v-bmv,field=AmountOfChargedEnergy,sensor=bmv700,unit=kWh value=756.6",
				"floatValue,device=24v-bmv,field=AmountOfDischargedEnergy,sensor=bmv700,unit=kWh value=363.1",
				"floatValue,device=24v-bmv,field=Consumed,sensor=bmv700,unit=Ah value=-7.2",
				"floatValue,device=24v-bmv,field=Current,sensor=bmv700,unit=A value=-0.7",
				"floatValue,device=24v-bmv,field=StateOfCharge,sensor=bmv700,unit=% value=99",
				"floatValue,device=24v-bmv,field=Power,sensor=bmv700,unit=W value=-18",
				"floatValue,device=24v-bmv,field=TimeToGo,sensor=bmv700,unit=min value=14400",
			},
			ExpectedTimeStamp: time.Date(2019, time.January, 6, 23, 40, 3, 0, time.UTC),
		}, {
			Topic: "piegn/tele/ve/24v-bmv",
			Payload: `{
  "NextTele":"2019-01-06T23:40:13",
  "TimeZone":"UTC",
  "Model":"bmv700",
  "Values":{
    "Power":{"Value":-18,"Unit":"W"}
  }
}`,
			ExpectedLines:     []string{"floatValue,device=24v-bmv,field=Power,sensor=bmv700,unit=W value=-18"},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic:             "piegn/tele/ve/24v-bmv",
			Payload:           "invalid",
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic:             "invalid",
			Payload:           "{}",
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Now(),
		}, {
			// invalid timezone; is logged but should continue output something anyway
			Topic: "piegn/tele/ve/24v-bmv",
			Payload: `{
  "Time":"2019-01-06T23:40:03",
  "NextTele":"2019-01-06T23:40:13",
  "TimeZone":"GMT+1",
  "Model":"bmv700",
  "Values":{
    "Current":{"Value":-0.7,"Unit":"A"}
  }
}`,
			ExpectedLines:     []string{"floatValue,device=24v-bmv,field=Current,sensor=bmv700,unit=A value=-0.7"},
			ExpectedTimeStamp: time.Date(2019, time.January, 6, 23, 40, 3, 0, time.UTC),
		},
	}

	if h, err := GetHandler("go-ve-sensor"); err != nil {
		t.Errorf("did not expect an error while getting handler: %s", err)
	} else {
		testStimuliResponse(t, mockCtrl, mockConfig, h, stimuli)
	}
}
