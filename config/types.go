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

type configRead struct {
	Version         *int                        `yaml:"Version"`
	MqttClients     mqttClientConfigReadMap     `yaml:"MqttClients"`
	InfluxDbClients influxDbClientConfigReadMap `yaml:"InfluxDbClients"`
	Converters      converterReadMap            `yaml:"Converters"`
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

type mqttClientConfigRead struct {
	Broker            string `yaml:"Broker"`
	User              string `yaml:"User"`
	Password          string `yaml:"Password"`
	ClientId          string `yaml:"ClientId"`
	Qos               *byte  `yaml:"Qos"`
	TopicPrefix       string `yaml:"TopicPrefix"`
	AvailabilityTopic string `yaml:"AvailabilityTopic"`
	LogMessages       *bool  `yaml:"LogMessages"`
}

type mqttClientConfigReadMap map[string]mqttClientConfigRead

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

type influxDbClientConfigRead struct {
	Address         string `yaml:"Address"`
	User            string `yaml:"User"`
	Password        string `yaml:"Password"`
	Database        string `yaml:"Database"`
	WriteInterval   string `yaml:"WriteInterval"`
	TimePrecision   string `yaml:"TimePrecision "`
	LogLineProtocol *bool  `yaml:"LogLineProtocol"`
}

type influxDbClientConfigReadMap map[string]influxDbClientConfigRead

type ConverterConfig struct {
	Name              string   `yaml:"Name"`
	Implementation    string   `yaml:"Implementation"`
	TargetMeasurement string   `yaml:"TargetMeasurement"`
	MqttTopics        []string `yaml:"MqttTopics"`
	MqttClients       []string `yaml:"MqttClients"`
	InfluxDbClients   []string `yaml:"InfluxDbClients"`
	LogHandleOnce     bool     `yaml:"LogHandleOnce"`
}

type converterConfigRead struct {
	Implementation    string   `yaml:"Implementation"`
	TargetMeasurement string   `yaml:"TargetMeasurement"`
	MqttTopics        []string `yaml:"MqttTopics"`
	MqttClients       []string `yaml:"MqttClients"`
	InfluxDbClients   []string `yaml:"InfluxDbClients"`
	LogHandleOnce     *bool    `yaml:"LogHandleOnce"`
}

type converterReadMap map[string]converterConfigRead
