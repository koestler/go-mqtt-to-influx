package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influx/v2/converter/mock"
	"testing"
	"time"
)

func TestGoVeSensorV2(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := converter_mock.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().Name().Return("test-converter").AnyTimes()

	mockTMConfig := converter_mock.NewMockTopicMatcherConfig(mockCtrl)
	mockTMConfig.EXPECT().Topic().Return("piegn/tele/iot-device/%Device%/state").AnyTimes()
	mockTMConfig.EXPECT().Device().Return("+").AnyTimes()
	mockTMConfig.EXPECT().DeviceIsDynamic().Return(true).AnyTimes()

	stimuli := TestStimuliResponse{
		{
			Topic: "piegn/tele/iot-device/24v-bmv/state",
			Payload: `{
    "Time": "2022-12-29T11:19:19Z",
    "NextTelemetry": "2022-12-29T11:19:21Z",
    "Model": "SmartShunt 500A\/50mV",
    "NumericValues": {
        "AmountOfChargedEnergy": {
            "Cat": "Historic",
            "Desc": "Amount of charged energy",
            "Val": 16.71,
            "Unit": "kWh"
        },
        "CurrentHighRes": {
            "Cat": "Essential",
            "Desc": "Current",
            "Val": 0.18,
            "Unit": "A"
        },
        "MainVoltage": {
            "Cat": "Essential",
            "Desc": "Main voltage",
            "Val": 26.11,
            "Unit": "V"
        },
        "NumberOfCycles": {
            "Cat": "Historic",
            "Desc": "Number of cycles",
            "Val": 0
        },
        "Power": {
            "Cat": "Essential",
            "Desc": "Power",
            "Val": 5,
            "Unit": "W"
        },
        "ProductId": {
            "Cat": "Product",
            "Desc": "Product id",
            "Val": 4272130304
        },
        "SOC": {
            "Cat": "Essential",
            "Desc": "State of charge",
            "Val": 38.72,
            "Unit": "%"
        },
        "TTG": {
            "Cat": "Monitor",
            "Desc": "Time to go",
            "Val": 0,
            "Unit": "min"
        },
        "TimeSinceFullCharge": {
            "Cat": "Historic",
            "Desc": "Time since full charge",
            "Val": 473231,
            "Unit": "s"
        },
        "Uptime": {
            "Cat": "Product",
            "Desc": "Device uptime",
            "Val": 480744,
            "Unit": "s"
        }
    },
    "TextValues": {
        "ModelName": {
            "Cat": "Product",
            "Desc": "Model name",
            "Val": "BMV-SmartShunt 500A\/50mV"
        },
        "SerialNumber": {
            "Cat": "Product",
            "Desc": "Serial number",
            "Val": "HQ2117NTVX4"
        },
        "SynchronizationState": {
            "Cat": "Monitor",
            "Desc": "Synchronization state",
            "Val": "true"
        }
    }
}`,
			ExpectedLines: []string{
				"clock,device=24v-bmv timeValue=\"2022-12-29T11:19:19Z\"",
				"telemetry,category=Essential,description=Current,device=24v-bmv,field=CurrentHighRes,sensor=SmartShunt,unit=A floatValue=0.18",
				"telemetry,category=Essential,description=Main\\ voltage,device=24v-bmv,field=MainVoltage,sensor=SmartShunt,unit=V floatValue=26.11",
				"telemetry,category=Essential,description=Power,device=24v-bmv,field=Power,sensor=SmartShunt,unit=W floatValue=5",
				"telemetry,category=Essential,description=State\\ of\\ charge,device=24v-bmv,field=SOC,sensor=SmartShunt,unit=% floatValue=38.72",
				"telemetry,category=Historic,description=Amount\\ of\\ charged\\ energy,device=24v-bmv,field=AmountOfChargedEnergy,sensor=SmartShunt,unit=kWh floatValue=16.71",
				"telemetry,category=Historic,description=Number\\ of\\ cycles,device=24v-bmv,field=NumberOfCycles,sensor=SmartShunt,unit= floatValue=0",
				"telemetry,category=Historic,description=Time\\ since\\ full\\ charge,device=24v-bmv,field=TimeSinceFullCharge,sensor=SmartShunt,unit=s floatValue=473231",
				"telemetry,category=Monitor,description=Synchronization\\ state,device=24v-bmv,field=SynchronizationState,sensor=SmartShunt stringValue=\"true\"",
				"telemetry,category=Monitor,description=Time\\ to\\ go,device=24v-bmv,field=TTG,sensor=SmartShunt,unit=min floatValue=0",
				"telemetry,category=Product,description=Device\\ uptime,device=24v-bmv,field=Uptime,sensor=SmartShunt,unit=s floatValue=480744",
				"telemetry,category=Product,description=Model\\ name,device=24v-bmv,field=ModelName,sensor=SmartShunt stringValue=\"BMV-SmartShunt 500A/50mV\"",
				"telemetry,category=Product,description=Product\\ id,device=24v-bmv,field=ProductId,sensor=SmartShunt,unit= floatValue=4.272130304e+09",
				"telemetry,category=Product,description=Serial\\ number,device=24v-bmv,field=SerialNumber,sensor=SmartShunt stringValue=\"HQ2117NTVX4\"",
			},
			ExpectedTimeStamp: time.Now(),
		}, {
			// go-iotdevice <v2 wihout category/description fields
			Topic: "piegn/tele/iot-device/12v-solar/state",
			Payload: `{
  "NextTelemetry": "2022-08-19T14:52:24Z",
  "Model":"SmartSolar MPPT VE.Can 250/100",
  "NumericValues":{
    "Power":{"Cat": "Essential","Desc": "Power","Val":-18,"Unit":"W"}
  }
}`,
			ExpectedLines: []string{
				"telemetry,category=Essential,description=Power,device=12v-solar,field=Power,sensor=SmartSolar,unit=W floatValue=-18",
			},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic:             "piegn/tele/iot-device/24v-bmv/state",
			Payload:           "invalid",
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Now(),
		}, {
			Topic:             "invalid",
			Payload:           "{}",
			ExpectedLines:     []string{},
			ExpectedTimeStamp: time.Now(),
		},
		{
			Topic: "piegn/tele/iot-device/12v-solar/state",
			Payload: `{
    "Time": "2023-10-31T11:04:58+01:00",
    "NextTelemetry": "2023-10-31T11:05:08+01:00",
    "Model": "Teracom",
    "NumericValues": {
        "AI1": {
            "Cat": "Analog Inputs",
            "Desc": "Analog Input 1",
            "Val": 0.02,
            "Unit": "V"
        },
        "AI1Hys": {
            "Cat": "Settings",
            "Desc": "Analog Input 1 Hysteresis",
            "Val": 1,
            "Unit": "V"
        },
        "AI1Max": {
            "Cat": "Settings",
            "Desc": "Analog Input 1 Max",
            "Val": 60,
            "Unit": "V"
        },
        "AI1Min": {
            "Cat": "Settings",
            "Desc": "Analog Input 1 Min",
            "Val": 0,
            "Unit": "V"
        },
        "AI2": {
            "Cat": "Analog Inputs",
            "Desc": "Analog Input 2",
            "Val": 0.02,
            "Unit": "V"
        }
    },
    "TextValues": {
        "Date": {
            "Cat": "General",
            "Desc": "Date",
            "Val": "31.10.2023"
        },
        "DeviceName": {
            "Cat": "Device Info",
            "Desc": "Device Name",
            "Val": "TCW241"
        },
        "FWVer": {
            "Cat": "Device Info",
            "Desc": "Firmware Vesion",
            "Val": "TCW241-v1.248"
        },
        "HostName": {
            "Cat": "Device Info",
            "Desc": "Host Name",
            "Val": "TCW241"
        },
        "Id": {
            "Cat": "Device Info",
            "Desc": "Id",
            "Val": "5C:32:C5:00:C8:72"
        },
        "Time": {
            "Cat": "General",
            "Desc": "Time",
            "Val": "13:07:30"
        }
    },
    "EnumValues": {
        "AI1Alarm": {
            "Cat": "Alarms",
            "Desc": "Analog Input 1",
            "Idx": 1,
            "Val": "ALARMED"
        },
        "DI1": {
            "Cat": "Digital Inputs",
            "Desc": "Digital Input 1",
            "Idx": 1,
            "Val": "CLOSED"
        }
    }
}`,
			ExpectedLines: []string{
				"clock,device=12v-solar timeValue=\"2023-10-31T11:04:58+01:00\"",
				"telemetry,category=Alarms,description=Analog\\ Input\\ 1,device=12v-solar,field=AI1Alarm,sensor=Teracom intValue=1i,stringValue=\"ALARMED\"",
				"telemetry,category=Analog\\ Inputs,description=Analog\\ Input\\ 1,device=12v-solar,field=AI1,sensor=Teracom,unit=V floatValue=0.02",
				"telemetry,category=Analog\\ Inputs,description=Analog\\ Input\\ 2,device=12v-solar,field=AI2,sensor=Teracom,unit=V floatValue=0.02",
				"telemetry,category=Device\\ Info,description=Device\\ Name,device=12v-solar,field=DeviceName,sensor=Teracom stringValue=\"TCW241\"",
				"telemetry,category=Device\\ Info,description=Firmware\\ Vesion,device=12v-solar,field=FWVer,sensor=Teracom stringValue=\"TCW241-v1.248\"",
				"telemetry,category=Device\\ Info,description=Host\\ Name,device=12v-solar,field=HostName,sensor=Teracom stringValue=\"TCW241\"",
				"telemetry,category=Device\\ Info,description=Id,device=12v-solar,field=Id,sensor=Teracom stringValue=\"5C:32:C5:00:C8:72\"",
				"telemetry,category=Digital\\ Inputs,description=Digital\\ Input\\ 1,device=12v-solar,field=DI1,sensor=Teracom intValue=1i,stringValue=\"CLOSED\"",
				"telemetry,category=General,description=Date,device=12v-solar,field=Date,sensor=Teracom stringValue=\"31.10.2023\"",
				"telemetry,category=General,description=Time,device=12v-solar,field=Time,sensor=Teracom stringValue=\"13:07:30\"",
				"telemetry,category=Settings,description=Analog\\ Input\\ 1\\ Hysteresis,device=12v-solar,field=AI1Hys,sensor=Teracom,unit=V floatValue=1",
				"telemetry,category=Settings,description=Analog\\ Input\\ 1\\ Max,device=12v-solar,field=AI1Max,sensor=Teracom,unit=V floatValue=60",
				"telemetry,category=Settings,description=Analog\\ Input\\ 1\\ Min,device=12v-solar,field=AI1Min,sensor=Teracom,unit=V floatValue=0",
			},
			ExpectedTimeStamp: time.Now(),
		},
	}

	if h, err := GetHandler("go-iotdevice"); err != nil {
		t.Errorf("did not expect an error while getting handler: %s", err)
	} else {
		testStimuliResponse(t, mockCtrl, mockConfig, mockTMConfig, h, stimuli)
	}
}
