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
		MqttClient:     i.MqttClient.TransformAndValidate(),
		InfluxDbClient: i.InfluxDbClient.TransformAndValidate(),
		Converters:     i.Converters.TransformAndValidate(),
	}

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

func (i *MqttClientConfigRead) TransformAndValidate() MqttClientConfig {
	if i == nil {
		log.Fatalf("config: MqttClientConfig section must be defined")
	}

	ret := MqttClientConfig{
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
		ret.AvailabilityTopic = "%Prefix%%ClientId%/LWT"
	}

	if i.DebugLog != nil && *i.DebugLog {
		ret.DebugLog = true
	}

	return ret
}

func (i *InfluxDbClientConfigRead) TransformAndValidate() InfluxDbClientConfig {
	if i == nil {
		log.Fatalf("config: InfluxDbClient section best be defined")
	}

	ret := InfluxDbClientConfig{
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

func (i ConverterReadMap) TransformAndValidate() []ConverterConfig {
	ret := make([]ConverterConfig, len(i))
	j := 0
	for name, converter := range i {
		ret[j] = converter.TransformAndValidate(name)
		j += 1
	}
	return ret
}

func (i ConverterConfigRead) TransformAndValidate(name string) ConverterConfig {
	ret := ConverterConfig{
		Name:              name,
		Implementation:    i.Implementation,
		TargetMeasurement: i.TargetMeasurement,
		MqttTopics:        i.MqttTopics,
	}

	if len(ret.MqttTopics) < 1 {
		log.Fatalf("config: Converters->%s->MqttTopics must not be empty", name)
	}

	return ret
}
