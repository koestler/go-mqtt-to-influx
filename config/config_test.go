package config

import (
	"testing"
)

func TestReadConfig(t *testing.T) {
	textConfigs := []string{`
Version: 0
LogConfig: True
LogWorkerStart: True
LogMqttDebug: True
MqttClients:
  piegn-mosquitto:
    Broker: "tcp://example.com:1883"
    User: go-mqtt-to-influxdb
    Password: Jeir2Jie4zee
    TopicPrefix: piegn/
    LogConfig: True
    LogWorkerStart: True
    LogMqttDebug: True

  local-moquitto:
    Broker: "tcp://172.17.0.5:1883"
    TopicPrefix: wiedikon/
    LogMessages: True

InfluxDbClients:
  piegn:
    Address: http://172.17.0.2:8086
    WriteInterval: 200ms
    LogLineProtocol: False
    LogMessages: True
  local:
    Address: http://172.17.0.4:8086
    WriteInterval: 0ms
    LogLineProtocol: False
    LogMessages: False

Converters:
  piegn-ve-sensor:
    Implementation: go-ve-sensor
    TargetMeasurement: floatValue
    LogHandleOnce: True
    MqttTopics:
      - piegn/tele/ve/#

  piegn-tasmota-lwt:
    Implementation: lwt
    TargetMeasurement: boolValue
    MqttTopics:
      - piegn/tele/+/LWT
      - piegn/tele/+/+/LWT
      - piegn/tele/+/+/+/LWT

  piegn-tasmota-state:
    Implementation: tasmota-state
    TargetMeasurement: tasmotaState
    MqttTopics:
      - piegn/tele/+/STATE
      - piegn/tele/+/+/STATE
      - piegn/tele/+/+/+/STATE

  piegn-tasmota-sensor:
    Implementation: tasmota-sensor
    TargetMeasurement: floatValue
    MqttTopics:
      - piegn/tele/+/SENSOR
      - piegn/tele/+/+/SENSOR
      - piegn/tele/+/+/+/SENSOR
`, `
Version: 0
MqttClients:
  piegn-mosquitto:
    Broker: "tcp://example.com:1883"
    User: go-mqtt-to-influxdb

InfluxDbClients:
  piegn:
    Address: http://172.17.0.2:8086

Converters:
  piegn-ve-sensor:
    Implementation: go-ve-sensor
    MqttTopics:
      - piegn/tele/+/SENSOR
`}

	config := ReadConfig("test", []byte(textConfigs[0]))

	if config.Version != 0 {
		t.Error("expect Version = 0")
	}

	if !config.LogConfig {
		t.Errorf("expect config.LogConfig to be True as configured")
	}

	if !config.LogWorkerStart {
		t.Errorf("expect config.LogWorkerStart to be True as configured")
	}

	if !config.LogMqttDebug {
		t.Errorf("expect config.LogMqttDebug to be True as configured")
	}

	if len(config.MqttClients) != 2 {
		t.Error("expect len(config.MqttClients) == 2")
	}

	if len(config.InfluxDbClients) != 2 {
		t.Error("expect len(config.InfluxDbClients) == 2")
	}

	if len(config.Converters) != 4 {
		t.Errorf("expect len(config.Converters) == 4")
	}

	config = ReadConfig("test", []byte(textConfigs[1]))
	// check default avlues

	if config.LogConfig {
		t.Errorf("expect config.LogConfig to be False by default")
	}

	if config.LogWorkerStart {
		t.Errorf("expect config.LogWorkerStart to be False by default")
	}

	if config.LogMqttDebug {
		t.Errorf("expect config.LogMqttDebug to be False by default")
	}

	// todo: trigger errors
}
