package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influx/v2/converter/mock"
	"testing"
	"time"
)

// example inputs

func TestTtnSensecap(t *testing.T) {
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
			Topic: "v3/piegn@ttn/devices/s2120-0/up",
			Payload: `{
    "end_device_ids": {
        "device_id": "s2120-0",
        "application_ids": {
            "application_id": "piegn"
        },
        "dev_eui": "2CF7F1C0443003DD",
        "join_eui": "53D800BB13E0ED90",
        "dev_addr": "260BC756"
    },
    "correlation_ids": [
        "gs:uplink:01HMPDTHDAX6ERYA9PDK7CSSJ4"
    ],
    "received_at": "2024-01-21T16:31:55.513418143Z",
    "uplink_message": {
        "session_key_id": "AY0sCQ6M4uie08eWl9Uo+g==",
        "f_port": 3,
        "f_cnt": 58,
        "frm_payload": "AQASIwAAAHAAAAACACoAAAAAJhg=",
        "decoded_payload": {
            "err": 0,
            "messages": [
                {
                    "measurementId": "4097",
                    "measurementValue": 1.8,
                    "type": "Air Temperature"
                },
                {
                    "measurementId": "4098",
                    "measurementValue": 35,
                    "type": "Air Humidity"
                },
                {
                    "measurementId": "4099",
                    "measurementValue": 112,
                    "type": "Light Intensity"
                },
                {
                    "measurementId": "4190",
                    "measurementValue": 0,
                    "type": "UV Index"
                },
                {
                    "measurementId": "4105",
                    "measurementValue": 0,
                    "type": "Wind Speed"
                },
                {
                    "measurementId": "4104",
                    "measurementValue": 42,
                    "type": "Wind Direction Sensor"
                },
                {
                    "measurementId": "4113",
                    "measurementValue": 0,
                    "type": "Rain Gauge"
                },
                {
                    "measurementId": "4101",
                    "measurementValue": 97520,
                    "type": "Barometric Pressure"
                }
            ],
            "payload": "010012230000007000000002002A000000002618",
            "valid": true
        },
        "rx_metadata": [
            {
                "gateway_ids": {
                    "gateway_id": "piegn-srv3",
                    "eui": "E45F01FFFEDECBE3"
                },
                "time": "2024-01-21T16:31:55.184858083Z",
                "timestamp": 3570357122,
                "rssi": -82,
                "channel_rssi": -82,
                "snr": 5.75,
                "uplink_token": "ChgKFgoKcGllZ24tc3J2MxII5F8B\/\/7ey+MQgqe9pg0aDAj7jbWtBhC4\/\/mRASDQ5\/jO9LES",
                "received_at": "2024-01-21T16:31:55.264975183Z"
            }
        ],
        "settings": {
            "data_rate": {
                "lora": {
                    "bandwidth": 125000,
                    "spreading_factor": 7,
                    "coding_rate": "4\/5"
                }
            },
            "frequency": "867900000",
            "timestamp": 3570357122,
            "time": "2024-01-21T16:31:55.184858083Z"
        },
        "received_at": "2024-01-21T16:31:55.307162234Z",
        "confirmed": true,
        "consumed_airtime": "0.071936s",
        "version_ids": {
            "brand_id": "sensecap",
            "model_id": "sensecaps2120-8-in-1",
            "hardware_version": "1.0",
            "firmware_version": "1.0",
            "band_id": "EU_863_870"
        },
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
				"lora,devEui=2CF7F1C0443003DD,device=s2120-0,gatewayEui=E45F01FFFEDECBE3,gatewayId=piegn-srv3 channelRssi=-82i,consumedAirtimeUs=71936i,gatewayIdx=0i,rssi=-82i,snr=5.75",
				"telemetry,device=s2120-0,field=Air\\ Humidity,sensor=sensecaps2120-8-in-1,unit=%\\ RH floatValue=35",
				"telemetry,device=s2120-0,field=Air\\ Temperature,sensor=sensecaps2120-8-in-1,unit=°C floatValue=1.8",
				"telemetry,device=s2120-0,field=Barometric\\ Pressure,sensor=sensecaps2120-8-in-1,unit=Pa floatValue=97520",
				"telemetry,device=s2120-0,field=Light\\ Intensity,sensor=sensecaps2120-8-in-1,unit=Lux floatValue=112",
				"telemetry,device=s2120-0,field=Rainfall,sensor=sensecaps2120-8-in-1,unit=mm/h floatValue=0",
				"telemetry,device=s2120-0,field=UV\\ Index,sensor=sensecaps2120-8-in-1,unit= floatValue=0",
				"telemetry,device=s2120-0,field=Wind\\ Direction,sensor=sensecaps2120-8-in-1,unit=° floatValue=42",
				"telemetry,device=s2120-0,field=Wind\\ Speed,sensor=sensecaps2120-8-in-1,unit=m/s floatValue=0",
			},
			// 2024-01-21T16:31:55.513418143Z
			ExpectedTimeStamp: time.Date(2024, time.January, 21, 16, 31, 55, 513418143, time.UTC),
		},
		{
			Topic: "v3/piegn@ttn/devices/s2120-0/up",
			Payload: `{
    "end_device_ids": {
        "device_id": "s2120-0",
        "application_ids": {
            "application_id": "piegn"
        },
        "dev_eui": "2CF7F1C0443003DD",
        "join_eui": "53D800BB13E0ED90",
        "dev_addr": "260BC756"
    },
    "correlation_ids": [
        "gs:uplink:01HMYKE4MT8Y8660D39MV6PX9K"
    ],
    "received_at": "2024-01-24T20:43:56.136229322Z",
    "uplink_message": {
        "session_key_id": "AY0sCQ6M4uie08eWl9Uo+g==",
        "f_port": 3,
        "f_cnt": 1041,
        "frm_payload": "AQBBQgAAAAAAAAACAE8AAAAAJdw=",
        "decoded_payload": {
            "err": 0,
            "messages": [
                {
                    "measurementId": "4097",
                    "measurementValue": 6.5,
                    "type": "Air Temperature"
                },
                {
                    "measurementId": "4098",
                    "measurementValue": 66,
                    "type": "Air Humidity"
                },
                {
                    "measurementId": "4099",
                    "measurementValue": 0,
                    "type": "Light Intensity"
                },
                {
                    "measurementId": "4190",
                    "measurementValue": 0,
                    "type": "UV Index"
                },
                {
                    "measurementId": "4105",
                    "measurementValue": 0,
                    "type": "Wind Speed"
                },
                {
                    "measurementId": "4104",
                    "measurementValue": 79,
                    "type": "Wind Direction Sensor"
                },
                {
                    "measurementId": "4113",
                    "measurementValue": 0,
                    "type": "Rain Gauge"
                },
                {
                    "measurementId": "4101",
                    "measurementValue": 96920,
                    "type": "Barometric Pressure"
                }
            ],
            "payload": "010041420000000000000002004F0000000025DC",
            "valid": true
        },
        "rx_metadata": [
            {
                "gateway_ids": {
                    "gateway_id": "piegn-srv3",
                    "eui": "E45F01FFFEDECBE3"
                },
                "time": "2024-01-24T20:43:55.866967916Z",
                "timestamp": 2570298224,
                "rssi": -80,
                "channel_rssi": -80,
                "snr": 6.5,
                "uplink_token": "ChgKFgoKcGllZ24tc3J2MxII5F8B\/\/7ey+MQ8MbOyQkaDAiL7cWtBhDe3Mq7AyCA29mN58QC",
                "received_at": "2024-01-24T20:43:55.895127767Z"
            }
        ],
        "settings": {
            "data_rate": {
                "lora": {
                    "bandwidth": 125000,
                    "spreading_factor": 7,
                    "coding_rate": "4\/5"
                }
            },
            "frequency": "867700000",
            "timestamp": 2570298224,
            "time": "2024-01-24T20:43:55.866967916Z"
        },
        "received_at": "2024-01-24T20:43:55.931223662Z",
        "confirmed": true,
        "consumed_airtime": "0.071936s",
        "version_ids": {
            "brand_id": "sensecap",
            "model_id": "sensecaps2120-8-in-1",
            "hardware_version": "1.0",
            "firmware_version": "1.0",
            "band_id": "EU_863_870"
        },
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
				"lora,devEui=2CF7F1C0443003DD,device=s2120-0,gatewayEui=E45F01FFFEDECBE3,gatewayId=piegn-srv3 channelRssi=-80i,consumedAirtimeUs=71936i,gatewayIdx=0i,rssi=-80i,snr=6.5",
				"telemetry,device=s2120-0,field=Air\\ Humidity,sensor=sensecaps2120-8-in-1,unit=%\\ RH floatValue=66",
				"telemetry,device=s2120-0,field=Air\\ Temperature,sensor=sensecaps2120-8-in-1,unit=°C floatValue=6.5",
				"telemetry,device=s2120-0,field=Barometric\\ Pressure,sensor=sensecaps2120-8-in-1,unit=Pa floatValue=96920",
				"telemetry,device=s2120-0,field=Light\\ Intensity,sensor=sensecaps2120-8-in-1,unit=Lux floatValue=0",
				"telemetry,device=s2120-0,field=Rainfall,sensor=sensecaps2120-8-in-1,unit=mm/h floatValue=0",
				"telemetry,device=s2120-0,field=UV\\ Index,sensor=sensecaps2120-8-in-1,unit= floatValue=0",
				"telemetry,device=s2120-0,field=Wind\\ Direction,sensor=sensecaps2120-8-in-1,unit=° floatValue=79",
				"telemetry,device=s2120-0,field=Wind\\ Speed,sensor=sensecaps2120-8-in-1,unit=m/s floatValue=0",
			},
			// 2024-01-24T20:43:56.136229322Z
			ExpectedTimeStamp: time.Date(2024, time.January, 24, 20, 43, 56, 136229322, time.UTC),
		},
	}

	if h, err := GetHandler("ttn"); err != nil {
		t.Errorf("did not expect an error while getting handler: %s", err)
	} else {
		testStimuliResponse(t, mockCtrl, mockConfig, mockTMConfig, h, stimuli)
	}
}
