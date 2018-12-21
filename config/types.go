package config

import "time"

type Config struct {
	Version         int                    `yaml:"Version"`         // must be 0
	MqttClients     []MqttClientConfig     `yaml:"MqttClient"`      // mandatory: at least 1 must be defined
	InfluxDbClients []InfluxDbClientConfig `yaml:"InfluxDbClients"` // mandatory: at least 1 must be defined
	Converters      []ConverterConfig      `yaml:"Converters"`      // mandatory: at least 1 must be defined
	HttpServer      HttpServerConfig       `yaml:"HttpServer"`      // optional: default Disabled
	Statistics      StatisticsConfig       `yaml:"Statistics"`      // optional: default Disabled
	LogConfig       bool                   `yaml:"LogConfig"`       // optional: default False
	LogWorkerStart  bool                   `yaml:"LogWorkerStart"`  // optional: default False
	LogMqttDebug    bool                   `yaml:"LogMqttDebug"`    // optional: default False
}

type MqttClientConfig struct {
	Name              string `yaml:"Name"`              // defined automatically by map key
	Broker            string `yaml:"Broker"`            // mandatory
	User              string `yaml:"User"`              // optional: default empty
	Password          string `yaml:"Password"`          // optional: default empty
	ClientId          string `yaml:"ClientId"`          // optional: default go-mqtt-to-influxdb
	Qos               byte   `yaml:"Qos"`               // optional: default 0, must be 0, 1, 2
	AvailabilityTopic string `yaml:"AvailabilityTopic"` // optional: default %Prefix%tele/%ClientId%/LWT
	TopicPrefix       string `yaml:"TopicPrefix"`       // optional: default empty
	LogMessages       bool   `yaml:"LogMessages"`       // optional: default False
}

type InfluxDbClientConfig struct {
	Name            string        `yaml:"Name"`            // defined automatically by map key
	Address         string        `yaml:"Address"`         // mandatory
	User            string        `yaml:"User"`            // optional: default empty
	Password        string        `yaml:"Password"`        // optional: default empty
	Database        string        `yaml:"Database"`        // optional: default go-mqtt-to-influxdb
	WriteInterval   time.Duration `yaml:"WriteInterval"`   // optional: default 200ms
	TimePrecision   time.Duration `yaml:"TimePrecision"`   // optional: default 1s
	LogLineProtocol bool          `yaml:"LogLineProtocol"` // optional: default False
}

type ConverterConfig struct {
	Name              string   `yaml:"Name"`              // defined automatically by map key
	Implementation    string   `yaml:"Implementation"`    // mandatory
	TargetMeasurement string   `yaml:"TargetMeasurement"` // optional: default depends on implementation
	MqttTopics        []string `yaml:"MqttTopics"`        // mandatory: at least 1 must be defined
	MqttClients       []string `yaml:"MqttClients"`       // optional: defaults to all defined clients
	InfluxDbClients   []string `yaml:"InfluxDbClients"`   // optional: defaults to all defined clients
	LogHandleOnce     bool     `yaml:"LogHandleOnce"`     // optional: default False
}

type HttpServerConfig struct {
	enabled     bool   `yaml:"Enabled"`     // defined automatically if HttpServer section exists
	bind        string `yaml:"Bind"`        // optional: defaults to ::1 (ipv6 loopback)
	port        int    `yaml:"Port"`        // optional: defaults to 8042
	logRequests bool   `yaml:"LogRequests"` // optional:  default False
}

type StatisticsConfig struct {
	Enabled bool `yaml:"enabled"` // defined automatically if Statistics section exists
}

// Read structs are given to yaml for decoding and are slightly less exact in types
type configRead struct {
	Version         *int                        `yaml:"Version"`
	MqttClients     mqttClientConfigReadMap     `yaml:"MqttClients"`
	InfluxDbClients influxDbClientConfigReadMap `yaml:"InfluxDbClients"`
	Converters      converterReadMap            `yaml:"Converters"`
	HttpServer      *httpServerConfigRead       `yaml:"HttpServer"`
	Statistics      *statisticsConfigRead       `yaml:"Statistics"`
	LogConfig       *bool                       `yaml:"LogConfig"`
	LogWorkerStart  *bool                       `yaml:"LogWorkerStart"`
	LogMqttDebug    *bool                       `yaml:"LogMqttDebug"`
}

type mqttClientConfigRead struct {
	Broker            string  `yaml:"Broker"`
	User              string  `yaml:"User"`
	Password          string  `yaml:"Password"`
	ClientId          string  `yaml:"ClientId"`
	Qos               *byte   `yaml:"Qos"`
	AvailabilityTopic *string `yaml:"AvailabilityTopic"`
	TopicPrefix       string  `yaml:"TopicPrefix"`
	LogMessages       *bool   `yaml:"LogMessages"`
}

type mqttClientConfigReadMap map[string]mqttClientConfigRead

type influxDbClientConfigRead struct {
	Address         string `yaml:"Address"`
	User            string `yaml:"User"`
	Password        string `yaml:"Password"`
	Database        string `yaml:"Database"`
	WriteInterval   string `yaml:"WriteInterval"`
	TimePrecision   string `yaml:"TimePrecision"`
	LogLineProtocol *bool  `yaml:"LogLineProtocol"`
}

type influxDbClientConfigReadMap map[string]influxDbClientConfigRead

type converterConfigRead struct {
	Implementation    string   `yaml:"Implementation"`
	TargetMeasurement string   `yaml:"TargetMeasurement"`
	MqttTopics        []string `yaml:"MqttTopics"`
	MqttClients       []string `yaml:"MqttClients"`
	InfluxDbClients   []string `yaml:"InfluxDbClients"`
	LogHandleOnce     *bool    `yaml:"LogHandleOnce"`
}

type converterReadMap map[string]converterConfigRead

type httpServerConfigRead struct {
	Bind        string `yaml:"bind"`
	Port        *int   `yaml:"Port"`
	LogRequests *bool  `yaml:"logRequests"`
}

type statisticsConfigRead struct {
	HttpServer *httpServerConfigRead `yaml:"HttpServer"`
}
