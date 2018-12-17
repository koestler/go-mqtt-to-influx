package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"time"
)

const NameRegexp = "^[a-zA-Z0-9\\-]{1,32}$"

var nameMatcher = regexp.MustCompile(NameRegexp)

func ReadConfig(exe, source string) (config Config) {
	var err error
	yamlStr, err := ioutil.ReadFile(source)
	if err != nil {
		log.Fatalf("config: cannot read configuration: %v; use see `%s --help`", err, exe)
	}

	var configRead configRead

	err = yaml.Unmarshal(yamlStr, &configRead)
	if err != nil {
		log.Fatalf("config: cannot parse yaml: %s", err)
	}

	config = configRead.TransformAndValidate()

	return
}

func (config Config) PrintConfig() {
	newYamlStr, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("config: cannot encode yaml again: %s", err)
	}

	log.Print("config: use the following config:")
	for _, line := range strings.Split(string(newYamlStr), "\n") {
		log.Print("config: ", line)
	}
}

func (i configRead) TransformAndValidate() Config {
	ret := Config{
		MqttClients:     i.MqttClients.TransformAndValidate(),
		InfluxDbClients: i.InfluxDbClients.TransformAndValidate(),
	}
	ret.Converters = i.Converters.TransformAndValidate(ret.MqttClients, ret.InfluxDbClients)

	if i.Version == nil {
		log.Fatalf("config: Version must be defined; use Version=0")
	} else {
		ret.Version = *i.Version
		if ret.Version != 0 {
			log.Fatalf("config: Version=%d not supported", ret.Version)
		}
	}

	if i.LogConfig != nil && *i.LogConfig {
		ret.LogConfig = true
	}

	if i.LogWorkerStart != nil && *i.LogWorkerStart {
		ret.LogWorkerStart = true
	}

	if i.LogMqttDebug != nil && *i.LogMqttDebug {
		ret.LogMqttDebug = true
	}

	return ret
}

func (i mqttClientConfigReadMap) TransformAndValidate() []MqttClientConfig {
	if len(i) < 1 {
		log.Fatalf("config: MqttClients section must no be empty")
	}

	ret := make([]MqttClientConfig, len(i))
	j := 0
	for name, client := range i {
		ret[j] = client.TransformAndValidate(name)
		j += 1
	}
	return ret
}

func (i *mqttClientConfigRead) TransformAndValidate(name string) MqttClientConfig {
	if i == nil {
		log.Fatalf("config: MqttClientConfig section must be defined")
	}

	ret := MqttClientConfig{
		Name:              name,
		Broker:            i.Broker,
		User:              i.User,
		Password:          i.Password,
		ClientId:          i.ClientId,
		TopicPrefix:       i.TopicPrefix,
		AvailabilityTopic: i.AvailabilityTopic,
	}

	if !nameMatcher.MatchString(ret.Name) {
		log.Fatalf("config: MqttClientConfig->Name='%s' does not match %s", ret.Name, NameRegexp)
	}

	if len(ret.Broker) < 1 {
		log.Fatalf("config: MqttClientConfig->Broker must not be empty")
	}
	if len(ret.ClientId) < 1 {
		ret.ClientId = "go-mqtt-to-influxdb"
	}
	if i.Qos == nil {
		ret.Qos = 1
	} else if *i.Qos == 1 || *i.Qos == 2 {
		ret.Qos = *i.Qos
	}

	if len(ret.AvailabilityTopic) < 1 {
		ret.AvailabilityTopic = "%Prefix%tele/%ClientId%/LWT"
	}

	if i.LogMessages != nil && *i.LogMessages {
		ret.LogMessages = true
	}

	return ret
}

func (i influxDbClientConfigReadMap) TransformAndValidate() []InfluxDbClientConfig {
	if len(i) < 1 {
		log.Fatalf("config: InfluxDbClients section must no be empty")
	}

	ret := make([]InfluxDbClientConfig, len(i))
	j := 0
	for name, client := range i {
		ret[j] = client.TransformAndValidate(name)
		j += 1
	}
	return ret
}

func (i influxDbClientConfigRead) TransformAndValidate(name string) InfluxDbClientConfig {
	ret := InfluxDbClientConfig{
		Name:     name,
		Address:  i.Address,
		User:     i.User,
		Password: i.Password,
		Database: i.Database,
	}

	if !nameMatcher.MatchString(ret.Name) {
		log.Fatalf("config: InfluxDbClientConfig->Name='%s' does not match %s", ret.Name, NameRegexp)
	}

	if len(ret.Address) < 1 {
		log.Fatalf("config: InfluxDbClientConfig->Address must not be empty")
	}

	if len(ret.Database) < 1 {
		ret.Database = "go-mqtt-to-influxdb"
	}

	if len(i.WriteInterval) < 1 {
		// use default 0
		ret.WriteInterval = 0
	} else if writeInterval, err := time.ParseDuration(i.WriteInterval); err != nil {
		log.Fatalf("config: InfluxDbClientConfig->WriteInterval parse error: %s", err)
	} else {
		ret.WriteInterval = writeInterval
	}

	if len(i.TimePrecision) < 1 {
		// use default 1s
		ret.TimePrecision = time.Second
	} else if timePrecision, err := time.ParseDuration(i.TimePrecision); err != nil {
		log.Fatalf("config: InfluxDbClientConfig->TimePrecision parse error: %s", err)
	} else {
		ret.TimePrecision = timePrecision
	}

	if i.LogLineProtocol != nil && *i.LogLineProtocol {
		ret.LogLineProtocol = true
	}

	return ret
}

func (i converterReadMap) TransformAndValidate(
	mqttClients []MqttClientConfig,
	influxDbClients []InfluxDbClientConfig,
) []ConverterConfig {
	if len(i) < 1 {
		log.Fatalf("config: Converters section must no be empty")
	}

	ret := make([]ConverterConfig, len(i))
	j := 0
	for name, converter := range i {
		ret[j] = converter.TransformAndValidate(name, mqttClients, influxDbClients)
		j += 1
	}
	return ret
}

var implementationsAndDefaultMeasurement = map[string]string{
	"go-ve-sensor":   "floatValue",
	"lwt":            "boolValue",
	"tasmota-state":  "boolValue",
	"tasmota-sensor": "floatValue",
}

func (i converterConfigRead) TransformAndValidate(
	name string,
	mqttClients []MqttClientConfig,
	influxDbClients []InfluxDbClientConfig,
) ConverterConfig {
	ret := ConverterConfig{
		Name:              name,
		Implementation:    i.Implementation,
		TargetMeasurement: i.TargetMeasurement,
		MqttTopics:        i.MqttTopics,
		MqttClients:       i.MqttClients,
		InfluxDbClients:   i.InfluxDbClients,
	}

	if !nameMatcher.MatchString(ret.Name) {
		log.Fatalf("config: Converters->Name='%s' does not match %s", ret.Name, NameRegexp)
	}

	if def, ok := implementationsAndDefaultMeasurement[ret.Implementation]; !ok {
		log.Fatalf("config: Converters->%s->Implementation='%s' is unkown", name, ret.Implementation)
	} else if len(ret.TargetMeasurement) < 1 {
		ret.TargetMeasurement = def
	}

	// validate that all listed MqttClients exist
	for _, clientName := range ret.MqttClients {
		found := false
		for _, client := range mqttClients {
			if clientName == client.Name {
				found = true
				break
			}
		}

		if !found {
			log.Fatalf("config: Converters->%s->MqttClient='%s' is not defined", name, clientName)
		}
	}

	// validate that all listed InfluxDbClients exist
	for _, clientName := range ret.InfluxDbClients {
		found := false
		for _, client := range influxDbClients {
			if clientName == client.Name {
				found = true
				break
			}
		}

		if !found {
			log.Fatalf("config: Converters->%s->InfluxDbClient='%s' is not defined", name, clientName)
		}
	}

	if len(ret.MqttTopics) < 1 {
		log.Fatalf("config: Converters->%s->MqttTopics must not be empty", name)
	}

	if i.LogHandleOnce != nil && *i.LogHandleOnce {
		ret.LogHandleOnce = true
	}

	return ret
}
