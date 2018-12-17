package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"
)

func ReadConfig(source string) (config Config) {
	configDir := filepath.Dir(source) + "/"

	log.Printf("config: load configuration source=%v, configDir=%v", source, configDir)

	var err error
	yamlStr, err := ioutil.ReadFile(source)
	if err != nil {
		log.Fatalf("config: cannot read configuration: %v", err)
	}

	var configRead ConfigRead

	err = yaml.Unmarshal(yamlStr, &configRead)
	if err != nil {
		log.Fatalf("config: cannot parse yaml: %v", err)
	}

	config = configRead.TransformAndValidate()

	newYamlStr, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("config: cannot encode yaml again:", err)
	}

	log.Print("config: use the following config:")
	for _, line := range strings.Split(string(newYamlStr), "\n") {
		log.Print("config: ", line)
	}

	return
}

func (i ConfigRead) TransformAndValidate() Config {
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

	return ret
}

func (i MqttClientConfigReadMap) TransformAndValidate() []MqttClientConfig {
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

func (i *MqttClientConfigRead) TransformAndValidate(name string) MqttClientConfig {
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

	if i.DebugLog != nil && *i.DebugLog {
		ret.DebugLog = true
	}

	return ret
}

func (i InfluxDbClientConfigReadMap) TransformAndValidate() []InfluxDbClientConfig {
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

func (i InfluxDbClientConfigRead) TransformAndValidate(name string) InfluxDbClientConfig {
	ret := InfluxDbClientConfig{
		Name:          name,
		Address:       i.Address,
		User:          i.User,
		Password:      i.Password,
		Database:      i.Database,
		WriteInterval: 0,
	}

	if len(ret.Address) < 1 {
		log.Fatalf("config: InfluxDbClientConfig->Address must not be empty")
	}

	if len(ret.Database) < 1 {
		ret.Database = "go-mqtt-to-influxdb"
	}

	if len(i.WriteInterval) < 1 {
		// use default 0
	} else if writeInterval, err := time.ParseDuration(i.WriteInterval); err != nil {
		log.Fatalf("config: InfluxDbClientConfig->WriteInterval parse error: %v", err)
	} else {
		ret.WriteInterval = writeInterval
	}

	return ret
}

func (i ConverterReadMap) TransformAndValidate(
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

var implementationsAndDefaultMeaseurement = map[string]string{
	"go-ve-sensor":   "floatValue",
	"lwt":            "boolValue",
	"tasmota-state":  "boolValue",
	"tasmota-sensor": "floatValue",
}

func (i ConverterConfigRead) TransformAndValidate(
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

	if def, ok := implementationsAndDefaultMeaseurement[ret.Implementation]; !ok {
		log.Fatalf("config:  Converters->%s->Implementation='%s' is unkown", name, ret.Implementation)
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
			log.Fatalf("config: Converters->%s->MqttClient=%s is not defined", name, clientName)
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
			log.Fatalf("config: Converters->%s->InfluxDbClient=%s is not defined", name, clientName)
		}
	}

	if len(ret.MqttTopics) < 1 {
		log.Fatalf("config: Converters->%s->MqttTopics must not be empty", name)
	}

	return ret
}
