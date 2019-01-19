package config

import "time"

func (c MqttClientConfig) Name() string {
	return c.name
}

func (c MqttClientConfig) Broker() string {
	return c.broker
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

func (c MqttClientConfig) AvailabilityTopic() string {
	return c.availabilityTopic
}

func (c MqttClientConfig) TopicPrefix() string {
	return c.topicPrefix
}

func (c MqttClientConfig) LogMessages() bool {
	return c.logMessages
}

func (c InfluxClientConfig) Name() string {
	return c.name
}

func (c InfluxClientConfig) Address() string {
	return c.address
}

func (c InfluxClientConfig) User() string {
	return c.user
}

func (c InfluxClientConfig) Password() string {
	return c.password
}

func (c InfluxClientConfig) Database() string {
	return c.database
}

func (c InfluxClientConfig) WriteInterval() time.Duration {
	return c.writeInterval
}

func (c InfluxClientConfig) TimePrecision() time.Duration {
	return c.timePrecision
}

func (c InfluxClientConfig) LogLineProtocol() bool {
	return c.logLineProtocol
}

func (c ConverterConfig) Name() string {
	return c.name
}

func (c ConverterConfig) Implementation() string {
	return c.implementation
}

func (c ConverterConfig) TargetMeasurement() string {
	return c.targetMeasurement
}

func (c ConverterConfig) MqttTopics() []string {
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

func (c StatisticsConfig) Enabled() bool {
	return c.enabled
}

func (c StatisticsConfig) HistoryResolution() time.Duration {
	return c.historyResolution
}

func (c StatisticsConfig) HistoryMaxAge() time.Duration {
	return c.historyMaxAge
}
