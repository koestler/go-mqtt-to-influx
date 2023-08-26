package config

import (
	"net/url"
	"regexp"
	"time"
)

type Config struct {
	version             int                   // must be 0
	httpServer          HttpServerConfig      // optional: default Disabled
	localDb             LocalDbConfig         // optional: default Disabled
	statistics          StatisticsConfig      // optional: default Disabled
	logConfig           bool                  // optional: default False
	logWorkerStart      bool                  // optional: default False
	mqttClients         []MqttClientConfig    // mandatory: at least 1 must be defined
	influxClients       []InfluxClientConfig  // mandatory: at least 1 must be defined
	converters          []ConverterConfig     // mandatory: at least 1 must be defined
	influxAuxiliaryTags []InfluxAuxiliaryTags // optional: default empty
}

type HttpServerConfig struct {
	enabled     bool   // defined automatically if HttpServer section exists
	bind        string // optional: defaults to [::1] (ipv6 loopback)
	port        int    // optional: defaults to 8000
	logRequests bool   // optional:  default False
}

type LocalDbConfig struct {
	enabled bool   // defined automatically if LocalDbConfig section exists
	path    string // optional: defaults ./go-mqtt-to-influx.db
}

type StatisticsConfig struct {
	enabled           bool          // defined automatically if Statistics section exists
	historyResolution time.Duration // optional: defaults to 10s
	historyMaxAge     time.Duration // optional: default to 10min
}

type MqttClientConfig struct {
	name              string        // defined automatically by map key
	broker            *url.URL      // mandatory
	protocolVersion   int           // optional: default 5
	user              string        // optional: default empty
	password          string        // optional: default empty
	clientId          string        // optional: default go-mqtt-to-influx-UUID
	qos               byte          // optional: default 1, must be 0, 1, 2
	keepAlive         time.Duration // optional: default 60s
	connectRetryDelay time.Duration // optional: default 10s
	connectTimeout    time.Duration // optional: default 5s
	availabilityTopic string        // optional: default %Prefix%tele/%ClientId%/status
	topicPrefix       string        // optional: default empty
	logDebug          bool          // optional: default False
	logMessages       bool          // optional: default False
}

type InfluxClientConfig struct {
	name              string        // defined automatically by map key
	url               string        // mandatory
	token             string        // mandatory
	org               string        // mandatory
	bucket            string        // mandatory
	writeInterval     time.Duration // optional: default 10s
	retryInterval     time.Duration // optional: default 10s
	aggregateInterval time.Duration // optional: default 60s
	timePrecision     time.Duration // optional: default 1s
	connectTimeout    time.Duration // optional: default 5s
	batchSize         uint          // optional: default 5000
	retryQueueLimit   uint          // optional: default 20
	logDebug          bool          // optional: default False
}

type ConverterConfig struct {
	name           string            // defined automatically by map key
	implementation string            // mandatory
	mqttTopics     []MqttTopicConfig // mandatory: at least 1 must be defined
	mqttClients    []string          // optional: defaults to all defined clients
	influxClients  []string          // optional: defaults to all defined clients
	logHandleOnce  bool              // optional: default False
	logDebug       bool              // optional: default False
}

type MqttTopicConfig struct {
	topic  string // mandatory: must contain %Device%
	device string // optional: default "+"
}

type InfluxAuxiliaryTags struct {
	tag       string            // optional: defaults to "device"
	equals    *string           // optional: if not set, matches must be set
	matches   *string           // optional: if not set, equals must be set
	matcher   *regexp.Regexp    // used internally
	tagValues map[string]string // mandatory: must not be empty
}
