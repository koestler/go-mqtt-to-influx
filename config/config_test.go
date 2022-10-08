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
    Url: http://172.17.0.2:8086
    Token: foobar
    Org: myorg
    Bucket: mybucket

Converters:
  c0:
    Implementation: go-iotdevice
    MqttTopics:
     - Topic: t0/%Device%
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
    Implementation: go-iotdevice
    MqttTopics:
      - Topic: x/%Device%
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
    Url: http://172.17.0.2:8086
    Token: "foobar-token"
    Org: Piegn
    Bucket: iot

Converters:
  piegn-ve-sensor:
    Implementation: go-iotdevice
    MqttTopics:
      - Topic: piegn/tele/%Device%/SENSOR
`

	ValidComplexConfig = `
Version: 0
LogConfig: True
LogWorkerStart: True
HttpServer:
  Bind: 0.0.0.0
  Port: 80
  LogRequests: True

LocalDb:
  Enabled: True
  Path: /tmp/foobar.db

Statistics:
  Enabled: True
  HistoryResolution: 100ms
  HistoryMaxAge: 1h
  
MqttClients:
  0-piegn-mosquitto:
    Broker: "tcp://example.com:1883"
    ProtocolVersion: 5
    User: Bob
    Password: Jeir2Jie4zee
    ClientId: "config-tester"
    Qos: 2
    KeepAlive: 7s
    ConnectRetryDelay: 8s 
    ConnectTimeout: 9s
    AvailabilityTopic: test/%Prefix%tele/%ClientId%/LWT
    TopicPrefix: piegn/
    LogMessages: False
    LogDebug: True

  1-local-mosquitto:
    Broker: "tcp://172.17.0.5:1883"
    TopicPrefix: wiedikon/
    LogMessages: True
    LogDebug: False

InfluxClients:
  0-piegn:
    Url: http://172.17.0.2:8086
    Token: mytoken
    Org: myorg
    Bucket: mybucket
    WriteInterval: 400ms
    TimePrecision: 1ms
    LogDebug: True
  1-local:
    Url: http://172.17.0.4:8086
    Token: mytoken
    Org: myorg
    Bucket: mybucket
    LogDebug: False

InfluxAuxiliaryTags:
  - Tag: device
    Equals: foo
    TagValues:
      a: foo
      b: bar
      c: "another String"
  - Tag: field
    Matches: ^temperature[0-9]+
    TagValues:
      sort: sensor

Converters:
  0-piegn-ve-sensor:
    Implementation: go-iotdevice
    MqttTopics:
      - Topic: piegn/tele/ve/#
        Device: fixed-ve
    MqttClients:
      - 0-piegn-mosquitto
      - 1-local-mosquitto
    InfluxClients:
      - 0-piegn
      - 1-local
    LogHandleOnce: True

  1-piegn-tasmota-availability:
    Implementation: availability
    MqttTopics:
      - Topic: piegn/tele/%Device%/LWT
        Device: +
      - Topic: piegn/tele/%Device%/LWT
        Device: +/+
      - Topic: piegn/tele/%Device%/LWT
        Device: +/+/+

  2-piegn-tasmota-state:
    Implementation: tasmota-state
    MqttTopics:
      - Topic: piegn/tele/%Device%/STATE
        Device: +
      - Topic: piegn/tele/%Device%/STATE
        Device: +/+
      - Topic: piegn/tele/%Device%/STATE
        Device: +/+/+

  3-piegn-tasmota-sensor:
    Implementation: tasmota-sensor
    MqttTopics:
      - Topic: piegn/tele/foobar/SENSOR
        Device: fixed-device
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

	if !containsError("Url", err) {
		t.Error("expect an error for missing Url")
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
	if config.Version() != 0 {
		t.Error("expect Version = 0")
	}

	if !config.LogConfig() {
		t.Errorf("expect LogConfig to be True as configured")
	}

	if !config.LogWorkerStart() {
		t.Errorf("expect LogWorkerStart to be True as configured")
	}

	// mqttClients section
	if len(config.MqttClients()) != 2 {
		t.Error("expect len(config.MqttClients) == 2")
	}

	if config.MqttClients()[0].Name() != "0-piegn-mosquitto" {
		t.Errorf("expect Name of first MqttClient to be '0-piegn-mosquitto' but got '%s'",
			config.MqttClients()[0].Name(),
		)
	}

	if config.MqttClients()[1].Name() != "1-local-mosquitto" {
		t.Errorf("expect Name of second MqttClient to be '1-local-mosquitto' but got '%s'",
			config.MqttClients()[1].Name(),
		)
	}

	if config.MqttClients()[0].Broker().String() != "tcp://example.com:1883" {
		t.Error("expect Broker of first MqttClient to be 'tcp://example.com:1883'")
	}

	if config.MqttClients()[0].ProtocolVersion() != 5 {
		t.Error("expect ProtocolVersion of first MqttClient to be 5")
	}

	if config.MqttClients()[0].User() != "Bob" {
		t.Error("expect User of first MqttClient to be 'Bob'")
	}

	if config.MqttClients()[0].Password() != "Jeir2Jie4zee" {
		t.Error("expect Password of first MqttClient to be 'Jeir2Jie4zee'")
	}

	if config.MqttClients()[0].ClientId() != "config-tester" {
		t.Error("expect Password of first MqttClient to be 'config-tester'")
	}

	if config.MqttClients()[0].Qos() != 2 {
		t.Error("expect Qos of first MqttClient to be 2")
	}

	if config.MqttClients()[0].KeepAlive().String() != "7s" {
		t.Error("expect KeepAlive of first MqttClient to be 7s")
	}

	if config.MqttClients()[0].ConnectRetryDelay().String() != "8s" {
		t.Error("expect ConnectRetryDelay of first MqttClient to be 8s")
	}

	if config.MqttClients()[0].ConnectTimeout().String() != "9s" {
		t.Error("expect ConnectTimeout of first MqttClient to be 9s")
	}

	expectedTopic := "test/%Prefix%tele/%ClientId%/LWT"
	if config.MqttClients()[0].AvailabilityTopic() != expectedTopic {
		t.Errorf("expect AvailabilityTopic of first MqttClient to be '%s'", expectedTopic)
	}

	if config.MqttClients()[0].TopicPrefix() != "piegn/" {
		t.Error("expect TopicPrefix of first MqttClient to be 'piegn/'")
	}

	if config.MqttClients()[0].LogMessages() {
		t.Error("expect LogMessages of first MqttClient to be False")
	}

	if !config.MqttClients()[0].LogDebug() {
		t.Error("expect LogDebug of first MqttClient to be True")
	}

	if !config.MqttClients()[1].LogMessages() {
		t.Error("expect LogMessages of second MqttClient to be True")
	}

	if config.MqttClients()[1].LogDebug() {
		t.Error("expect LogDebug of second MqttClient to be False")
	}

	// influxClients section
	if len(config.InfluxClients()) != 2 {
		t.Error("expect len(config.InfluxClients) == 2")
	}

	if config.InfluxClients()[0].Name() != "0-piegn" {
		t.Errorf("expect Name of first InfluxClient to be '0-piegn' but got '%s'",
			config.InfluxClients()[0].Name(),
		)
	}

	if config.InfluxClients()[1].Name() != "1-local" {
		t.Errorf("expect Name of first InfluxClient to be '1-local' but got '%s'",
			config.InfluxClients()[1].Name(),
		)
	}

	if config.InfluxClients()[0].Url() != "http://172.17.0.2:8086" {
		t.Error("expect Address of first InfluxClient to be 'http://172.17.0.2:8086'")
	}

	if config.InfluxClients()[0].Token() != "mytoken" {
		t.Error("expect Token of first InfluxClient to be 'mytoken'")
	}

	if config.InfluxClients()[0].Org() != "myorg" {
		t.Error("expect Org of first InfluxClient to be 'myorg'")
	}

	if config.InfluxClients()[0].Bucket() != "mybucket" {
		t.Error("expect Bucket of first InfluxClient to be 'mybucket'")
	}

	if config.InfluxClients()[0].WriteInterval().String() != "400ms" {
		t.Error("expect WriteInterval of first InfluxClient to be '400ms'")
	}

	if config.InfluxClients()[0].TimePrecision().String() != "1ms" {
		t.Error("expect TimePrecision of first InfluxClient to be '1ms'")
	}

	if !config.InfluxClients()[0].LogDebug() {
		t.Error("expect LogDebug of first InfluxClient to be True")
	}

	// influxAuxiliaryTags section
	if len(config.InfluxAuxiliaryTags()) != 2 {
		t.Error("expect len(config.InfluxAuxiliaryTags) == 2")
	}

	{
		device := config.InfluxAuxiliaryTags()[0].Equals()
		if device == nil || *device != "foo" {
			t.Error("expect Device to be foo")
		}
	}

	if config.InfluxAuxiliaryTags()[0].Tag() != "device" {
		t.Error("expect Tag of first InfluxAuxiliaryTags to be device")
	}

	if !config.InfluxAuxiliaryTags()[0].MatchString("foo") {
		t.Error("expect DeviceMatcher of first InfluxAuxiliaryTags to match foo")
	}

	if config.InfluxAuxiliaryTags()[0].MatchString("bar") {
		t.Error("expect DeviceMatcher of first InfluxAuxiliaryTags not to match bar")
	}

	if config.InfluxAuxiliaryTags()[0].MatchString("fooBar") {
		t.Error("expect DeviceMatcher of first InfluxAuxiliaryTags not to match fooBar")
	}

	if len(config.InfluxAuxiliaryTags()[0].tagValues) != 3 {
		t.Error("expect len(TagValues) of first InfluxAuxiliaryTags to be 3")
	}

	if config.InfluxAuxiliaryTags()[1].Tag() != "field" {
		t.Error("expect Tag of first InfluxAuxiliaryTags to be field")
	}

	{
		device := config.InfluxAuxiliaryTags()[1].Equals()
		if device != nil {
			t.Error("expect Device to be nil")
		}
	}

	if len(config.InfluxAuxiliaryTags()[1].tagValues) != 1 {
		t.Error("expect len(TagValues) of second InfluxAuxiliaryTags to be 1")
	}

	if !config.InfluxAuxiliaryTags()[1].MatchString("temperature1") {
		t.Error("expect DeviceMatcher of second InfluxAuxiliaryTags to match temperature1")
	}

	if !config.InfluxAuxiliaryTags()[1].MatchString("temperature1-a") {
		t.Error("expect DeviceMatcher of second InfluxAuxiliaryTags to match temperature1-a")
	}

	if config.InfluxAuxiliaryTags()[1].MatchString("temperatureA") {
		t.Error("expect DeviceMatcher of second InfluxAuxiliaryTags not to match temperatureA")
	}

	// Converters section
	if len(config.Converters()) != 4 {
		t.Errorf("expect len(config.Converters) == 4")
	}

	if config.Converters()[0].Name() != "0-piegn-ve-sensor" {
		t.Errorf("expect Name of first Converter to be '0-piegn-ve-sensor' but got '%s'",
			config.Converters()[0].Name(),
		)
	}

	if config.Converters()[0].Implementation() != "go-iotdevice" {
		t.Error("expect Implementation of first Converter to be 'go-iotdevice'")
	}

	if len(config.Converters()[0].MqttTopics()) != 1 || config.Converters()[0].MqttTopics()[0].Topic() != "piegn/tele/ve/#" {
		t.Errorf("expect Topic of first MqttTopics of first Converter to be 'piegn/tele/ve/#' got '%s'",
			config.Converters()[0].MqttTopics()[0].Topic(),
		)
	}

	if len(config.Converters()[0].MqttTopics()) != 1 || config.Converters()[0].MqttTopics()[0].Device() != "fixed-ve" {
		t.Errorf("expect Device of first MqttTopics of first Converter to be 'fixed-ve' got '%s'",
			config.Converters()[0].MqttTopics()[0].Device(),
		)
	}

	if config.Converters()[1].MqttTopics()[0].Topic() != "piegn/tele/%Device%/LWT" {
		t.Errorf("expect Topic of first MqttTopics of second Converter to be 'piegn/tele/%%Device%%/LWT' got '%s'",
			config.Converters()[1].MqttTopics()[0].Topic(),
		)
	}

	if config.Converters()[1].MqttTopics()[0].Device() != "+" {
		t.Errorf("expect Device of first MqttTopics of second Converter to be '+' got '%s'",
			config.Converters()[1].MqttTopics()[0].Device(),
		)
	}

	if len(config.Converters()[0].MqttClients()) != 2 ||
		config.Converters()[0].MqttClients()[0] != "0-piegn-mosquitto" ||
		config.Converters()[0].MqttClients()[1] != "1-local-mosquitto" {
		t.Errorf("expect MqttClients of first Converter to be ['0-piegn-mosquitto', '1-local-mosquitto'] got %v",
			config.Converters()[0].MqttClients(),
		)
	}

	if len(config.Converters()[0].InfluxClients()) != 2 ||
		config.Converters()[0].InfluxClients()[0] != "0-piegn" ||
		config.Converters()[0].InfluxClients()[1] != "1-local" {
		t.Errorf("expect InfluxClients of first Converter to be ['0-piegn', '1-local'] got %v",
			config.Converters()[0].InfluxClients(),
		)
	}

	if !config.Converters()[0].LogHandleOnce() {
		t.Error("expect LogHandleOnce of first Converter to be True")
	}

	// HttpServer
	if !config.HttpServer().Enabled() {
		t.Error("expect HttpServer->Enabled to be True")
	}

	if config.HttpServer().Bind() != "0.0.0.0" {
		t.Error("expect HttpServer->Bind to be '0.0.0.0'")
	}

	if config.HttpServer().Port() != 80 {
		t.Error("expect HttpServer->Port to be 80")
	}

	if !config.HttpServer().LogRequests() {
		t.Error("expect HttpServer->LogRequests to be True")
	}

	// LocalDb
	if !config.LocalDb().Enabled() {
		t.Error("expect LocalDb->Enabled to be True")
	}

	if config.LocalDb().Path() != "/tmp/foobar.db" {
		t.Errorf("expect LocalDb->Path to be '/tmp/foobar.db', got '%s'",
			config.LocalDb().Path())
	}

	// Statistics
	if !config.Statistics().Enabled() {
		t.Error("expect Statistics->Enabled to be True")
	}

	if config.Statistics().HistoryResolution().String() != "100ms" {
		t.Errorf("expect Statistics->HistoryResolution to be '100ms', got '%s'",
			config.Statistics().HistoryResolution())
	}

	if config.Statistics().HistoryMaxAge().String() != "1h0m0s" {
		t.Errorf("expect Statistics->HistoryMaxAge to be '1h0m0s', got '%s'",
			config.Statistics().HistoryMaxAge())
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
	if config.LogConfig() {
		t.Error("expect LogConfig to be False by default")
	}

	if config.LogWorkerStart() {
		t.Error("expect LogWorkerStart to be False by default")
	}

	// influxClients section
	if config.MqttClients()[0].ProtocolVersion() != 3 {
		t.Error("expect default MqttClient->Protocol to be 3")
	}

	if config.MqttClients()[0].User() != "" {
		t.Error("expect default MqttClient->User to be empty")
	}

	if config.MqttClients()[0].Password() != "" {
		t.Error("expect default MqttClient->Password to be empty")
	}

	if !strings.Contains(config.MqttClients()[0].ClientId(), "go-mqtt-to-influx-") {
		t.Error("expect default MqttClient->ClientId to contain 'go-mqtt-to-influx'")
	}

	if config.MqttClients()[0].Qos() != 1 {
		t.Error("expect default MqttClient->Qos to be 1")
	}

	if config.MqttClients()[0].KeepAlive().String() != "10s" {
		t.Error("expect default MqttClient->KeepAlive to be 10s")
	}

	if config.MqttClients()[0].ConnectRetryDelay().String() != "1m0s" {
		t.Error("expect default MqttClient->ConnectRetryDelay to be 1m0s")
	}

	if config.MqttClients()[0].ConnectTimeout().String() != "10s" {
		t.Error("expect default MqttClient->ConnectTimeout to be 10s")
	}

	expectedAvailabilityTopic := "%Prefix%tele/%ClientId%/status"
	if config.MqttClients()[0].AvailabilityTopic() != expectedAvailabilityTopic {
		t.Errorf("expect default MqttClient->AvailabilityTopic to be '%s'", expectedAvailabilityTopic)
	}

	if config.MqttClients()[0].TopicPrefix() != "" {
		t.Error("expect default MqttClient->TopicPrefix to be empty")
	}

	if config.MqttClients()[0].LogMessages() {
		t.Error("expect default MqttClient->LogMessages to be False")
	}

	if config.MqttClients()[0].LogDebug() {
		t.Error("expect default MqttClient->LogDebug to be False")
	}

	// influxClients section
	if config.InfluxClients()[0].WriteInterval().String() != "5s" {
		t.Error("expect default InfluxClient->WriteInterval to be 5s")
	}

	if config.InfluxClients()[0].TimePrecision().String() != "1s" {
		t.Error("expect default InfluxClient->TimePrecision to be 1s")
	}

	if config.InfluxClients()[0].LogDebug() {
		t.Error("expect default InfluxClient->LogDebug to be False")
	}

	// Converters section
	if len(config.Converters()[0].MqttClients()) != 0 {
		t.Error("expect default Converter->MqttClients to be empty")
	}

	if len(config.Converters()[0].InfluxClients()) != 0 {
		t.Error("expect default Converter->InfluxClients to be empty")
	}

	if config.Converters()[0].MqttTopics()[0].Device() != "+" {
		t.Error("expect default Converter->MqttTopics->Device to be '+'")
	}

	// HttpServer
	if config.HttpServer().Enabled() {
		t.Error("expect default HttpServer->Enabled to be False")
	}

	if config.HttpServer().Bind() != "[::1]" {
		t.Error("expect default HttpServer->Bind to be '[::1]'")
	}

	if config.HttpServer().Port() != 8042 {
		t.Error("expect default HttpServer->Port to be 8042")
	}

	if config.HttpServer().LogRequests() {
		t.Error("expect default HttpServer->LogRequests to be False")
	}

	// LocalDb
	if !config.LocalDb().Enabled() {
		t.Error("expect LocalDb->Enabled to be True")
	}

	if config.LocalDb().Path() != "./go-mqtt-to-influx.db" {
		t.Errorf("expect LocalDb->Path to be './go-mqtt-to-influx.db', got '%s'",
			config.LocalDb().Path())
	}

	// Statistics
	if config.Statistics().Enabled() {
		t.Error("expect default Statistics->Enabled to be False")
	}

	if config.Statistics().HistoryResolution().String() != "1s" {
		t.Errorf("expect default Statistics->HistoryResolution to be '1s', got '%s'",
			config.Statistics().HistoryResolution())
	}

	if config.Statistics().HistoryMaxAge().String() != "10m0s" {
		t.Errorf("expect default Statistics->HistoryMaxAge to be '10m0s', got '%s'",
			config.Statistics().HistoryMaxAge())
	}
}
