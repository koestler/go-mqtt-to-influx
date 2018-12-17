package config

import "time"

type Config struct {
	Version         int                    `yaml:"Version"`
	MqttClients     []MqttClientConfig     `yaml:"MqttClient"`
	InfluxDbClients []InfluxDbClientConfig `yaml:"InfluxDbClients"`
	Converters      []ConverterConfig      `yaml:"Converters"`
	LogConfig       bool                   `yaml:"LogConfig"`
	LogWorkerStart  bool                   `yaml:"LogWorkerStart"`
	LogMqttDebug    bool                   `yaml:"LogMqttDebug"`
}

type ConfigRead struct {
	Version         *int                        `yaml:"Version"`
	MqttClients     MqttClientConfigReadMap     `yaml:"MqttClients"`
	InfluxDbClients InfluxDbClientConfigReadMap `yaml:"InfluxDbClients"`
	Converters      ConverterReadMap            `yaml:"Converters"`
	LogConfig       *bool                       `yaml:"LogConfig"`
	LogWorkerStart  *bool                       `yaml:"LogWorkerStart"`
	LogMqttDebug    *bool                       `yaml:"LogMqttDebug"`
}

type MqttClientConfig struct {
	Name              string `yaml:"Name"`
	Broker            string `yaml:"Broker"`
	User              string `yaml:"User"`
	Password          string `yaml:"Password"`
	ClientId          string `yaml:"ClientId"`
	Qos               byte   `yaml:"Qos"`
	AvailabilityTopic string `yaml:"AvailabilityTopic"`
	TopicPrefix       string `yaml:"TopicPrefix"`
	LogMessages       bool   `yaml:"LogMessages"`
}

type MqttClientConfigRead struct {
	Broker            string `yaml:"Broker"`
	User              string `yaml:"User"`
	Password          string `yaml:"Password"`
	ClientId          string `yaml:"ClientId"`
	Qos               *byte  `yaml:"Qos"`
	TopicPrefix       string `yaml:"TopicPrefix"`
	AvailabilityTopic string `yaml:"AvailabilityTopic"`
	LogMessages       *bool  `yaml:"LogMessages"`
}

type MqttClientConfigReadMap map[string]MqttClientConfigRead

type InfluxDbClientConfig struct {
	Name            string        `yaml:"Name"`
	Address         string        `yaml:"Address"`
	User            string        `yaml:"User"`
	Password        string        `yaml:"Password"`
	Database        string        `yaml:"Database"`
	WriteInterval   time.Duration `yaml:"WriteInterval"`
	TimePrecision   time.Duration `yaml:"TimePrecision "`
	LogLineProtocol bool          `yaml:"LogLineProtocol"`
}

type InfluxDbClientConfigRead struct {
	Address         string `yaml:"Address"`
	User            string `yaml:"User"`
	Password        string `yaml:"Password"`
	Database        string `yaml:"Database"`
	WriteInterval   string `yaml:"WriteInterval"`
	TimePrecision   string `yaml:"TimePrecision "`
	LogLineProtocol *bool  `yaml:"LogLineProtocol"`
}

type InfluxDbClientConfigReadMap map[string]InfluxDbClientConfigRead

type ConverterConfig struct {
	Name              string   `yaml:"Name"`
	Implementation    string   `yaml:"Implementation"`
	TargetMeasurement string   `yaml:"TargetMeasurement"`
	MqttTopics        []string `yaml:"MqttTopics"`
	MqttClients       []string `yaml:"MqttClients"`
	InfluxDbClients   []string `yaml:"InfluxDbClients"`
	LogHandleOnce     bool     `yaml:"LogHandleOnce"`
}

type ConverterConfigRead struct {
	Implementation    string   `yaml:"Implementation"`
	TargetMeasurement string   `yaml:"TargetMeasurement"`
	MqttTopics        []string `yaml:"MqttTopics"`
	MqttClients       []string `yaml:"MqttClients"`
	InfluxDbClients   []string `yaml:"InfluxDbClients"`
	LogHandleOnce     *bool    `yaml:"LogHandleOnce"`
}

type ConverterReadMap map[string]ConverterConfigRead
