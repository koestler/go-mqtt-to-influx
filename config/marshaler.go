package config

func (c Config) MarshalYAML() (interface{}, error) {
	return configRead{
		Version: &c.Version,
		MqttClients: func() mqttClientConfigReadMap {
			mqttClients := make(mqttClientConfigReadMap, len(c.MqttClients))
			for _, c := range c.MqttClients {
				mqttClients[c.Name()] = c.convertToRead()
			}
			return mqttClients
		}(),
		InfluxDbClients: func() influxDbClientConfigReadMap {
			influxDbClients := make(influxDbClientConfigReadMap, len(c.InfluxDbClients))
			for _, c := range c.InfluxDbClients {
				influxDbClients[c.Name()] = c.convertToRead()
			}
			return influxDbClients
		}(),
		Converters: func() converterConfigReadMap {
			converters := make(converterConfigReadMap, len(c.Converters))
			for _, c := range c.Converters {
				converters[c.Name()] = c.convertToRead()
			}
			return converters
		}(),
		HttpServer: func() *httpServerConfigRead {
			if !c.HttpServer.Enabled() {
				return nil
			}
			r := c.HttpServer.convertToRead()
			return &r
		}(),
		Statistics: func() *statisticsConfigRead {
			if !c.Statistics.Enabled() {
				return nil
			}
			r := c.Statistics.convertToRead()
			return &r
		}(),
		LogConfig:      &c.LogConfig,
		LogWorkerStart: &c.LogWorkerStart,
		LogMqttDebug:   &c.LogMqttDebug,
	}, nil
}

func (c MqttClientConfig) convertToRead() mqttClientConfigRead {
	return mqttClientConfigRead{
		Broker:            c.broker,
		User:              c.user,
		Password:          c.password,
		ClientId:          c.clientId,
		Qos:               &c.qos,
		AvailabilityTopic: &c.availabilityTopic,
		TopicPrefix:       c.topicPrefix,
		LogMessages:       &c.logMessages,
	}
}

func (c InfluxDbClientConfig) convertToRead() influxDbClientConfigRead {
	return influxDbClientConfigRead{
		Address:         c.address,
		User:            c.user,
		Password:        c.password,
		Database:        c.database,
		WriteInterval:   c.writeInterval.String(),
		TimePrecision:   c.timePrecision.String(),
		LogLineProtocol: &c.logLineProtocol,
	}
}

func (c ConverterConfig) convertToRead() converterConfigRead {
	return converterConfigRead{
		Implementation:    c.implementation,
		TargetMeasurement: c.targetMeasurement,
		MqttTopics:        c.mqttTopics,
		MqttClients:       c.mqttClients,
		LogHandleOnce:     &c.logHandleOnce,
	}
}

func (c HttpServerConfig) convertToRead() httpServerConfigRead {
	return httpServerConfigRead{
		Bind:        c.bind,
		Port:        &c.port,
		LogRequests: &c.logRequests,
	}
}

func (c StatisticsConfig) convertToRead() statisticsConfigRead {
	return statisticsConfigRead{
		Enabled:           &c.enabled,
		HistoryResolution: c.HistoryResolution().String(),
		HistoryMaxAge:     c.HistoryMaxAge().String(),
	}
}
