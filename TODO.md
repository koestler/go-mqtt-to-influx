* Write statistics to influxdb
* Implement shelly converter
* make bucket dynamic instead
* check what happens mqtt server disconnects / reconnects
* handle unreachable influxdb, store data in local database
* refactor mqttClients into mqtt client pool (same as influx)
* handle mqtt reconnect:
```
MqttDebugLog: 2022/08/29 09:21:15 [client]   Connect comms goroutine - error triggered EOF
2022/08/29 09:21:17 mqttClient[0-ttn]: set availability to topic='v3/piegn@ttn/tele/koestler-srv5-mqtt-to-influx/status', payload='online'
```

* ttn-dragino; handle case:
```
2022/08/29 08:54:12 ttn-dragino[0-ttn-dragino]: could not extract any sensor data; payload='{"end_device_ids":{"device_id":"lht52-temp-1","application_ids":{"application_id":"piegn"},"dev_eui":"A8404146C184579B","join_eui":"A840410000000100","dev_addr":"260BA38C"},"correlation_ids":["as:up:01GBMCXVYBMZZFHZHFGKDCKSQ3","gs:conn:01GBMCN55RTR7JQEMJHEFFWZ45","gs:up:host:01GBMCN5609J54R7HATNS2Y6YQ","gs:uplink:01GBMCXVQWV4W3NEDZGTTNTVKR","ns:uplink:01GBMCXVQXNW58CRTTQQGQ652F","rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01GBMCXVQXC7TPXZZX58F7CF7Z","rpc:/ttn.lorawan.v3.NsAs/HandleUplink:01GBMCXVYA4YS3DVKWH32P0FQB"],"received_at":"2022-08-29T08:54:12.171174729Z","uplink_message":{"session_key_id":"AYJyW27X6QEJiBQt2m4Ifg==","f_port":5,"f_cnt":1702,"frm_payload":"CQEAAQAL7g==","decoded_payload":{"Bat_mV":3054,"Firmware_Version":"100","Freq_Band":1,"Sensor_Model":9,"Sub_Band":0},"rx_metadata":[{"gateway_ids":{"gateway_id":"piegn-1","eui":"AC1F09FFFE0609F4"},"time":"2022-08-29T08:54:11.913059949Z","timestamp":283036707,"rssi":-99,"channel_rssi":-99,"snr":11,"uplink_token":"ChUKEwoHcGllZ24tMRIIrB8J//4GCfQQo5j7hgEaDAiz/bGYBhCbtoHMAyC40bWynggqDAiz/bGYBhDt4LCzAw==","received_at":"2022-08-29T08:54:11.938350266Z"}],"settings":{"data_rate":{"lora":{"bandwidth":125000,"spreading_factor":7}},"coding_rate":"4/5","frequency":"867500000","timestamp":283036707,"time":"2022-08-29T08:54:11.913059949Z"},"received_at":"2022-08-29T08:54:11.965576697Z","consumed_airtime":"0.056576s","version_ids":{"brand_id":"dragino","model_id":"lht52","hardware_version":"_unknown_hw_version_","firmware_version":"1.0","band_id":"EU_863_870"},"network_ids":{"net_id":"000013","tenant_id":"ttn","cluster_id":"eu1","cluster_address":"eu1.cloud.thethings.network"}}}'
```