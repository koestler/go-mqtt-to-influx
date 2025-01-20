package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influx/v2/converter/mock"
	"testing"
	"time"
)

func TestTtnFencyboy(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := converter_mock.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().Name().Return("test-converter").AnyTimes()

	mockTMConfig := converter_mock.NewMockTopicMatcherConfig(mockCtrl)
	mockTMConfig.EXPECT().Topic().Return("v3/piegn@ttn/devices/%Device%/up").AnyTimes()
	mockTMConfig.EXPECT().Device().Return("+").AnyTimes()
	mockTMConfig.EXPECT().DeviceIsDynamic().Return(true).AnyTimes()

	stimuli := TestStimuliResponse{
		{
			Topic: "v3/piegn@ttn/devices/fencyboy-0/up",
			Payload: `{
    "@type": "type.googleapis.com/ttn.lorawan.v3.ApplicationUp",
    "end_device_ids": {
      "device_id": "fencyboy-0",
      "application_ids": {
        "application_id": "piegn"
      },
      "dev_eui": "2B94F841240A45D3",
      "join_eui": "6BF6188C16A33EDC",
      "dev_addr": "260B373F"
    },
    "correlation_ids": [
      "gs:uplink:01JHWAD9JRSAV1F4SPZQ143FK3"
    ],
    "received_at": "2025-01-18T09:02:55.274197028Z",
    "uplink_message": {
      "session_key_id": "AZR08PzYRzOPxmZbZzVX2A==",
      "f_port": 1,
      "f_cnt": 104,
      "frm_payload": "HQ0IAdIoBgFCJ78ogA0hRBY4xQIU",
      "decoded_payload": {
        "ACTIVE_MODE": true,
        "BATTERYVOLTAGE": 3.361,
        "FENCEVOLTAGE": 10246,
        "FENCEVOLTAGEMAX": 10368,
        "FENCEVOLTAGEMIN": 10175,
        "FENCE_VOLTAGE_STD": 32.2,
        "IMPULSES": 466,
        "REMAINING_CAPACITY": 600.8870239257812,
        "TEMPERATURE": 5.32
      },
      "rx_metadata": [
        {
          "gateway_ids": {
            "gateway_id": "piegn-srv3",
            "eui": "E45F01FFFEDECBE3"
          },
          "time": "2025-01-18T09:02:54.979598045Z",
          "timestamp": 987106386,
          "rssi": -89,
          "channel_rssi": -89,
          "snr": 4.25,
          "uplink_token": "ChgKFgoKcGllZ24tc3J2MxII5F8B//7ey+MQ0pjY1gMaCwi/3K28BhDpwNgeINDAgKHdhAg=",
          "received_at": "2025-01-18T09:02:55.022361821Z"
        }
      ],
      "settings": {
        "data_rate": {
          "lora": {
            "bandwidth": 125000,
            "spreading_factor": 7,
            "coding_rate": "4/5"
          }
        },
        "frequency": "867300000",
        "timestamp": 987106386,
        "time": "2025-01-18T09:02:54.979598045Z"
      },
      "received_at": "2025-01-18T09:02:55.066145227Z",
      "consumed_airtime": "0.077056s",
      "network_ids": {
        "net_id": "000013",
        "ns_id": "EC656E0000000181",
        "tenant_id": "ttn",
        "cluster_id": "eu1",
        "cluster_address": "eu1.cloud.thethings.network"
      }
    }
  }`,
			ExpectedLines: []string{
				"lora,devEui=2B94F841240A45D3,device=fencyboy-0,gatewayEui=E45F01FFFEDECBE3,gatewayId=piegn-srv3 channelRssi=-89i,consumedAirtimeUs=77056i,gatewayIdx=0i,rssi=-89i,snr=4.25",
				"telemetry,device=fencyboy-0,field=ActiveMode,sensor=fencyboy boolValue=true",
				"telemetry,device=fencyboy-0,field=BatteryVoltage,sensor=fencyboy,unit=V floatValue=3.361",
				"telemetry,device=fencyboy-0,field=FenceVoltage,sensor=fencyboy,unit=V floatValue=10246",
				"telemetry,device=fencyboy-0,field=FenceVoltageMax,sensor=fencyboy,unit=V floatValue=10368",
				"telemetry,device=fencyboy-0,field=FenceVoltageMin,sensor=fencyboy,unit=V floatValue=10175",
				"telemetry,device=fencyboy-0,field=FenceVoltageStd,sensor=fencyboy floatValue=32.2",
				"telemetry,device=fencyboy-0,field=Impulses,sensor=fencyboy intValue=466i",
				"telemetry,device=fencyboy-0,field=RemainingCapacity,sensor=fencyboy,unit=mAh floatValue=600.8870239257812",
				"telemetry,device=fencyboy-0,field=Temperature,sensor=fencyboy,unit=°C floatValue=5.32",
			},
			// 2025-01-18T09:02:55.274197028Z
			ExpectedTimeStamp: time.Date(2025, time.January, 18, 9, 2, 55, 274197028, time.UTC),
		},
	}

	if h, err := GetHandler("ttn-fencyboy"); err != nil {
		t.Errorf("did not expect an error while getting handler: %s", err)
	} else {
		testStimuliResponse(t, mockCtrl, mockConfig, mockTMConfig, h, stimuli)
	}
}
