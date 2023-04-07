package config

import (
	"net/url"
	"regexp"
	"time"
)

// getters for Config struct

func (c Config) Version() int {
	return c.version
}

func (c Config) HttpServer() HttpServerConfig {
	return c.httpServer
}

func (c Config) LocalDb() LocalDbConfig {
	return c.localDb
}

func (c Config) Statistics() StatisticsConfig {
	return c.statistics
}

func (c Config) LogConfig() bool {
	return c.logConfig
}

func (c Config) LogWorkerStart() bool {
	return c.logWorkerStart
}

func (c Config) MqttClients() []*MqttClientConfig {
	return c.mqttClients
}

func (c Config) InfluxClients() []*InfluxClientConfig {
	return c.influxClients
}

func (c Config) Converters() []*ConverterConfig {
	return c.converters
}

func (c Config) InfluxAuxiliaryTags() []*InfluxAuxiliaryTags {
	return c.influxAuxiliaryTags
}

// getters for HttpServerConfig struct

func (c HttpServerConfig) Enabled() bool {
	return c.enabled
}

func (c HttpServerConfig) Bind() string {
	return c.bind
}

func (c HttpServerConfig) Port() int {
	return c.port
}

func (c HttpServerConfig) LogRequests() bool {
	return c.logRequests
}

// getters for LocalDbConfig struct

func (c LocalDbConfig) Enabled() bool {
	return c.enabled
}

func (c LocalDbConfig) Path() string {
	return c.path
}

// getters for StatisticsConfig struct

func (c StatisticsConfig) Enabled() bool {
	return c.enabled
}

func (c StatisticsConfig) HistoryResolution() time.Duration {
	return c.historyResolution
}

func (c StatisticsConfig) HistoryMaxAge() time.Duration {
	return c.historyMaxAge
}

// getters for MqttClientConfig struct

func (c MqttClientConfig) Name() string {
	return c.name
}

// Broker always returns a non-nil pointer.
func (c MqttClientConfig) Broker() *url.URL {
	return c.broker
}

func (c MqttClientConfig) ProtocolVersion() int {
	return c.protocolVersion
}

func (c MqttClientConfig) User() string {
	return c.user
}

func (c MqttClientConfig) Password() string {
	return c.password
}

func (c MqttClientConfig) ClientId() string {
	return c.clientId
}

func (c MqttClientConfig) Qos() byte {
	return c.qos
}

func (c MqttClientConfig) KeepAlive() time.Duration {
	return c.keepAlive
}

func (c MqttClientConfig) ConnectRetryDelay() time.Duration {
	return c.connectRetryDelay
}

func (c MqttClientConfig) ConnectTimeout() time.Duration {
	return c.connectTimeout
}

func (c MqttClientConfig) AvailabilityTopic() string {
	return c.availabilityTopic
}

func (c MqttClientConfig) TopicPrefix() string {
	return c.topicPrefix
}

func (c MqttClientConfig) LogDebug() bool {
	return c.logDebug
}

func (c MqttClientConfig) LogMessages() bool {
	return c.logMessages
}

// getters for InfluxClientConfig struct

func (c InfluxClientConfig) Name() string {
	return c.name
}

func (c InfluxClientConfig) Url() string {
	return c.url
}

func (c InfluxClientConfig) Token() string {
	return c.token
}

func (c InfluxClientConfig) Org() string {
	return c.org
}

func (c InfluxClientConfig) Bucket() string {
	return c.bucket
}

func (c InfluxClientConfig) WriteInterval() time.Duration {
	return c.writeInterval
}

func (c InfluxClientConfig) RetryInterval() time.Duration {
	return c.retryInterval
}

func (c InfluxClientConfig) AggregateInterval() time.Duration {
	return c.aggregateInterval
}

func (c InfluxClientConfig) TimePrecision() time.Duration {
	return c.timePrecision
}

func (c InfluxClientConfig) ConnectTimeout() time.Duration {
	return c.connectTimeout
}

func (c InfluxClientConfig) BatchSize() uint {
	return c.batchSize
}

func (c InfluxClientConfig) RetryQueueLimit() uint {
	return c.retryQueueLimit
}

func (c InfluxClientConfig) LogDebug() bool {
	return c.logDebug
}

// getters for ConverterConfig struct

func (c ConverterConfig) Name() string {
	return c.name
}

func (c ConverterConfig) Implementation() string {
	return c.implementation
}

func (c ConverterConfig) MqttTopics() []*MqttTopicConfig {
	return c.mqttTopics
}

func (c ConverterConfig) MqttClients() []string {
	return c.mqttClients
}

func (c ConverterConfig) InfluxClients() []string {
	return c.influxClients
}

func (c ConverterConfig) LogHandleOnce() bool {
	return c.logHandleOnce
}

func (c ConverterConfig) LogDebug() bool {
	return c.logDebug
}

// getters for MqttTopicConfig struct

func (c MqttTopicConfig) Topic() string {
	return c.topic
}

func (c MqttTopicConfig) Device() string {
	return c.device
}

var deviceDynamicMatcher = regexp.MustCompile(`\+`)

func (c MqttTopicConfig) DeviceIsDynamic() bool {
	return deviceDynamicMatcher.MatchString(c.device)
}

// getters for InfluxAuxiliaryTags struct

func (c InfluxAuxiliaryTags) Tag() string {
	return c.tag
}

func (c InfluxAuxiliaryTags) Equals() *string {
	return c.equals
}

func (c InfluxAuxiliaryTags) Matches() *string {
	return c.matches
}

func (c InfluxAuxiliaryTags) MatchString(value string) bool {
	if c.equals != nil {
		return *c.equals == value
	}
	if c.matcher != nil {
		return c.matcher.MatchString(value)
	}

	return false
}

func (c InfluxAuxiliaryTags) TagValues() map[string]string {
	return c.tagValues
}
