package config

import (
	"bytes"
	"log"
	"os"
	"strings"
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
  m1:
    Broker: "tcp://example.com:1883"

InfluxClients:
  i0:
    Address: http://172.17.0.2:8086
  i1:
    Address: http://172.17.0.2:8086

Converters:
  c0:
    Implementation: go-ve-sensor
    MqttTopics:
     - t0
`

	InvalidMandatoryFieldsMissingConfig = `
Version: 0
MqttClients:
  m0:
InfluxClients:
  i0:
Converters:
  c0:
`
	InvalidValuesConfig = `
Version: 0
MqttClients:
  piegn mosquitto:
    Broker: "tcp://example.com:1883"
    Qos: 4

InfluxClients:
  piegn_foo:
    Address: http://172.17.0.2:8086
    WriteInterval: hello
    TimePrecision: x
  negative:
    Address: http://172.17.0.2:8086
    WriteInterval: -1s
    TimePrecision: -5ms

Statistics:
  Enabled: True
  HistoryResolution: 0s
  HistoryMaxAge: -1s

Converters:
  äöü:
    Implementation: go-ve-sensor
    MqttTopics:
      - x
    MqttClients:
      - inexistant-mqtt-client
    InfluxClients:
      - inexistant-influx-db-client
`

	ValidDefaultConfig = `
Version: 0
MqttClients:
  piegn-mosquitto:
    Broker: "tcp://example.com:1883"

InfluxClients:
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
HttpServer:
  Bind: 0.0.0.0
  Port: 80
  LogRequests: True
Statistics:
  Enabled: True
  HistoryResolution: 100ms
  HistoryMaxAge: 1h
  
MqttClients:
  0-piegn-mosquitto:
    Broker: "tcp://example.com:1883"
    User: Bob
    Password: Jeir2Jie4zee
    ClientId: "config-tester"
    Qos: 2
    AvailabilityTopic: test/%Prefix%tele/%clientId%/LWT
    TopicPrefix: piegn/
    LogMessages: False

  1-local-mosquitto:
    Broker: "tcp://172.17.0.5:1883"
    TopicPrefix: wiedikon/
    LogMessages: True

InfluxClients:
  0-piegn:
    Address: http://172.17.0.2:8086
    User: Alice
    Password: An2iu2egheijeG
    Database: test-database
    WriteInterval: 400ms
    TimePrecision: 1ms
    LogLineProtocol: True
  1-local:
    Address: http://172.17.0.4:8086
    WriteInterval: 0ms
    LogLineProtocol: False

Converters:
  0-piegn-ve-sensor:
    Implementation: go-ve-sensor
    TargetMeasurement: testfloatValue
    MqttTopics:
      - piegn/tele/ve/#
    MqttClients:
      - 0-piegn-mosquitto
      - 1-local-mosquitto
    InfluxClients:
      - 0-piegn
      - 1-local
    LogHandleOnce: True

  1-piegn-tasmota-lwt:
    Implementation: lwt
    TargetMeasurement: boolValue
    MqttTopics:
      - piegn/tele/+/LWT
      - piegn/tele/+/+/LWT
      - piegn/tele/+/+/+/LWT

  2-piegn-tasmota-state:
    Implementation: tasmota-state
    TargetMeasurement: tasmotaState
    MqttTopics:
      - piegn/tele/+/STATE
      - piegn/tele/+/+/STATE
      - piegn/tele/+/+/+/STATE

  3-piegn-tasmota-sensor:
    Implementation: tasmota-sensor
    TargetMeasurement: floatValue
    MqttTopics:
      - piegn/tele/+/SENSOR
      - piegn/tele/+/+/SENSOR
      - piegn/tele/+/+/+/SENSOR
`
)

func containsError(needle string, err []error) bool {
	for _, e := range err {
		if strings.Contains(e.Error(), needle) {
			return true
		}
	}
	return false
}

func TestReadConfig_InvalidSyntax(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidSyntaxConfig))
	if len(err) != 1 {
		t.Error("expect one error for invalid file")
	}
}

func TestReadConfig_NoVersion(t *testing.T) {
	_, err := ReadConfig([]byte(""))

	if !containsError("Version must be defined", err) {
		t.Error("no version given; expect 'Version must be defined'")
	}
}

func TestReadConfig_InvalidEmpty(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidEmptyConfig))
	if len(err) != 3 {
		t.Error("expect 3 errors; for empty MqttClients, empty InfluxClients, and empty Converters")
	}
}

func TestReadConfig_InvalidUnknownVersion(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidUnknownVersionConfig))
	if len(err) != 1 || err[0].Error() != "Version=42 is not supported." {
		t.Errorf("expect 1 error: 'Version=42 is not supported.' but got: %v", err)
	}
}

func TestReadConfig_InvalidMandatoryFieldsMissing(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidMandatoryFieldsMissingConfig))

	t.Logf("InvalidMandatoryFieldsMissingConfig returned err=%v", err)

	if !containsError("Broker", err) {
		t.Error("expect an error for missing Broker")
	}

	if !containsError("Address", err) {
		t.Error("expect an error for missing Address")
	}

	if !containsError("Implementation", err) {
		t.Error("expect an error for missing Implementation")
	}

	if !containsError("MqttTopics", err) {
		t.Error("expect an error for missing MqttTopics")
	}
}

func TestReadConfig_InvalidValues(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidValuesConfig))

	t.Logf("InvalidValuesConfig returned err=%v", err)

	for _, name := range []string{"piegn mosquitto", "piegn_foo", "äöü"} {
		if !containsError(name, err) {
			t.Errorf("expect invalid name '%s' to be returned as error", name)
		}
	}

	for _, name := range []string{"inexistant-mqtt-client", "inexistant-influx-db-client"} {
		if !containsError(name, err) {
			t.Errorf("expect inexistant name reference '%s' to be returned as error", name)
		}
	}

	if !containsError("Qos", err) {
		t.Error("expect invalid Qos value of 4 to be returned as error")
	}

	if !containsError("hello", err) {
		t.Error("expect invalid writeInterval='hello' to be returned as error")
	}

	if !containsError("-1s", err) {
		t.Error("expect invalid writeInterval='-1s' to be returned as error")
	}

	if !containsError("HistoryResolution", err) {
		t.Error("expect invalid HistoryResolution='0s' to be returned as error")
	}

	if !containsError("HistoryMaxAge", err) {
		t.Error("expect invalid HistoryMaxAge='-1s' to be returned as error")
	}

}

// check that a complex example setting all available options is correctly read
func TestReadConfig_Complex(t *testing.T) {
	config, err := ReadConfig([]byte(ValidComplexConfig))
	if len(err) > 0 {
		t.Errorf("did not expect any errors, got: %v", err)
	}

	t.Logf("config=%v", config)

	// General Section
	if config.Version != 0 {
		t.Error("expect Version = 0")
	}

	if !config.LogConfig {
		t.Errorf("expect LogConfig to be True as configured")
	}

	if !config.LogWorkerStart {
		t.Errorf("expect LogWorkerStart to be True as configured")
	}

	if !config.LogMqttDebug {
		t.Errorf("expect LogMqttDebug to be True as configured")
	}

	// mqttClients section
	if len(config.MqttClients) != 2 {
		t.Error("expect len(config.MqttClients) == 2")
	}

	if config.MqttClients[0].Name() != "0-piegn-mosquitto" {
		t.Errorf("expect Name of first MqttClient to be '0-piegn-mosquitto' but got '%s'",
			config.MqttClients[0].Name(),
		)
	}

	if config.MqttClients[1].Name() != "1-local-mosquitto" {
		t.Errorf("expect Name of second MqttClient to be '1-local-mosquitto' but got '%s'",
			config.MqttClients[1].Name(),
		)
	}

	if config.MqttClients[0].Broker() != "tcp://example.com:1883" {
		t.Error("expect Broker of first MqttClient to be 'tcp://example.com:1883'")
	}

	if config.MqttClients[0].User() != "Bob" {
		t.Error("expect User of first MqttClient to be 'Bob'")
	}

	if config.MqttClients[0].Password() != "Jeir2Jie4zee" {
		t.Error("expect Password of first MqttClient to be 'Jeir2Jie4zee'")
	}

	if config.MqttClients[0].ClientId() != "config-tester" {
		t.Error("expect Password of first MqttClient to be 'config-tester'")
	}

	if config.MqttClients[0].Qos() != 2 {
		t.Error("expect Qos of first MqttClient to be 2")
	}

	expectedTopic := "test/%Prefix%tele/%clientId%/LWT"
	if config.MqttClients[0].AvailabilityTopic() != expectedTopic {
		t.Errorf("expect AvailabilityTopic of first MqttClient to be '%s'", expectedTopic)
	}

	if config.MqttClients[0].TopicPrefix() != "piegn/" {
		t.Error("expect TopicPrefix of first MqttClient to be 'piegn/'")
	}

	if config.MqttClients[0].LogMessages() {
		t.Error("expect LogMessages of first MqttClient to be False")
	}

	if !config.MqttClients[1].LogMessages() {
		t.Error("expect LogMessages of second MqttClient to be True")
	}

	// influxClients section
	if len(config.InfluxClients) != 2 {
		t.Error("expect len(config.InfluxClients) == 2")
	}

	if config.InfluxClients[0].Name() != "0-piegn" {
		t.Errorf("expect Name of first InfluxClient to be '0-piegn' but got '%s'",
			config.InfluxClients[0].Name(),
		)
	}

	if config.InfluxClients[1].Name() != "1-local" {
		t.Errorf("expect Name of first InfluxClient to be '1-local' but got '%s'",
			config.InfluxClients[1].Name(),
		)
	}

	if config.InfluxClients[0].Address() != "http://172.17.0.2:8086" {
		t.Error("expect Address of first InfluxClient to be 'http://172.17.0.2:8086'")
	}

	if config.InfluxClients[0].User() != "Alice" {
		t.Error("expect User of first InfluxClient to be 'Alice'")
	}

	if config.InfluxClients[0].Password() != "An2iu2egheijeG" {
		t.Error("expect Password of first InfluxClient to be 'An2iu2egheijeG'")
	}

	if config.InfluxClients[0].Database() != "test-database" {
		t.Error("expect Database of first InfluxClient to be 'test-database'")
	}

	if config.InfluxClients[0].WriteInterval().String() != "400ms" {
		t.Error("expect WriteInterval of first InfluxClient to be '400ms'")
	}

	if config.InfluxClients[0].TimePrecision().String() != "1ms" {
		t.Error("expect TimePrecision of first InfluxClient to be '1ms'")
	}

	if !config.InfluxClients[0].LogLineProtocol() {
		t.Error("expect LogLineProtocol of first InfluxClient to be True")
	}

	// Converters section
	if len(config.Converters) != 4 {
		t.Errorf("expect len(config.Converters) == 4")
	}

	if config.Converters[0].Name() != "0-piegn-ve-sensor" {
		t.Errorf("expect Name of first Converter to be '0-piegn-ve-sensor' but got '%s'",
			config.Converters[0].Name(),
		)
	}

	if config.Converters[0].Implementation() != "go-ve-sensor" {
		t.Error("expect Implementation of first Converter to be 'go-ve-sensor'")
	}

	if config.Converters[0].TargetMeasurement() != "testfloatValue" {
		t.Error("expect TargetMeasurement of first Converter to be 'testfloatValue'")
	}

	if len(config.Converters[0].MqttTopics()) != 1 || config.Converters[0].MqttTopics()[0] != "piegn/tele/ve/#" {
		t.Errorf("expect MqttTopics of first Converter to be ['piegn/tele/ve/#'] got %v",
			config.Converters[0].MqttTopics(),
		)
	}

	if len(config.Converters[0].MqttClients()) != 2 ||
		config.Converters[0].MqttClients()[0] != "0-piegn-mosquitto" ||
		config.Converters[0].MqttClients()[1] != "1-local-mosquitto" {
		t.Errorf("expect MqttClients of first Converter to be ['0-piegn-mosquitto', '1-local-mosquitto'] got %v",
			config.Converters[0].MqttClients(),
		)
	}

	if len(config.Converters[0].InfluxClients()) != 2 ||
		config.Converters[0].InfluxClients()[0] != "0-piegn" ||
		config.Converters[0].InfluxClients()[1] != "1-local" {
		t.Errorf("expect InfluxClients of first Converter to be ['0-piegn', '1-local'] got %v",
			config.Converters[0].InfluxClients(),
		)
	}

	if !config.Converters[0].LogHandleOnce() {
		t.Error("expect LogHandleOnce of first Converter to be True")
	}

	// HttpServer
	if !config.HttpServer.Enabled() {
		t.Error("expect HttpServer->Enabled to be True")
	}

	if config.HttpServer.Bind() != "0.0.0.0" {
		t.Error("expect HttpServer->Bind to be '0.0.0.0'")
	}

	if config.HttpServer.Port() != 80 {
		t.Error("expect HttpServer->Port to be 80")
	}

	if !config.HttpServer.LogRequests() {
		t.Error("expect HttpServer->LogRequests to be True")
	}

	// Statistics
	if !config.Statistics.Enabled() {
		t.Error("expect Statistics->Enabled to be True")
	}

	if config.Statistics.HistoryResolution().String() != "100ms" {
		t.Errorf("expect Statistics->HistoryResolution to be '100ms', got '%s'",
			config.Statistics.HistoryResolution())
	}

	if config.Statistics.HistoryMaxAge().String() != "1h0m0s" {
		t.Errorf("expect Statistics->HistoryMaxAge to be '1h0m0s', got '%s'",
			config.Statistics.HistoryMaxAge())
	}

	// test config output does not crash
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	if err := config.PrintConfig(); err != nil {
		t.Errorf("expected no error. Got: %s", err)
	}
	t.Log(buf.String())
}

// check that default values are set as documented in types.go
func TestReadConfig_Default(t *testing.T) {
	config, err := ReadConfig([]byte(ValidDefaultConfig))
	if len(err) > 0 {
		t.Errorf("did not expect any errors, got %v", err)
	}

	// General Section
	if config.LogConfig {
		t.Error("expect LogConfig to be False by default")
	}

	if config.LogWorkerStart {
		t.Error("expect LogWorkerStart to be False by default")
	}

	if config.LogMqttDebug {
		t.Error("expect LogMqttDebug to be False by default")
	}

	// influxClients section
	if config.MqttClients[0].User() != "" {
		t.Error("expect default MqttClient->User to be empty")
	}

	if config.MqttClients[0].Password() != "" {
		t.Error("expect default MqttClient->Password to be empty")
	}

	if config.MqttClients[0].ClientId() != "go-mqtt-to-influx" {
		t.Error("expect default MqttClient->ClientId to be 'go-mqtt-to-influx'")
	}

	if config.MqttClients[0].Qos() != 0 {
		t.Error("expect default MqttClient->Qos to be 0")
	}

	expectedAvailabilityTopic := "%Prefix%tele/%clientId%/LWT"
	if config.MqttClients[0].AvailabilityTopic() != expectedAvailabilityTopic {
		t.Errorf("expect default MqttClient->AvailabilityTopic to be '%s'", expectedAvailabilityTopic)
	}

	if config.MqttClients[0].TopicPrefix() != "" {
		t.Error("expect default MqttClient->TopicPrefix to be empty")
	}

	if config.MqttClients[0].LogMessages() {
		t.Error("expect default MqttClient->LogMessages to be False")
	}

	// influxClients section
	if config.InfluxClients[0].User() != "" {
		t.Error("expect default InfluxClient->User to be empty")
	}

	if config.InfluxClients[0].Password() != "" {
		t.Error("expect default InfluxClient->Password to be empty")
	}

	if config.InfluxClients[0].Database() != "go-mqtt-to-influx" {
		t.Error("expect default InfluxClient->Database to be 'go-mqtt-to-influx'")
	}

	if config.InfluxClients[0].WriteInterval().String() != "200ms" {
		t.Error("expect default InfluxClient->WriteInterval to be 200ms")
	}

	if config.InfluxClients[0].TimePrecision().String() != "1s" {
		t.Error("expect default InfluxClient->TimePrecision to be 1s")
	}

	if config.InfluxClients[0].LogLineProtocol() {
		t.Error("expect default InfluxClient->LogLineProtocol to be False")
	}

	// Converters section
	if config.Converters[0].TargetMeasurement() != "floatValue" {
		t.Error("expect default Converter->TargetMeasurement to be 'floatValue'")
	}

	if len(config.Converters[0].MqttClients()) != 0 {
		t.Error("expect default Converter->MqttClients to be empty")
	}

	if len(config.Converters[0].InfluxClients()) != 0 {
		t.Error("expect default Converter->InfluxClients to be empty")
	}

	// HttpServer
	if config.HttpServer.Enabled() {
		t.Error("expect default HttpServer->Enabled to be False")
	}

	if config.HttpServer.Bind() != "[::1]" {
		t.Error("expect default HttpServer->Bind to be '[::1]'")
	}

	if config.HttpServer.Port() != 8042 {
		t.Error("expect default HttpServer->Port to be 8042")
	}

	if config.HttpServer.LogRequests() {
		t.Error("expect default HttpServer->LogRequests to be False")
	}

	// Statistics
	if config.Statistics.Enabled() {
		t.Error("expect default Statistics->Enabled to be False")
	}

	if config.Statistics.HistoryResolution().String() != "1s" {
		t.Errorf("expect default Statistics->HistoryResolution to be '1s', got '%s'",
			config.Statistics.HistoryResolution())
	}

	if config.Statistics.HistoryMaxAge().String() != "10m0s" {
		t.Errorf("expect default Statistics->HistoryMaxAge to be '10m0s', got '%s'",
			config.Statistics.HistoryMaxAge())
	}
}
