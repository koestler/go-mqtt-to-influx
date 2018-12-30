package config

func (c MqttClientConfig) MarshalYAML() (interface{}, error) {
	return mqttClientConfigRead{
		Broker:            c.broker,
		User:              c.user,
		Password:          c.password,
		ClientId:          c.clientId,
		Qos:               &c.qos,
		AvailabilityTopic: &c.availabilityTopic,
		TopicPrefix:       c.topicPrefix,
		LogMessages:       &c.logMessages,
	}, nil
}

func (c InfluxDbClientConfig) MarshalYAML() (interface{}, error) {
	return influxDbClientConfigRead{
		Address:         c.address,
		User:            c.user,
		Password:        c.password,
		Database:        c.database,
		WriteInterval:   c.writeInterval.String(),
		TimePrecision:   c.timePrecision.String(),
		LogLineProtocol: &c.logLineProtocol,
	}, nil
}

func (c ConverterConfig) MarshalYAML() (interface{}, error) {
	return converterConfigRead{
		Implementation:    c.implementation,
		TargetMeasurement: c.targetMeasurement,
		MqttTopics:        c.mqttTopics,
		MqttClients:       c.mqttClients,
		LogHandleOnce:     &c.logHandleOnce,
	}, nil
}

func (c HttpServerConfig) MarshalYAML() (interface{}, error) {
	return httpServerConfigRead{
		Bind:        c.bind,
		Port:        &c.port,
		LogRequests: &c.logRequests,
	}, nil
}

func (c StatisticsConfig) MarshalYAML() (interface{}, error) {
	return statisticsConfigRead{
		Enabled: &c.enabled,
	}, nil
}
