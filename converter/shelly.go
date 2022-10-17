package converter

import (
	"encoding/json"
	"log"
	"time"
)

type shellyMessage struct {
}

func init() {
	registerHandler("shelly", shellyHandler)
}

// examples
// shelly 1
// shellies/shelly1/online true
// shellies/announce {"id":"shelly1","model":"SHSW-1","mac":"8CAAB54B98DF","ip":"192.168.8.93","new_fw":false,"fw_ver":"20220809-123240/v1.12-g99f7e0b"}
// shellies/shelly1/announce {"id":"shelly1","model":"SHSW-1","mac":"8CAAB54B98DF","ip":"192.168.8.93","new_fw":false,"fw_ver":"20220809-123240/v1.12-g99f7e0b"}
// shellies/shelly1/info {"wifi_sta":{"connected":true,"ssid":"piegn-iot","ip":"192.168.8.93","rssi":-59},"cloud":{"enabled":false,"connected":false},"mqtt":{"connected":true},"time":"","unixtime":0,"serial":1,"has_update":false,"mac":"8CAAB54B98DF","cfg_changed_cnt":0,"actions_stats":{"skipped":0},"relays":[{"ison":false,"has_timer":false,"timer_started":0,"timer_duration":0,"timer_remaining":0,"source":"input"}],"meters":[{"power":0.00,"is_valid":true}],"inputs":[{"input":0,"event":"","event_cnt":0}],"ext_sensors":{},"ext_temperature":{},"ext_humidity":{},"update":{"status":"unknown","has_update":false,"new_version":"","old_version":"20220809-123240/v1.12-g99f7e0b"},"ram_total":51688,"ram_free":40892,"fs_size":233681,"fs_free":150600,"uptime":1}
// shellies/shelly1/relay/0 off
// shellies/shelly1/input/0 0
// shellies/shelly1/input_event/0 {"event":"","event_cnt":0}
// shellies/shelly1/relay/0 on
// shellies/shelly1/relay/0 off
// shellies/shelly1/online false
//
// shelly 1pm
// shellies/plus1pm/events/rpc {"src":"shellyplus1pm-4417939341dc","dst":"shellies/plus1pm/events","method":"NotifyStatus","params":{"ts":1.61,"mqtt":{"connected":true}}}
// shellies/plus1pm/online true
// shellies/plus1pm/status/mqtt {"connected":true}
// shellies/plus1pm/events/rpc {"src":"shellyplus1pm-4417939341dc","dst":"shellies/plus1pm/events","method":"NotifyStatus","params":{"ts":1661517385.44,"switch:0":{"id":0,"voltage":232.7}}}
// shellies/plus1pm/status/switch:0 {"id":0, "source":"init", "output":false, "apower":0.0, "voltage":232.7, "current":0.000, "aenergy":{"total":0.000,"by_minute":[0.000,0.000,0.000],"minute_ts":1661517385},"temperature":{"tC":39.2, "tF":102.5}}
// shellies/plus1pm/events/rpc {"src":"shellyplus1pm-4417939341dc","dst":"shellies/plus1pm/events","method":"NotifyStatus","params":{"ts":1661517398.50,"sys":{"available_updates":{"beta":{"version":"0.11.0-beta3"}}}}}
// shellies/plus1pm/status/sys {"mac":"4417939341DC","restart_required":false,"time":"14:36","unixtime":1661517398,"uptime":15,"ram_size":254968,"ram_free":156660,"fs_size":458752,"fs_free":221184,"cfg_rev":8,"available_updates":{"beta":{"version":"0.11.0-beta3"}}}
// shellies/plus1pm/events/rpc {"src":"shellyplus1pm-4417939341dc","dst":"shellies/plus1pm/events","method":"NotifyStatus","params":{"ts":1661517410.45,"switch:0":{"id":0,"temperature":{"tC":44.06,"tF":111.30}}}}
// shellies/plus1pm/status/switch:0 {"id":0, "source":"init", "output":false, "apower":0.0, "voltage":232.7, "current":0.000, "aenergy":{"total":0.000,"by_minute":[0.000,0.000,0.000],"minute_ts":1661517407},"temperature":{"tC":44.1, "tF":111.3}}
// shellies/plus1pm/events/rpc {"src":"shellyplus1pm-4417939341dc","dst":"shellies/plus1pm/events","method":"NotifyStatus","params":{"ts":1661517421.44,"switch:0":{"id":0,"aenergy":{"by_minute":[0.000,0.000,0.000],"minute_ts":1661517419,"total":0.000}}}}
// shellies/plus1pm/status/switch:0 {"id":0, "source":"init", "output":false, "apower":0.0, "voltage":232.8, "current":0.000, "aenergy":{"total":0.000,"by_minute":[0.000,0.000,0.000],"minute_ts":1661517419},"temperature":{"tC":45.2, "tF":113.4}}
// shellies/plus1pm/events/rpc {"src":"shellyplus1pm-4417939341dc","dst":"shellies/plus1pm/events","method":"NotifyStatus","params":{"ts":1661517432.85,"switch:0":{"id":0,"output":true,"source":"WS_in"}}}
// shellies/plus1pm/status/switch:0 {"id":0, "source":"WS_in", "output":true, "apower":0.0, "voltage":232.6, "current":0.000, "aenergy":{"total":0.000,"by_minute":[0.000,0.000,0.000],"minute_ts":1661517429},"temperature":{"tC":46.0, "tF":114.7}}
// shellies/plus1pm/online false
//
// shelly em3
// shellies/shellies/shellyem3/online false
// shellies/shellies/shellyem3/online true
// shellies/announce {"id":"shellies/shellyem3","model":"SHEM-3","mac":"244CAB4354E2","ip":"192.168.8.100","new_fw":false,"fw_ver":"20220415-105853/v1.11.7-25-gb3b096857-v1.11.7-3em"}
// shellies/shellies/shellyem3/announce {"id":"shellies/shellyem3","model":"SHEM-3","mac":"244CAB4354E2","ip":"192.168.8.100","new_fw":false,"fw_ver":"20220415-105853/v1.11.7-25-gb3b096857-v1.11.7-3em"}
// shellies/shellies/shellyem3/info {"wifi_sta":{"connected":true,"ssid":"piegn-iot","ip":"192.168.8.100","rssi":-62},"cloud":{"enabled":false,"connected":false},"mqtt":{"connected":true},"time":"","unixtime":0,"serial":1,"has_update":false,"mac":"244CAB4354E2","cfg_changed_cnt":0,"actions_stats":{"skipped":0},"relays":[{"ison":false,"has_timer":false,"timer_started":0,"timer_duration":0,"timer_remaining":0,"overpower":false,"is_valid":true,"source":"input"}],"emeters":[{"power":0.00,"pf":0.00,"current":0.01,"voltage":230.35,"is_valid":true,"total":0.0,"total_returned":0.0},{"power":0.00,"pf":0.00,"current":0.01,"voltage":0.12,"is_valid":true,"total":0.0,"total_returned":0.0},{"power":0.00,"pf":0.00,"current":0.01,"voltage":0.09,"is_valid":true,"total":0.0,"total_returned":0.0}],"total_power":0.00,"fs_mounted":true,"update":{"status":"unknown","has_update":false,"new_version":"","old_version":"20220415-105853/v1.11.7-25-gb3b096857-v1.11.7-3em"},"ram_total":49288,"ram_free":28776,"fs_size":233681,"fs_free":157126,"uptime":7}
// shellies/shellies/shellyem3/relay/0 off
// shellies/shellies/shellyem3/emeter/0/power 0.00
// shellies/shellies/shellyem3/emeter/0/pf 0.00
// shellies/shellies/shellyem3/emeter/0/current 0.01
// shellies/shellies/shellyem3/emeter/0/voltage 230.35
// shellies/shellies/shellyem3/emeter/0/total 0.0
// shellies/shellies/shellyem3/emeter/0/total_returned 0.0
// shellies/shellies/shellyem3/emeter/1/power 0.00
// shellies/shellies/shellyem3/emeter/1/pf 0.00
// shellies/shellies/shellyem3/emeter/1/current 0.01
// shellies/shellies/shellyem3/emeter/1/voltage 0.12
// shellies/shellies/shellyem3/emeter/1/total 0.0
// shellies/shellies/shellyem3/emeter/1/total_returned 0.0
// shellies/shellies/shellyem3/emeter/2/power 0.00
// shellies/shellies/shellyem3/emeter/2/pf 0.00
// shellies/shellies/shellyem3/emeter/2/current 0.01
// shellies/shellies/shellyem3/emeter/2/voltage 0.09
// shellies/shellies/shellyem3/emeter/2/total 0.0
// shellies/shellies/shellyem3/emeter/2/total_returned 0.0
// shellies/shellies/shellyem3/relay/0 on
// shellies/shellies/shellyem3/relay/0 off

func shellyHandler(c Config, tm TopicMatcher, input Input, outputFunc OutputFunc) {
	// use our time
	timeStamp := time.Now()

	// parse topic
	device, err := tm.MatchDevice(input.Topic())
	if err != nil {
		log.Printf("tasmota-sensor[%s]: cannot extract device from topic='%s err=%s", c.Name(), input.Topic(), err)
		return
	}

	// parse payload
	var message tasmotaSensorMessage
	if err := json.Unmarshal(input.Payload(), &message); err != nil {
		log.Printf("tasmota-sensor[%s]: cannot json decode: %s", c.Name(), err)
		return
	}
}
