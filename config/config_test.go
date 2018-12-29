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
mqttClients:
  m0:
    broker: "tcp://example.com:1883"
  m1:
    broker: "tcp://example.com:1883"

influxDbClients:
  i0:
    address: http://172.17.0.2:8086
  i1:
    address: http://172.17.0.2:8086

Converters:
  c0:
    implementation: go-ve-sensor
    mqttTopics:
     - t0
`

	InvalidMandatoryFieldsMissingConfig = `
Version: 0
mqttClients:
  m0:
influxDbClients:
  i0:
Converters:
  c0:
`
	InvalidValuesConfig = `
Version: 0
mqttClients:
  piegn mosquitto:
    broker: "tcp://example.com:1883"
    qos: 4

influxDbClients:
  piegn_foo:
    address: http://172.17.0.2:8086
    writeInterval: hello
    timePrecision: x
  negative:
    address: http://172.17.0.2:8086
    writeInterval: -1s
    timePrecision: -5ms

Converters:
  äöü:
    implementation: go-ve-sensor
    mqttTopics:
      - x
    mqttClients:
      - inexistant-mqtt-client
    influxDbClients:
      - inexistant-influx-db-client
`

	ValidDefaultConfig = `
Version: 0
mqttClients:
  piegn-mosquitto:
    broker: "tcp://example.com:1883"

influxDbClients:
  piegn:
    address: http://172.17.0.2:8086

Converters:
  piegn-ve-sensor:
    implementation: go-ve-sensor
    mqttTopics:
      - piegn/tele/+/SENSOR
`

	ValidComplexConfig = `
Version: 0
LogConfig: True
LogWorkerStart: True
LogMqttDebug: True
HttpServer:
  bind: 0.0.0.0
  Port: 80
  logRequests: True
Statistics:
  enabled: True
mqttClients:
  0-piegn-mosquitto:
    broker: "tcp://example.com:1883"
    user: Bob
    password: Jeir2Jie4zee
    clientId: "config-tester"
    qos: 2
    availabilityTopic: test/%Prefix%tele/%clientId%/LWT
    topicPrefix: piegn/
    logMessages: False

  1-local-mosquitto:
    broker: "tcp://172.17.0.5:1883"
    topicPrefix: wiedikon/
    logMessages: True

influxDbClients:
  0-piegn:
    address: http://172.17.0.2:8086
    user: Alice
    password: An2iu2egheijeG
    database: test-database
    writeInterval: 400ms
    timePrecision: 1ms
    logLineProtocol: True
  1-local:
    address: http://172.17.0.4:8086
    writeInterval: 0ms
    logLineProtocol: False

Converters:
  0-piegn-ve-sensor:
    implementation: go-ve-sensor
    targetMeasurement: testfloatValue
    mqttTopics:
      - piegn/tele/ve/#
    mqttClients:
      - 0-piegn-mosquitto
      - 1-local-mosquitto
    influxDbClients:
      - 0-piegn
      - 1-local
    logHandleOnce: True

  1-piegn-tasmota-lwt:
    implementation: lwt
    targetMeasurement: boolValue
    mqttTopics:
      - piegn/tele/+/LWT
      - piegn/tele/+/+/LWT
      - piegn/tele/+/+/+/LWT

  2-piegn-tasmota-state:
    implementation: tasmota-state
    targetMeasurement: tasmotaState
    mqttTopics:
      - piegn/tele/+/STATE
      - piegn/tele/+/+/STATE
      - piegn/tele/+/+/+/STATE

  3-piegn-tasmota-sensor:
    implementation: tasmota-sensor
    targetMeasurement: floatValue
    mqttTopics:
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
		t.Error("expect 3 errors; for empty mqttClients, empty influxDbClients, and empty Converters")
	}
}

func TestReadConfig_InvalidUnknownVersion(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidUnknownVersionConfig))
	if len(err) != 1 || err[0].Error() != "Version=42 not supported" {
		t.Errorf("expect 1 error: 'Version=42 not supported' but got: %v", err)
	}
}

func TestReadConfig_InvalidMandatoryFieldsMissing(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidMandatoryFieldsMissingConfig))

	t.Logf("InvalidMandatoryFieldsMissingConfig returned err=%v", err)

	if !containsError("broker", err) {
		t.Error("expect an error for missing broker")
	}

	if !containsError("address", err) {
		t.Error("expect an error for missing address")
	}

	if !containsError("implementation", err) {
		t.Error("expect an error for missing implementation")
	}

	if !containsError("mqttTopics", err) {
		t.Error("expect an error for missing mqttTopics")
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

	if !containsError("qos", err) {
		t.Error("expect invalid qos value of 4 to be returned as error")
	}

	if !containsError("hello", err) {
		t.Error("expect invalid writeInterval='hello' to be returned as error")
	}

	if !containsError("-1s", err) {
		t.Error("expect invalid writeInterval='-1s' to be returned as error")
	}

	if !containsError("-5ms", err) {
		t.Error("expect invalid timePrecision='-5ms' to be returned as error")
	}
}

// check that a complex example setting all available options is correctly read
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
		t.Error("expect len(config.mqttClients) == 2")
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
		t.Error("expect broker of first MqttClient to be 'tcp://example.com:1883'")
	}

	if config.MqttClients[0].User() != "Bob" {
		t.Error("expect user of first MqttClient to be 'Bob'")
	}

	if config.MqttClients[0].Password() != "Jeir2Jie4zee" {
		t.Error("expect password of first MqttClient to be 'Jeir2Jie4zee'")
	}

	if config.MqttClients[0].ClientId() != "config-tester" {
		t.Error("expect password of first MqttClient to be 'config-tester'")
	}

	if config.MqttClients[0].Qos() != 2 {
		t.Error("expect qos of first MqttClient to be 2")
	}

	expectedTopic := "test/%Prefix%tele/%clientId%/LWT"
	if config.MqttClients[0].AvailabilityTopic() != expectedTopic {
		t.Errorf("expect availabilityTopic of first MqttClient to be '%s'", expectedTopic)
	}

	if config.MqttClients[0].TopicPrefix() != "piegn/" {
		t.Error("expect topicPrefix of first MqttClient to be 'piegn/'")
	}

	if config.MqttClients[0].LogMessages() {
		t.Error("expect logMessages of first MqttClient to be False")
	}

	if !config.MqttClients[1].LogMessages() {
		t.Error("expect logMessages of second MqttClient to be True")
	}

	// influxDbClients section
	if len(config.InfluxDbClients) != 2 {
		t.Error("expect len(config.influxDbClients) == 2")
	}

	if config.InfluxDbClients[0].Name() != "0-piegn" {
		t.Errorf("expect Name of first InfluxDbClient to be '0-piegn' but got '%s'",
			config.InfluxDbClients[0].Name(),
		)
	}

	if config.InfluxDbClients[1].Name() != "1-local" {
		t.Errorf("expect Name of first InfluxDbClient to be '1-local' but got '%s'",
			config.InfluxDbClients[1].Name(),
		)
	}

	if config.InfluxDbClients[0].Address() != "http://172.17.0.2:8086" {
		t.Error("expect address of first InfluxdbClient to be 'http://172.17.0.2:8086'")
	}

	if config.InfluxDbClients[0].User() != "Alice" {
		t.Error("expect user of first InfluxdbClient to be 'Alice'")
	}

	if config.InfluxDbClients[0].Password() != "An2iu2egheijeG" {
		t.Error("expect password of first InfluxDbClient to be 'An2iu2egheijeG'")
	}

	if config.InfluxDbClients[0].Database() != "test-database" {
		t.Error("expect database of first InfluxDbClient to be 'test-database'")
	}

	if config.InfluxDbClients[0].WriteInterval().String() != "400ms" {
		t.Error("expect writeInterval of first InfluxDbClient to be '400ms'")
	}

	if config.InfluxDbClients[0].TimePrecision().String() != "1ms" {
		t.Error("expect timePrecision of first InfluxDbClient to be '1ms'")
	}

	if !config.InfluxDbClients[0].LogLineProtocol() {
		t.Error("expect logLineProtocol of first InfluxDbClient to be True")
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
		t.Error("expect implementation of first Converter to be 'go-ve-sensor'")
	}

	if config.Converters[0].TargetMeasurement() != "testfloatValue" {
		t.Error("expect targetMeasurement of first Converter to be 'testfloatValue'")
	}

	if len(config.Converters[0].MqttTopics()) != 1 || config.Converters[0].MqttTopics()[0] != "piegn/tele/ve/#" {
		t.Errorf("expect mqttTopics of first Converter to be ['piegn/tele/ve/#'] got %v",
			config.Converters[0].MqttTopics(),
		)
	}

	if len(config.Converters[0].MqttClients()) != 2 ||
		config.Converters[0].MqttClients()[0] != "0-piegn-mosquitto" ||
		config.Converters[0].MqttClients()[1] != "1-local-mosquitto" {
		t.Errorf("expect mqttClients of first Converter to be ['0-piegn-mosquitto', '1-local-mosquitto'] got %v",
			config.Converters[0].MqttClients(),
		)
	}

	if len(config.Converters[0].InfluxDbClients()) != 2 ||
		config.Converters[0].InfluxDbClients()[0] != "0-piegn" ||
		config.Converters[0].InfluxDbClients()[1] != "1-local" {
		t.Errorf("expect influxDbClients of first Converter to be ['0-piegn', '1-local'] got %v",
			config.Converters[0].InfluxDbClients(),
		)
	}

	if !config.Converters[0].LogHandleOnce() {
		t.Error("expect logHandleOnce of first Converter to be True")
	}

	// HttpServer
	if !config.HttpServer.Enabled() {
		t.Error("expect HttpServer->enabled to be True")
	}

	if config.HttpServer.Bind() != "0.0.0.0" {
		t.Error("expect HttpServer->bind to be '0.0.0.0'")
	}

	if config.HttpServer.Port() != 80 {
		t.Error("expect HttpServer->Port to be 80")
	}

	if !config.HttpServer.LogRequests() {
		t.Error("expect HttpServer->logRequests to be True")
	}

	// Statistics
	if !config.Statistics.Enabled() {
		t.Error("expect Statistics->enabled to be True")
	}

	// test config output does not crash
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	config.PrintConfig()
	t.Log(buf.String())
}

// check that default values are set as documented in types.go
func TestReadConfig_Default(t *testing.T) {
	config, err := ReadConfig([]byte(ValidDefaultConfig))
	if len(err) > 0 {
		t.Error("did not expect any errors")
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

	// influxDbClients section
	if config.MqttClients[0].User() != "" {
		t.Error("expect default MqttClient->user to be empty")
	}

	if config.MqttClients[0].Password() != "" {
		t.Error("expect default MqttClient->password to be empty")
	}

	if config.MqttClients[0].ClientId() != "go-mqtt-to-influxdb" {
		t.Error("expect default MqttClient->clientId to be 'go-mqtt-to-influxdb'")
	}

	if config.MqttClients[0].Qos() != 0 {
		t.Error("expect default MqttClient->qos to be 0")
	}

	expectedAvailabilityTopic := "%Prefix%tele/%clientId%/LWT"
	if config.MqttClients[0].AvailabilityTopic() != expectedAvailabilityTopic {
		t.Errorf("expect default MqttClient->availabilityTopic to be '%s'", expectedAvailabilityTopic)
	}

	if config.MqttClients[0].TopicPrefix() != "" {
		t.Error("expect default MqttClient->topicPrefix to be empty")
	}

	if config.MqttClients[0].LogMessages() {
		t.Error("expect default MqttClient->logMessages to be False")
	}

	// influxDbClients section
	if config.InfluxDbClients[0].User() != "" {
		t.Error("expect default InfluxDbClient->user to be empty")
	}

	if config.InfluxDbClients[0].Password() != "" {
		t.Error("expect default InfluxDbClient->password to be empty")
	}

	if config.InfluxDbClients[0].Database() != "go-mqtt-to-influxdb" {
		t.Error("expect default InfluxDbClient->database to be 'go-mqtt-to-influxdb'")
	}

	if config.InfluxDbClients[0].WriteInterval().String() != "200ms" {
		t.Error("expect default InfluxDbClient->writeInterval to be 200ms")
	}

	if config.InfluxDbClients[0].TimePrecision().String() != "1s" {
		t.Error("expect default InfluxDbClient->timePrecision to be 1s")
	}

	if config.InfluxDbClients[0].LogLineProtocol() {
		t.Error("expect default InfluxDbClient->logLineProtocol to be False")
	}

	// Converters section
	if config.Converters[0].TargetMeasurement() != "floatValue" {
		t.Error("expect default Converter->targetMeasurement to be 'floatValue'")
	}

	if len(config.Converters[0].MqttClients()) != 0 {
		t.Error("expect default Converter->mqttClients to be empty")
	}

	if len(config.Converters[0].InfluxDbClients()) != 0 {
		t.Error("expect default Converter->influxDbClients to be empty")
	}

	// HttpServer
	if config.HttpServer.Enabled() {
		t.Error("expect default HttpServer->enabled to be False")
	}

	if config.HttpServer.Bind() != "[::1]" {
		t.Error("expect default HttpServer->bind to be '[::1]'")
	}

	if config.HttpServer.Port() != 8042 {
		t.Error("expect default HttpServer->Port to be 8042")
	}

	if config.HttpServer.LogRequests() {
		t.Error("expect default HttpServer->logRequests to be False")
	}

	// Statistics
	if config.Statistics.Enabled() {
		t.Error("expect default Statistics->enabled to be False")
	}
}
