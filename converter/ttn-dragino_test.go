package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influx/v2/converter/mock"
	"testing"
	"time"
)

// example inputs

func TestTtsDragino(t *testing.T) {
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
			Topic:   "v3/piegn@ttn/devices/lsn50-temp-0/up",
			Payload: `{"end_device_ids":{"device_id":"lht52-temp-0","application_ids":{"application_id":"piegn"},"dev_eui":"A8404188A184579F","join_eui":"A840410000000100","dev_addr":"260BC1C2"},"correlation_ids":["as:up:01GBBDKJT4C0CH8SZNZE92HQBT","gs:conn:01GBAYF8X1EA1GRCQ40BQ1VDRY","gs:up:host:01GBAYF8X9AD21GXYEAKZYFXYP","gs:uplink:01GBBDKJKM548KSK2W5P3PT5WS","ns:uplink:01GBBDKJKNE18PQH576PNK6AN7","rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01GBBDKJKN8DCF63NPD4K6Q1DM","rpc:/ttn.lorawan.v3.NsAs/HandleUplink:01GBBDKJT317S4RHZW570PXNFB"],"received_at":"2022-08-25T21:12:53.828332832Z","uplink_message":{"session_key_id":"AYJyWmbZaAS/afxNG6hOKw==","f_port":2,"f_cnt":1445,"frm_payload":"Cz0B9H//AWMH5bY=","decoded_payload":{"Ext":1,"Hum_SHT":50,"Systimestamp":1661461942,"TempC_DS":327.67,"TempC_SHT":28.77},"rx_metadata":[{"gateway_ids":{"gateway_id":"piegn-0","eui":"DCA632FFFEA53817"},"time":"2022-08-25T21:12:53.549410Z","timestamp":2983740355,"rssi":-103,"channel_rssi":-103,"snr":-4.75,"uplink_token":"ChUKEwoHcGllZ24tMBII3KYy//6lOBcQw4fhjgsaDAjVy5+YBhDP3eanAiC446Km680DKgwI1cufmAYQ0Kn9hQI=","gps_time":"2022-08-25T21:12:53.549410Z","received_at":"2022-08-25T21:12:53.604523434Z"}],"settings":{"data_rate":{"lora":{"bandwidth":125000,"spreading_factor":7}},"coding_rate":"4/5","frequency":"868100000","timestamp":2983740355,"time":"2022-08-25T21:12:53.549410Z"},"received_at":"2022-08-25T21:12:53.621251971Z","consumed_airtime":"0.061696s","version_ids":{"brand_id":"dragino","model_id":"lht52","hardware_version":"_unknown_hw_version_","firmware_version":"1.0","band_id":"EU_863_870"},"network_ids":{"net_id":"000013","tenant_id":"ttn","cluster_id":"eu1","cluster_address":"eu1.cloud.thethings.network"}}}`,
			ExpectedLines: []string{
				"clock,device=lsn50-temp-0 timeValue=\"2022-08-25T21:12:22Z\"",
				"lora,devEui=A8404188A184579F,device=lsn50-temp-0,gatewayEui=DCA632FFFEA53817,gatewayId=piegn-0 channelRssi=-103i,consumedAirtimeUs=61696i,gatewayIdx=0i,rssi=-103i,snr=-4.75",
				"telemetry,device=lsn50-temp-0,field=Ext,sensor=lht52 intValue=1i",
				"telemetry,device=lsn50-temp-0,field=HumSHT,sensor=lht52,unit=% floatValue=50",
				"telemetry,device=lsn50-temp-0,field=Systimestamp,sensor=lht52,unit=ms intValue=1661461942i",
				"telemetry,device=lsn50-temp-0,field=TempCDs,sensor=lht52,unit=V floatValue=327.67",
				"telemetry,device=lsn50-temp-0,field=TempCSht,sensor=lht52,unit=°C floatValue=28.77",
			},
			// 2022-08-25T21:12:53.828332832Z
			ExpectedTimeStamp: time.Date(2022, time.August, 25, 21, 12, 53, 828332832, time.UTC),
		},
		{
			Topic:   "v3/piegn@ttn/devices/lsn50-temp-1/up",
			Payload: `{"end_device_ids":{"device_id":"lsn50-temp-1","application_ids":{"application_id":"piegn"},"dev_eui":"A84041291183FF1B","join_eui":"A840410000000101","dev_addr":"260BA906"},"correlation_ids":["as:up:01GBBDCNV7JZSD6HDPZYC0XQ53","gs:conn:01GBAYF8X1EA1GRCQ40BQ1VDRY","gs:up:host:01GBAYF8X9AD21GXYEAKZYFXYP","gs:uplink:01GBBDCNMR3JERDZ3E1ZQ7S5E5","ns:uplink:01GBBDCNMS6R8KJ0H95NFXD5JH","rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01GBBDCNMSXAQCKYS2MG49HFTH","rpc:/ttn.lorawan.v3.NsAs/HandleUplink:01GBBDCNV6YHQ0SGHFHCCFWJV7"],"received_at":"2022-08-25T21:09:07.559134515Z","uplink_message":{"session_key_id":"AYJp+jd6NpouOZVJ1CunJQ==","f_port":2,"f_cnt":1522,"frm_payload":"DkcBGwAADADMAXc=","decoded_payload":{"ALARM_status":"FALSE","BatV":3.655,"Temp_Black":37.5,"Temp_Red":28.3,"Temp_White":20.4,"Work_mode":"DS18B20"},"rx_metadata":[{"gateway_ids":{"gateway_id":"piegn-0","eui":"DCA632FFFEA53817"},"time":"2022-08-25T21:09:07.296480Z","timestamp":2757487388,"rssi":-33,"channel_rssi":-33,"snr":7.75,"uplink_token":"ChUKEwoHcGllZ24tMBII3KYy//6lOBcQnNbvogoaDAjzyZ+YBhCH0qCoASDgyrm4oMcDKgwI88mfmAYQgNqvjQE=","gps_time":"2022-08-25T21:09:07.296480Z","received_at":"2022-08-25T21:09:07.338255833Z"}],"settings":{"data_rate":{"lora":{"bandwidth":125000,"spreading_factor":7}},"coding_rate":"4/5","frequency":"867500000","timestamp":2757487388,"time":"2022-08-25T21:09:07.296480Z"},"received_at":"2022-08-25T21:09:07.353637671Z","consumed_airtime":"0.061696s","version_ids":{"brand_id":"dragino","model_id":"lsn50v2-d20-d22-d23","hardware_version":"_unknown_hw_version_","firmware_version":"1.7.4","band_id":"EU_863_870"},"network_ids":{"net_id":"000013","tenant_id":"ttn","cluster_id":"eu1","cluster_address":"eu1.cloud.thethings.network"}}}`,
			ExpectedLines: []string{
				"lora,devEui=A84041291183FF1B,device=lsn50-temp-1,gatewayEui=DCA632FFFEA53817,gatewayId=piegn-0 channelRssi=-33i,consumedAirtimeUs=61696i,gatewayIdx=0i,rssi=-33i,snr=7.75",
				"telemetry,device=lsn50-temp-1,field=AlarmStatus,sensor=lsn50v2-d20-d22-d23 boolValue=false",
				"telemetry,device=lsn50-temp-1,field=BatV,sensor=lsn50v2-d20-d22-d23,unit=V floatValue=3.655",
				"telemetry,device=lsn50-temp-1,field=TempBlack,sensor=lsn50v2-d20-d22-d23,unit=°C floatValue=37.5",
				"telemetry,device=lsn50-temp-1,field=TempRed,sensor=lsn50v2-d20-d22-d23,unit=°C floatValue=28.3",
				"telemetry,device=lsn50-temp-1,field=TempWhite,sensor=lsn50v2-d20-d22-d23,unit=°C floatValue=20.4",
				"telemetry,device=lsn50-temp-1,field=WorkMode,sensor=lsn50v2-d20-d22-d23 stringValue=\"DS18B20\"",
			},
			// 2022-08-25T21:09:07.559134515Z
			ExpectedTimeStamp: time.Date(2022, time.August, 25, 21, 9, 7, 559134515, time.UTC),
		},
		{
			Topic:   "v3/piegn@ttn/devices/lsn50-temp-2/up",
			Payload: `{"end_device_ids":{"device_id":"lsn50-temp-2","application_ids":{"application_id":"piegn"},"dev_eui":"A840415E818498DF","join_eui":"A840410000000101","dev_addr":"260B3B83"},"correlation_ids":["as:up:01GBBDJ4KD0250B9QNZWFYJ2PT","gs:conn:01GBAYF8X1EA1GRCQ40BQ1VDRY","gs:up:host:01GBAYF8X9AD21GXYEAKZYFXYP","gs:uplink:01GBBDJ4CWEEZC6X87M4S2GKNK","ns:uplink:01GBBDJ4CXYVTT6SPGKTETPES9","rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01GBBDJ4CXMK7BZ0FACF8V03MQ","rpc:/ttn.lorawan.v3.NsAs/HandleUplink:01GBBDJ4KCM2Z8HNFBJ3SBP97G"],"received_at":"2022-08-25T21:12:06.509402131Z","uplink_message":{"session_key_id":"AYJp/NotlCyht+iBV0/mfA==","f_port":2,"f_cnt":1522,"frm_payload":"DlQAAAERAADdAnM=","decoded_payload":{"ADC_CH0V":0.273,"BatV":3.668,"Digital_IStatus":"L","Door_status":"OPEN","EXTI_Trigger":"FALSE","Hum_SHT":62.7,"TempC1":0,"TempC_SHT":22.1,"Work_mode":"IIC"},"rx_metadata":[{"gateway_ids":{"gateway_id":"piegn-0","eui":"DCA632FFFEA53817"},"time":"2022-08-25T21:12:06.248346Z","timestamp":2936439283,"rssi":-91,"channel_rssi":-91,"snr":8.25,"uplink_token":"ChUKEwoHcGllZ24tMBII3KYy//6lOBcQ84Oa+AoaDAimy5+YBhDi+raPASC4uq6Lu8wDKgsIpsufmAYQkOu1dg==","gps_time":"2022-08-25T21:12:06.248346Z","received_at":"2022-08-25T21:12:06.286328391Z"}],"settings":{"data_rate":{"lora":{"bandwidth":125000,"spreading_factor":7}},"coding_rate":"4/5","frequency":"867100000","timestamp":2936439283,"time":"2022-08-25T21:12:06.248346Z"},"received_at":"2022-08-25T21:12:06.301646615Z","consumed_airtime":"0.061696s","version_ids":{"brand_id":"dragino","model_id":"lsn50v2-s31","hardware_version":"_unknown_hw_version_","firmware_version":"1.0","band_id":"EU_863_870"},"network_ids":{"net_id":"000013","tenant_id":"ttn","cluster_id":"eu1","cluster_address":"eu1.cloud.thethings.network"}}}`,
			ExpectedLines: []string{
				"lora,devEui=A840415E818498DF,device=lsn50-temp-2,gatewayEui=DCA632FFFEA53817,gatewayId=piegn-0 channelRssi=-91i,consumedAirtimeUs=61696i,gatewayIdx=0i,rssi=-91i,snr=8.25",
				"telemetry,device=lsn50-temp-2,field=AdcCh0,sensor=lsn50v2-s31,unit=V floatValue=0.273",
				"telemetry,device=lsn50-temp-2,field=BatV,sensor=lsn50v2-s31,unit=V floatValue=3.668",
				"telemetry,device=lsn50-temp-2,field=DigitalIStatus,sensor=lsn50v2-s31 stringValue=\"L\"",
				"telemetry,device=lsn50-temp-2,field=DoorStatus,sensor=lsn50v2-s31 stringValue=\"OPEN\"",
				"telemetry,device=lsn50-temp-2,field=EXTITrigger,sensor=lsn50v2-s31 boolValue=false",
				"telemetry,device=lsn50-temp-2,field=HumSHT,sensor=lsn50v2-s31,unit=% floatValue=62.7",
				"telemetry,device=lsn50-temp-2,field=TempC1,sensor=lsn50v2-s31,unit=°C floatValue=0",
				"telemetry,device=lsn50-temp-2,field=TempCSht,sensor=lsn50v2-s31,unit=°C floatValue=22.1",
				"telemetry,device=lsn50-temp-2,field=WorkMode,sensor=lsn50v2-s31 stringValue=\"IIC\"",
			},
			// 2022-08-25T21:12:06.509402131Z
			ExpectedTimeStamp: time.Date(2022, time.August, 25, 21, 12, 6, 509402131, time.UTC),
		},
		{
			Topic:   "v3/piegn@ttn/devices/lsn50-temp-3/up",
			Payload: `{"end_device_ids":{"device_id":"lsn50-temp-3","application_ids":{"application_id":"piegn"},"dev_eui":"A840417E618498EF","join_eui":"A840410000000101","dev_addr":"260B2A65"},"correlation_ids":["as:up:01GBBDHDDXJWBTA1MM01F0AKJS","gs:conn:01GBAYF8X1EA1GRCQ40BQ1VDRY","gs:up:host:01GBAYF8X9AD21GXYEAKZYFXYP","gs:uplink:01GBBDHD7CYYYYAD5V9D7RXZAM","ns:uplink:01GBBDHD7D33J0PB72JVJDM1Y6","rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01GBBDHD7D5VS1MZXDKSP77HMY","rpc:/ttn.lorawan.v3.NsAs/HandleUplink:01GBBDHDDW066XPCSMN657A1XJ"],"received_at":"2022-08-25T21:11:42.781705807Z","uplink_message":{"session_key_id":"AYJp++HkYeG6lUhMoF3AEw==","f_port":2,"f_cnt":1522,"frm_payload":"DiD//wAADP////8=","decoded_payload":{"ADC_CH0V":0,"BatV":3.616,"Digital_IStatus":"L","Door_status":"OPEN","EXTI_Trigger":"FALSE","TempC1":-0.1,"TempC2":-0.1,"TempC3":-0.1,"Work_mode":"3DS18B20"},"rx_metadata":[{"gateway_ids":{"gateway_id":"piegn-0","eui":"DCA632FFFEA53817"},"time":"2022-08-25T21:11:42.515414Z","timestamp":2912706347,"rssi":-101,"channel_rssi":-101,"snr":0.25,"uplink_token":"ChUKEwoHcGllZ24tMBII3KYy//6lOBcQq77x7AoaDAiOy5+YBhDe5oyRAiD4/87W4ssDKgwIjsufmAYQ8K/i9QE=","gps_time":"2022-08-25T21:11:42.515414Z","received_at":"2022-08-25T21:11:42.559622221Z"}],"settings":{"data_rate":{"lora":{"bandwidth":125000,"spreading_factor":7}},"coding_rate":"4/5","frequency":"867300000","timestamp":2912706347,"time":"2022-08-25T21:11:42.515414Z"},"received_at":"2022-08-25T21:11:42.573526505Z","consumed_airtime":"0.061696s","version_ids":{"brand_id":"dragino","model_id":"lsn50v2-d20","hardware_version":"_unknown_hw_version_","firmware_version":"1.1","band_id":"EU_863_870"},"network_ids":{"net_id":"000013","tenant_id":"ttn","cluster_id":"eu1","cluster_address":"eu1.cloud.thethings.network"}}}`,
			ExpectedLines: []string{
				"lora,devEui=A840417E618498EF,device=lsn50-temp-3,gatewayEui=DCA632FFFEA53817,gatewayId=piegn-0 channelRssi=-101i,consumedAirtimeUs=61696i,gatewayIdx=0i,rssi=-101i,snr=0.25",
				"telemetry,device=lsn50-temp-3,field=AdcCh0,sensor=lsn50v2-d20,unit=V floatValue=0",
				"telemetry,device=lsn50-temp-3,field=BatV,sensor=lsn50v2-d20,unit=V floatValue=3.616",
				"telemetry,device=lsn50-temp-3,field=DigitalIStatus,sensor=lsn50v2-d20 stringValue=\"L\"",
				"telemetry,device=lsn50-temp-3,field=DoorStatus,sensor=lsn50v2-d20 stringValue=\"OPEN\"",
				"telemetry,device=lsn50-temp-3,field=EXTITrigger,sensor=lsn50v2-d20 boolValue=false",
				"telemetry,device=lsn50-temp-3,field=TempC1,sensor=lsn50v2-d20,unit=°C floatValue=-0.1",
				"telemetry,device=lsn50-temp-3,field=TempC2,sensor=lsn50v2-d20,unit=°C floatValue=-0.1",
				"telemetry,device=lsn50-temp-3,field=TempC3,sensor=lsn50v2-d20,unit=°C floatValue=-0.1",
				"telemetry,device=lsn50-temp-3,field=WorkMode,sensor=lsn50v2-d20 stringValue=\"3DS18B20\"",
			},
			// 2022-08-25T21:11:42.781705807Z
			ExpectedTimeStamp: time.Date(2022, time.August, 25, 21, 11, 42, 781705807, time.UTC),
		},
		{
			Topic:   "v3/piegn@ttn/devices/lht52-temp-1/up",
			Payload: `{"end_device_ids":{"device_id":"lht52-temp-1","application_ids":{"application_id":"piegn"},"dev_eui":"A8404146C184579B","join_eui":"A840410000000100","dev_addr":"260BA38C"},"correlation_ids":["as:up:01GBBDNNZ9XK95TVR86SD4X9Z5","gs:conn:01GBAYF8X1EA1GRCQ40BQ1VDRY","gs:up:host:01GBAYF8X9AD21GXYEAKZYFXYP","gs:uplink:01GBBDNNRSE3RT3PRTZXEH0WRT","ns:uplink:01GBBDNNRST5DVJH16P41Q52X0","rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01GBBDNNRSYP2RF1AZJSS3N5XH","rpc:/ttn.lorawan.v3.NsAs/HandleUplink:01GBBDNNZ8AFXVRJYXPE29FDQ6"],"received_at":"2022-08-25T21:14:02.601203717Z","uplink_message":{"session_key_id":"AYJyW27X6QEJiBQt2m4Ifg==","f_port":2,"f_cnt":1445,"frm_payload":"C2sB2X//AWMH5fo=","decoded_payload":{"Ext":1,"Hum_SHT":47.3,"Systimestamp":1661462010,"TempC_DS":327.67,"TempC_SHT":29.23},"rx_metadata":[{"gateway_ids":{"gateway_id":"piegn-0","eui":"DCA632FFFEA53817"},"time":"2022-08-25T21:14:02.360206Z","timestamp":3052551163,"rssi":-94,"channel_rssi":-94,"snr":3,"uplink_token":"ChUKEwoHcGllZ24tMBII3KYy//6lOBcQ+/fIrwsaDAiazJ+YBhCx+qq7ASD4mOnR688DKgwImsyfmAYQsJ3hqwE=","gps_time":"2022-08-25T21:14:02.360206Z","received_at":"2022-08-25T21:14:02.378406422Z"}],"settings":{"data_rate":{"lora":{"bandwidth":125000,"spreading_factor":7}},"coding_rate":"4/5","frequency":"868500000","timestamp":3052551163,"time":"2022-08-25T21:14:02.360206Z"},"received_at":"2022-08-25T21:14:02.393727916Z","consumed_airtime":"0.061696s","version_ids":{"brand_id":"dragino","model_id":"lht52","hardware_version":"_unknown_hw_version_","firmware_version":"1.0","band_id":"EU_863_870"},"network_ids":{"net_id":"000013","tenant_id":"ttn","cluster_id":"eu1","cluster_address":"eu1.cloud.thethings.network"}}}`,
			ExpectedLines: []string{
				"clock,device=lht52-temp-1 timeValue=\"2022-08-25T21:13:30Z\"",
				"lora,devEui=A8404146C184579B,device=lht52-temp-1,gatewayEui=DCA632FFFEA53817,gatewayId=piegn-0 channelRssi=-94i,consumedAirtimeUs=61696i,gatewayIdx=0i,rssi=-94i,snr=3",
				"telemetry,device=lht52-temp-1,field=Ext,sensor=lht52 intValue=1i",
				"telemetry,device=lht52-temp-1,field=HumSHT,sensor=lht52,unit=% floatValue=47.3",
				"telemetry,device=lht52-temp-1,field=Systimestamp,sensor=lht52,unit=ms intValue=1661462010i",
				"telemetry,device=lht52-temp-1,field=TempCDs,sensor=lht52,unit=V floatValue=327.67",
				"telemetry,device=lht52-temp-1,field=TempCSht,sensor=lht52,unit=°C floatValue=29.23",
			},
			// 2022-08-25T21:14:02.601203717Z
			ExpectedTimeStamp: time.Date(2022, time.August, 25, 21, 14, 2, 601203717, time.UTC),
		},
	}

	if h, err := GetHandler("ttn-dragino"); err != nil {
		t.Errorf("did not expect an error while getting handler: %s", err)
	} else {
		testStimuliResponse(t, mockCtrl, mockConfig, mockTMConfig, h, stimuli)
	}
}
