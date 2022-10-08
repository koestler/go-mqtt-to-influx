package config

import (
	"net/url"
	"regexp"
	"time"
)

type Config struct {
	version             int                    `yaml:"Version"`       // must be 0
	mqttClients         []*MqttClientConfig    `yaml:"MqttClient"`    // mandatory: at least 1 must be defined
	influxClients       []*InfluxClientConfig  `yaml:"InfluxClients"` // mandatory: at least 1 must be defined
	influxAuxiliaryTags []*InfluxAuxiliaryTags `yaml:"InfluxAuxiliaryTags"`
	converters          []*ConverterConfig     `yaml:"Converters"`     // mandatory: at least 1 must be defined
	httpServer          HttpServerConfig       `yaml:"HttpServer"`     // optional: default Disabled
	localDb             LocalDbConfig          `yaml:"LocalDb"`        // optional: default Disasbled
	statistics          StatisticsConfig       `yaml:"Statistics"`     // optional: default Disabled
	logConfig           bool                   `yaml:"LogConfig"`      // optional: default False
	logWorkerStart      bool                   `yaml:"LogWorkerStart"` // optional: default False
}

type MqttClientConfig struct {
	name              string        // defined automatically by map key
	broker            *url.URL      // mandatory
	protocolVersion   int           // optional: default 3
	user              string        // optional: default empty
	password          string        // optional: default empty
	clientId          string        // optional: default go-mqtt-to-influx-UUID
	qos               byte          // optional: default 1, must be 0, 1, 2
	keepAlive         time.Duration // optional: default 10s
	connectRetryDelay time.Duration // optional: default 1m
	connectTimeout    time.Duration // optional: default 10s
	availabilityTopic string        // optional: default %Prefix%tele/%ClientId%/status
	topicPrefix       string        // optional: default empty
	logDebug          bool          // optional: default False
	logMessages       bool          // optional: default False
}

type InfluxClientConfig struct {
	name          string        // defined automatically by map key
	url           string        // mandatory
	token         string        // mandatory
	org           string        // mandatory
	bucket        string        // mandatory
	writeInterval time.Duration // optional: default 5s
	timePrecision time.Duration // optional: default 1s
	logDebug      bool          // optional: default False
}

type InfluxAuxiliaryTags struct {
	tag       string // optional: defaults to device
	equals    *string
	matches   *string
	matcher   *regexp.Regexp
	tagValues map[string]string
}

type ConverterConfig struct {
	name           string             // defined automatically by map key
	implementation string             // mandatory
	mqttTopics     []*MqttTopicConfig // mandatory: at least 1 must be defined
	mqttClients    []string           // optional: defaults to all defined clients
	influxClients  []string           // optional: defaults to all defined clients
	logHandleOnce  bool               // optional: default False
}

type MqttTopicConfig struct {
	topic  string // mandatory: must contain %Device%
	device string // optional: default "+"
}

type HttpServerConfig struct {
	enabled     bool   // defined automatically if HttpServer section exists
	bind        string // optional: defaults to ::1 (ipv6 loopback)
	port        int    // optional: defaults to 8042
	logRequests bool   // optional:  default False
}

type LocalDbConfig struct {
	enabled bool   // defined automatically if LocalDbConfig section exists
	path    string `yaml:"Path"` // optional: defaults ./go-mqtt-to-influx.db
}

type StatisticsConfig struct {
	enabled           bool          // defined automatically if Statistics section exists
	historyResolution time.Duration // optional: defaults to 1s
	historyMaxAge     time.Duration // optional: default to 10min
}

// Read structs are given to yaml for decoding and are slightly less exact in types
type configRead struct {
	Version             *int                        `yaml:"Version"`
	MqttClients         mqttClientConfigReadMap     `yaml:"MqttClients"`
	InfluxClients       influxClientConfigReadMap   `yaml:"InfluxClients"`
	InfluxAuxiliaryTags influxAuxiliaryTagsReadList `yaml:"InfluxAuxiliaryTags"`
	Converters          converterConfigReadMap      `yaml:"Converters"`
	HttpServer          *httpServerConfigRead       `yaml:"HttpServer"`
	LocalDb             *localDbConfigRead          `yaml:"LocalDb"`
	Statistics          *statisticsConfigRead       `yaml:"Statistics"`
	LogConfig           *bool                       `yaml:"LogConfig"`
	LogWorkerStart      *bool                       `yaml:"LogWorkerStart"`
	LogMqttDebug        *bool                       `yaml:"LogMqttDebug"`
}

type mqttClientConfigRead struct {
	Broker            string  `yaml:"Broker"`
	ProtocolVersion   *int    `yaml:"ProtocolVersion"`
	User              string  `yaml:"User"`
	Password          string  `yaml:"Password"`
	ClientId          string  `yaml:"ClientId"`
	Qos               *byte   `yaml:"Qos"`
	KeepAlive         string  `yaml:"KeepAlive"`
	ConnectRetryDelay string  `yaml:"ConnectRetryDelay"`
	ConnectTimeout    string  `yaml:"ConnectTimeout"`
	AvailabilityTopic *string `yaml:"AvailabilityTopic"`
	TopicPrefix       string  `yaml:"TopicPrefix"`
	LogDebug          *bool   `yaml:"LogDebug"`
	LogMessages       *bool   `yaml:"LogMessages"`
}

type mqttClientConfigReadMap map[string]mqttClientConfigRead

type influxClientConfigRead struct {
	Url           string `yaml:"Url"`
	Token         string `yaml:"Token"`
	Org           string `yaml:"Org"`
	Bucket        string `yaml:"Bucket"`
	WriteInterval string `yaml:"WriteInterval"`
	TimePrecision string `yaml:"TimePrecision"`
	LogDebug      *bool  `yaml:"LogDebug"`
}

type influxClientConfigReadMap map[string]influxClientConfigRead

type influxAuxiliaryTagsRead struct {
	Tag       *string           `yaml:"Tag"`
	Equals    *string           `yaml:"Equals"`
	Matches   *string           `yaml:"Matches""`
	TagValues map[string]string `yaml:"TagValues"`
}

type influxAuxiliaryTagsReadList []influxAuxiliaryTagsRead

type converterConfigRead struct {
	Implementation string                  `yaml:"Implementation"`
	MqttTopics     mqttTopicConfigReadList `yaml:"MqttTopics"`
	MqttClients    []string                `yaml:"MqttClients"`
	InfluxClients  []string                `yaml:"InfluxClients"`
	LogHandleOnce  *bool                   `yaml:"LogHandleOnce"`
}

type converterConfigReadMap map[string]converterConfigRead

type mqttTopicConfigRead struct {
	Topic  string  `yaml:"Topic"`
	Device *string `yaml:"Device"`
}

type mqttTopicConfigReadList []mqttTopicConfigRead

type httpServerConfigRead struct {
	Bind        string `yaml:"Bind"`
	Port        *int   `yaml:"Port"`
	LogRequests *bool  `yaml:"LogRequests"`
}

type localDbConfigRead struct {
	Enabled *bool   `yaml:"Enabled"`
	Path    *string `yaml:"Path"`
}

type statisticsConfigRead struct {
	Enabled           *bool  `yaml:"Enabled"`
	HistoryResolution string `yaml:"HistoryResolution"`
	HistoryMaxAge     string `yaml:"HistoryMaxAge"`
}
