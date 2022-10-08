* Write statistics to influxdb
* Implement shelly converter
* make bucket dynamic instead
* check what happens mqtt server disconnects / reconnects
* handle unreachable influxdb, store data in local database
* handle mqtt reconnect:
```
MqttDebugLog: 2022/08/29 09:21:15 [client]   Connect comms goroutine - error triggered EOF
2022/08/29 09:21:17 mqttClient[0-ttn]: set availability to topic='v3/piegn@ttn/tele/koestler-srv5-mqtt-to-influx/status', payload='online'
```

* ttn-dragino; handle case:
```
2022/08/29 08:54:12 ttn-dragino[0-ttn-dragino]: could not extract any sensor data; payload='{"end_device_ids":{"device_id":"lht52-temp-1","application_ids":{"application_id":"piegn"},"dev_eui":"A8404146C184579B","join_eui":"A840410000000100","dev_addr":"260BA38C"},"correlation_ids":["as:up:01GBMCXVYBMZZFHZHFGKDCKSQ3","gs:conn:01GBMCN55RTR7JQEMJHEFFWZ45","gs:up:host:01GBMCN5609J54R7HATNS2Y6YQ","gs:uplink:01GBMCXVQWV4W3NEDZGTTNTVKR","ns:uplink:01GBMCXVQXNW58CRTTQQGQ652F","rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01GBMCXVQXC7TPXZZX58F7CF7Z","rpc:/ttn.lorawan.v3.NsAs/HandleUplink:01GBMCXVYA4YS3DVKWH32P0FQB"],"received_at":"2022-08-29T08:54:12.171174729Z","uplink_message":{"session_key_id":"AYJyW27X6QEJiBQt2m4Ifg==","f_port":5,"f_cnt":1702,"frm_payload":"CQEAAQAL7g==","decoded_payload":{"Bat_mV":3054,"Firmware_Version":"100","Freq_Band":1,"Sensor_Model":9,"Sub_Band":0},"rx_metadata":[{"gateway_ids":{"gateway_id":"piegn-1","eui":"AC1F09FFFE0609F4"},"time":"2022-08-29T08:54:11.913059949Z","timestamp":283036707,"rssi":-99,"channel_rssi":-99,"snr":11,"uplink_token":"ChUKEwoHcGllZ24tMRIIrB8J//4GCfQQo5j7hgEaDAiz/bGYBhCbtoHMAyC40bWynggqDAiz/bGYBhDt4LCzAw==","received_at":"2022-08-29T08:54:11.938350266Z"}],"settings":{"data_rate":{"lora":{"bandwidth":125000,"spreading_factor":7}},"coding_rate":"4/5","frequency":"867500000","timestamp":283036707,"time":"2022-08-29T08:54:11.913059949Z"},"received_at":"2022-08-29T08:54:11.965576697Z","consumed_airtime":"0.056576s","version_ids":{"brand_id":"dragino","model_id":"lht52","hardware_version":"_unknown_hw_version_","firmware_version":"1.0","band_id":"EU_863_870"},"network_ids":{"net_id":"000013","tenant_id":"ttn","cluster_id":"eu1","cluster_address":"eu1.cloud.thethings.network"}}}'
```
* bug during shutdown, writePoint is executed after channel is closed
* iotdevice / lora: use time of message and not local time when possible
* refactor empty influx client lists / empty mqtt client lists into config module


Documentation:
* Set Qos=0 to connect thethings.network
* shutdown:
```
  panic: send on closed channel

goroutine 83 [running]:
github.com/influxdata/influxdb-client-go/v2/api.(*WriteAPIImpl).WritePoint(0xc000132600, 0xc000568a20?)
/home/lk/go/pkg/mod/github.com/influxdata/influxdb-client-go/v2@v2.10.0/api/write.go:261 +0x16c
github.com/koestler/go-mqtt-to-influx/influxClient.Client.WritePoint({{0x8c9748, 0xc0001b13b0}, {0x8caad8, 0xc000190480}, {0x8c8640, 0xc000132600}, {0xc00017adc0, 0x16, 0x16}, {0x7f8c274b3880, ...}, ...}, ...)
/home/lk/git/go-mqtt-to-influx/influxClient/influxClient.go:133 +0x83
github.com/koestler/go-mqtt-to-influx/influxClient.(*ClientPool).WritePoint(0xc000280400?, {0x7f8c2429ef18, 0xc00007a1c0}, {0x0?, 0xc0005cf938?, 0x40c4ab?})
/home/lk/git/go-mqtt-to-influx/influxClient/InfluxClientPool.go:78 +0xe8
main.getMqttMessageHandler.func1.1({0x8c8398?, 0xc00007a1c0})
/home/lk/git/go-mqtt-to-influx/converter.go:91 +0x86
github.com/koestler/go-mqtt-to-influx/converter.tasmotaStateHandler({0x8c8130, 0xc0001c0700}, {0x8c6e90, 0xc000578078}, {0x8c6ee0, 0xc000568570}, 0xc00057a120)
/home/lk/git/go-mqtt-to-influx/converter/tasmotaState.go:70 +0x47c
main.getMqttMessageHandler.func1({{0xc0005720c0, 0x1e}, {0xc0003cc800, 0x96, 0x200}})
/home/lk/git/go-mqtt-to-influx/converter.go:86 +0x273
github.com/koestler/go-mqtt-to-influx/mqttClient.(*Client).AddRoute.func1(0xc00007a180)
/home/lk/git/go-mqtt-to-influx/mqttClient/mqttClient.go:130 +0x116
github.com/eclipse/paho.golang/paho.(*StandardRouter).Route(0xc00007ab00, 0xc00007a100)
/home/lk/go/pkg/mod/github.com/eclipse/paho.golang@v0.10.0/paho/router.go:97 +0x5f2
github.com/eclipse/paho.golang/paho.(*Client).routePublishPackets(0xc0002ba000)
/home/lk/go/pkg/mod/github.com/eclipse/paho.golang@v0.10.0/paho/client.go:382 +0xf9
github.com/eclipse/paho.golang/paho.(*Client).Connect.func3()
/home/lk/go/pkg/mod/github.com/eclipse/paho.golang@v0.10.0/paho/client.go:293 +0xba
created by github.com/eclipse/paho.golang/paho.(*Client).Connect
/home/lk/go/pkg/mod/github.com/eclipse/paho.golang@v0.10.0/paho/client.go:290 +0xb7c
➜  go-mqtt-to-influx git:(main) ✗ 
```

* db: use gzip.CompressWithGzip ?