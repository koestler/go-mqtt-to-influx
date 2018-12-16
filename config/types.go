package config

import "time"

type ConverterReadMap map[string]ConverterConfigRead

type Config struct {
	Version        int                  `yaml:"Version"`
	MqttClient     MqttClientConfig     `yaml:"MqttClient,omitempty"`
	InfluxDbClient InfluxDbClientConfig `yaml:"InfluxDbClient,omitempty"`
	Converters     []ConverterConfig    `yaml:"Converters"`
}

type ConfigRead struct {
	Version        *int                      `yaml:"Version"`
	MqttClient     *MqttClientConfigRead     `yaml:"MqttClient,omitempty"`
	InfluxDbClient *InfluxDbClientConfigRead `yaml:"InfluxDbClient,omitempty"`
	Converters     ConverterReadMap          `yaml:"Converters"`
}

type MqttClientConfig struct {
	Broker            string `yaml:"Broker"`
	User              string `yaml:"User,omitempty"`
	Password          string `yaml:"Password,omitempty"`
	ClientId          string `yaml:"ClientId"`
	Qos               byte   `yaml:"Qos"`
	DebugLog          bool   `yaml:"DebugLog"`
	TopicPrefix       string `yaml:"TopicPrefix,omitempty"`
	AvailabilityTopic string `yaml:"AvailabilityTopic,omitempty"`
}

type MqttClientConfigRead struct {
	Broker            string `yaml:"Broker"`
	User              string `yaml:"User,omitempty"`
	Password          string `yaml:"Password,omitempty"`
	ClientId          string `yaml:"ClientId"`
	Qos               *byte  `yaml:"Qos"`
	DebugLog          *bool  `yaml:"DebugLog"`
	TopicPrefix       string `yaml:"TopicPrefix,omitempty"`
	AvailabilityTopic string `yaml:"AvailabilityTopic,omitempty"`
}

type InfluxDbClientConfig struct {
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

type ConverterConfig struct {
	Name              string   `yaml:"Name"`
	Implementation    string   `yaml:"Implementation"`
	TargetMeasurement string   `yaml:"TargetMeasurement"`
	MqttTopics        []string `yaml:"MqttTopics"`
}

type ConverterConfigRead struct {
	Implementation    string   `yaml:"Implementation"`
	TargetMeasurement string   `yaml:"TargetMeasurement"`
	MqttTopics        []string `yaml:"MqttTopics"`
}
