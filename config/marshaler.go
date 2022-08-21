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
		InfluxClients: func() influxClientConfigReadMap {
			influxClients := make(influxClientConfigReadMap, len(c.InfluxClients))
			for _, c := range c.InfluxClients {
				influxClients[c.Name()] = c.convertToRead()
			}
			return influxClients
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

func (c InfluxClientConfig) convertToRead() influxClientConfigRead {
	return influxClientConfigRead{
		Url:           c.url,
		Token:         c.token,
		Org:           c.org,
		Bucket:        c.bucket,
		WriteInterval: c.writeInterval.String(),
		TimePrecision: c.timePrecision.String(),
		LogDebug:      &c.logDebug,
	}
}

func (c ConverterConfig) convertToRead() converterConfigRead {
	return converterConfigRead{
		Implementation: c.implementation,
		MqttTopics: func() []mqttTopicConfigRead {
			ret := make([]mqttTopicConfigRead, len(c.mqttTopics))
			for i, t := range c.mqttTopics {
				ret[i] = t.convertToRead()
			}
			return ret
		}(),
		MqttClients:   c.mqttClients,
		LogHandleOnce: &c.logHandleOnce,
	}
}

func (c MqttTopicConfig) convertToRead() mqttTopicConfigRead {
	return mqttTopicConfigRead{
		Topic:  c.topic,
		Device: &c.device,
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
