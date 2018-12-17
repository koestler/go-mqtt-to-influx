package config

import "time"

type Config struct {
	Version         int                    `yaml:"Version"`
	MqttClients     []MqttClientConfig     `yaml:"MqttClient"`
	InfluxDbClients []InfluxDbClientConfig `yaml:"InfluxDbClients"`
	Converters      []ConverterConfig      `yaml:"Converters"`
}

type ConfigRead struct {
	Version         *int                        `yaml:"Version"`
	MqttClients     MqttClientConfigReadMap     `yaml:"MqttClients"`
	InfluxDbClients InfluxDbClientConfigReadMap `yaml:"InfluxDbClients"`
	Converters      ConverterReadMap            `yaml:"Converters"`
}

type MqttClientConfig struct {
	Name              string `yaml:"Name"`
	Broker            string `yaml:"Broker"`
	User              string `yaml:"User"`
	Password          string `yaml:"Password"`
	ClientId          string `yaml:"ClientId"`
	Qos               byte   `yaml:"Qos"`
	DebugLog          bool   `yaml:"DebugLog"`
	TopicPrefix       string `yaml:"TopicPrefix"`
	AvailabilityTopic string `yaml:"AvailabilityTopic"`
}

type MqttClientConfigRead struct {
	Broker            string `yaml:"Broker"`
	User              string `yaml:"User"`
	Password          string `yaml:"Password"`
	ClientId          string `yaml:"ClientId"`
	Qos               *byte  `yaml:"Qos"`
	DebugLog          *bool  `yaml:"DebugLog"`
	TopicPrefix       string `yaml:"TopicPrefix"`
	AvailabilityTopic string `yaml:"AvailabilityTopic"`
}

type MqttClientConfigReadMap map[string]MqttClientConfigRead

type InfluxDbClientConfig struct {
	Name          string        `yaml:"Name"`
	Address       string        `yaml:"Address"`
	User          string        `yaml:"User"`
	Password      string        `yaml:"Password"`
	Database      string        `yaml:"Database"`
	WriteInterval time.Duration `yaml:"WriteInterval"`
}

type InfluxDbClientConfigRead struct {
	Address       string `yaml:"Address"`
	User          string `yaml:"User"`
	Password      string `yaml:"Password"`
	Database      string `yaml:"Database"`
	WriteInterval string `yaml:"WriteInterval"`
}

type InfluxDbClientConfigReadMap map[string]InfluxDbClientConfigRead

type ConverterConfig struct {
	Name              string   `yaml:"Name"`
	Implementation    string   `yaml:"Implementation"`
	TargetMeasurement string   `yaml:"TargetMeasurement"`
	MqttTopics        []string `yaml:"MqttTopics"`
	MqttClients       []string `yaml:"MqttClients"`
	InfluxDbClients   []string `yaml:"InfluxDbClients"`
}

type ConverterConfigRead struct {
	Implementation    string   `yaml:"Implementation"`
	TargetMeasurement string   `yaml:"TargetMeasurement"`
	MqttTopics        []string `yaml:"MqttTopics"`
	MqttClients       []string `yaml:"MqttClients"`
	InfluxDbClients   []string `yaml:"InfluxDbClients"`
}

type ConverterReadMap map[string]ConverterConfigRead
