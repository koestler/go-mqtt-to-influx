package config

import (
	"testing"
)

const (
	InvalidSyntaxConfig = `- -`

	InvalidEmptyConfig = `
Version: 0
`

	InvalidUnknownVersionConfig = `
Version: 42
MqttClients:
  m0:
    Broker: "tcp://example.com:1883"

InfluxDbClients:
  i0:
    Address: http://172.17.0.2:8086

Converters:
  c0:
    Implementation: go-ve-sensor
    MqttTopics:
     - t0
`

	InvalidMqttIncompleteConfig = `
Version: 0
MqttClients:
  m0:
    Broker: "tcp://example.com:1883"

InfluxDbClients:
  i0:
    Address: http://172.17.0.2:8086

Converters:
  c0:
`

	ValidDefaultConfig = `
Version: 0
MqttClients:
  piegn-mosquitto:
    Broker: "tcp://example.com:1883"

InfluxDbClients:
  piegn:
    Address: http://172.17.0.2:8086

Converters:
  piegn-ve-sensor:
    Implementation: go-ve-sensor
    MqttTopics:
      - piegn/tele/+/SENSOR
`

	ValidComplexConfig = `
Version: 0
LogConfig: True
LogWorkerStart: True
LogMqttDebug: True
MqttClients:
  piegn-mosquitto:
    Broker: "tcp://example.com:1883"
    User: Bob
    Password: Jeir2Jie4zee
    TopicPrefix: piegn/
    LogMessages: False

  local-moquitto:
    Broker: "tcp://172.17.0.5:1883"
    TopicPrefix: wiedikon/
    LogMessages: True

InfluxDbClients:
  piegn:
    Address: http://172.17.0.2:8086
    User: Alice
    Password: An2iu2egheijeG
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
`
)

func TestReadConfig_InvalidSyntax(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidSyntaxConfig))
	if len(err) != 1 {
		t.Error("expected one error for invalid file")
	}
}

func TestReadConfig_Empty(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidEmptyConfig))
	if len(err) != 3 {
		t.Error("expected 3 errors; for empty MqttClients, empty InfluxDbClients, and empty Converters")
	}
}

func TestReadConfig_UnknownVersion(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidUnknownVersionConfig))
	if len(err) != 1 || err[0].Error() != "Version=42 not supported" {
		t.Errorf("expected 1 error: 'Version=42 not supported' but got: %v", err)
	}
}

func TestReadConfig_MqttIncompleteConfig(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidMqttIncompleteConfig))
	if len(err) != 2 ||
		err[0].Error() != "Converters->c0->Implementation='' is unkown" ||
		err[1].Error() != "Converters->c0->MqttTopics must not be empty" {

		t.Errorf("expected 2 errors for missing Implementation / MqttTopics but got: %v", err)
	}
}

func TestReadConfig_Complex(t *testing.T) {
	config, err := ReadConfig([]byte(ValidComplexConfig))
	if len(err) > 0 {
		t.Error("did not expect any errors")
	}

	t.Logf("config=%v", config)

	// General Section
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

	// MqttClients section
	if len(config.MqttClients) != 2 {
		t.Error("expect len(config.MqttClients) == 2")
	}

	if config.MqttClients[0].Name != "piegn-mosquitto" {
		t.Errorf("expected Name of first MqttClient to be 'piegn-mosquitto' but got '%s'",
			config.MqttClients[0].Name,
		)
	}

	if config.MqttClients[0].User != "Bob" {
		t.Error("expected User of first MqttClient to be 'Bob'")
	}

	if config.MqttClients[0].Password != "Jeir2Jie4zee" {
		t.Error("expected Password of first MqttClient to be 'Jeir2Jie4zee'")
	}

	if config.MqttClients[0].TopicPrefix != "piegn/" {
		t.Error("expected TopicPrefix of first MqttClient to be 'piegn/")
	}

	if config.MqttClients[0].LogMessages {
		t.Error("expected LogMessages of first MqttClient to be False")
	}

	if config.MqttClients[1].Name != "local-moquitto" {
		t.Errorf("expected Name of second MqttClient to be 'local-moquitto' but got '%s'",
			config.MqttClients[1].Name,
		)
	}

	if !config.MqttClients[1].LogMessages {
		t.Error("expected LogMessages of second MqttClient to be True")
	}

	// InfluxDbClients section
	if len(config.InfluxDbClients) != 2 {
		t.Error("expect len(config.InfluxDbClients) == 2")
	}

	if config.InfluxDbClients[0].Name != "piegn" {
		t.Errorf("expected Name of first InfluxDbClient to be 'piegn' but got '%s'",
			config.InfluxDbClients[0].Name,
		)
	}

	if config.InfluxDbClients[0].User != "Alice" {
		t.Error("expected User of first InfluxdbClient to be 'Alice'")
	}

	if config.InfluxDbClients[0].Password != "An2iu2egheijeG" {
		t.Error("expected Password of first InfluxDbClient to be 'An2iu2egheijeG'")
	}

	if config.InfluxDbClients[0].WriteInterval.String() != "200ms" {
		t.Error("expected WriteInterval of first InfluxDbClient to be '200ms'")
	}

	if config.InfluxDbClients[1].Name != "local" {
		t.Errorf("expected Name of first InfluxDbClient to be 'local' but got '%s'",
			config.InfluxDbClients[1].Name,
		)
	}


	// Converters section
	if len(config.Converters) != 4 {
		t.Errorf("expect len(config.Converters) == 4")
	}
}

func TestReadConfig_Default(t *testing.T) {
	config, err := ReadConfig([]byte(ValidDefaultConfig))
	if len(err) > 0 {
		t.Error("did not expect any errors")
	}

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
