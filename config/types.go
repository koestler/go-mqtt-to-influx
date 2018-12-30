package config

import "time"

type Config struct {
	Version         int                    `yaml:"Version"`         // must be 0
	MqttClients     []MqttClientConfig     `yaml:"MqttClient"`      // mandatory: at least 1 must be defined
	InfluxDbClients []InfluxDbClientConfig `yaml:"influxDbClients"` // mandatory: at least 1 must be defined
	Converters      []ConverterConfig      `yaml:"Converters"`      // mandatory: at least 1 must be defined
	HttpServer      HttpServerConfig       `yaml:"HttpServer"`      // optional: default Disabled
	Statistics      StatisticsConfig       `yaml:"Statistics"`      // optional: default Disabled
	LogConfig       bool                   `yaml:"LogConfig"`       // optional: default False
	LogWorkerStart  bool                   `yaml:"LogWorkerStart"`  // optional: default False
	LogMqttDebug    bool                   `yaml:"LogMqttDebug"`    // optional: default False
}

type MqttClientConfig struct {
	name              string // defined automatically by map key
	broker            string // mandatory
	user              string // optional: default empty
	password          string // optional: default empty
	clientId          string // optional: default go-mqtt-to-influxdb
	qos               byte   // optional: default 0, must be 0, 1, 2
	availabilityTopic string // optional: default %Prefix%tele/%clientId%/LWT
	topicPrefix       string // optional: default empty
	logMessages       bool   // optional: default False
}

type InfluxDbClientConfig struct {
	name            string        // defined automatically by map key
	address         string        // mandatory
	user            string        // optional: default empty
	password        string        // optional: default empty
	database        string        // optional: default go-mqtt-to-influxdb
	writeInterval   time.Duration // optional: default 200ms
	timePrecision   time.Duration // optional: default 1s
	logLineProtocol bool          // optional: default False
}

type ConverterConfig struct {
	name              string   // defined automatically by map key
	implementation    string   // mandatory
	targetMeasurement string   // optional: default depends on implementation
	mqttTopics        []string // mandatory: at least 1 must be defined
	mqttClients       []string // optional: defaults to all defined clients
	influxDbClients   []string // optional: defaults to all defined clients
	logHandleOnce     bool     // optional: default False
}

type HttpServerConfig struct {
	enabled     bool   // defined automatically if HttpServer section exists
	bind        string // optional: defaults to ::1 (ipv6 loopback)
	port        int    // optional: defaults to 8042
	logRequests bool   // optional:  default False
}

type StatisticsConfig struct {
	enabled bool // defined automatically if Statistics section exists
}

// Read structs are given to yaml for decoding and are slightly less exact in types
type configRead struct {
	Version         *int                        `yaml:"Version"`
	MqttClients     mqttClientConfigReadMap     `yaml:"mqttClients"`
	InfluxDbClients influxDbClientConfigReadMap `yaml:"influxDbClients"`
	Converters      converterReadMap            `yaml:"Converters"`
	HttpServer      *httpServerConfigRead       `yaml:"HttpServer"`
	Statistics      *statisticsConfigRead       `yaml:"Statistics"`
	LogConfig       *bool                       `yaml:"LogConfig"`
	LogWorkerStart  *bool                       `yaml:"LogWorkerStart"`
	LogMqttDebug    *bool                       `yaml:"LogMqttDebug"`
}

type mqttClientConfigRead struct {
	Broker            string  `yaml:"broker"`
	User              string  `yaml:"user"`
	Password          string  `yaml:"password"`
	ClientId          string  `yaml:"clientId"`
	Qos               *byte   `yaml:"qos"`
	AvailabilityTopic *string `yaml:"availabilityTopic"`
	TopicPrefix       string  `yaml:"topicPrefix"`
	LogMessages       *bool   `yaml:"logMessages"`
}

type mqttClientConfigReadMap map[string]mqttClientConfigRead

type influxDbClientConfigRead struct {
	Address         string `yaml:"address"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	WriteInterval   string `yaml:"writeInterval"`
	TimePrecision   string `yaml:"timePrecision"`
	LogLineProtocol *bool  `yaml:"logLineProtocol"`
}

type influxDbClientConfigReadMap map[string]influxDbClientConfigRead

type converterConfigRead struct {
	Implementation    string   `yaml:"implementation"`
	TargetMeasurement string   `yaml:"targetMeasurement"`
	MqttTopics        []string `yaml:"mqttTopics"`
	MqttClients       []string `yaml:"mqttClients"`
	InfluxDbClients   []string `yaml:"influxDbClients"`
	LogHandleOnce     *bool    `yaml:"logHandleOnce"`
}

type converterReadMap map[string]converterConfigRead

type httpServerConfigRead struct {
	Bind        string `yaml:"bind"`
	Port        *int   `yaml:"Port"`
	LogRequests *bool  `yaml:"logRequests"`
}

type statisticsConfigRead struct {
	Enabled *bool `yaml:"Enabled"`
}
