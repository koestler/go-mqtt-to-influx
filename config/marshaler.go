package config

func (c Config) MarshalYAML() (interface{}, error) {
	return configRead{
		Version: &c.version,
		HttpServer: func() *httpServerConfigRead {
			if !c.httpServer.Enabled() {
				return nil
			}
			r := c.httpServer.convertToRead()
			return &r
		}(),
		LocalDb: func() *localDbConfigRead {
			if !c.localDb.Enabled() {
				return nil
			}
			r := c.localDb.convertToRead()
			return &r
		}(),
		Statistics: func() *statisticsConfigRead {
			if !c.statistics.Enabled() {
				return nil
			}
			r := c.statistics.convertToRead()
			return &r
		}(),
		LogConfig:      &c.logConfig,
		LogWorkerStart: &c.logWorkerStart,
		MqttClients: func() mqttClientConfigReadMap {
			mqttClients := make(mqttClientConfigReadMap, len(c.mqttClients))
			for _, c := range c.mqttClients {
				mqttClients[c.Name()] = c.convertToRead()
			}
			return mqttClients
		}(),
		InfluxClients: func() influxClientConfigReadMap {
			influxClients := make(influxClientConfigReadMap, len(c.influxClients))
			for _, c := range c.influxClients {
				influxClients[c.Name()] = c.convertToRead()
			}
			return influxClients
		}(),
		Converters: func() converterConfigReadMap {
			converters := make(converterConfigReadMap, len(c.converters))
			for _, c := range c.converters {
				converters[c.Name()] = c.convertToRead()
			}
			return converters
		}(),
		InfluxAuxiliaryTags: func() []influxAuxiliaryTagsRead {
			influxAuxiliaryTags := make([]influxAuxiliaryTagsRead, len(c.influxAuxiliaryTags))
			for i, c := range c.influxAuxiliaryTags {
				influxAuxiliaryTags[i] = c.convertToRead()
			}
			return influxAuxiliaryTags
		}(),
	}, nil
}

func (c HttpServerConfig) convertToRead() httpServerConfigRead {
	return httpServerConfigRead{
		Bind:        c.bind,
		Port:        &c.port,
		LogRequests: &c.logRequests,
	}
}

func (c LocalDbConfig) convertToRead() localDbConfigRead {
	return localDbConfigRead{
		Path: &c.path,
	}
}

func (c StatisticsConfig) convertToRead() statisticsConfigRead {
	return statisticsConfigRead{
		HistoryResolution: c.HistoryResolution().String(),
		HistoryMaxAge:     c.HistoryMaxAge().String(),
	}
}

func (c MqttClientConfig) convertToRead() mqttClientConfigRead {
	return mqttClientConfigRead{
		Broker:            c.broker.String(),
		ProtocolVersion:   &c.protocolVersion,
		User:              c.user,
		Password:          c.password,
		ClientId:          &c.clientId,
		Qos:               &c.qos,
		KeepAlive:         c.keepAlive.String(),
		ConnectRetryDelay: c.connectRetryDelay.String(),
		ConnectTimeout:    c.connectTimeout.String(),
		AvailabilityTopic: &c.availabilityTopic,
		TopicPrefix:       c.topicPrefix,
		LogDebug:          &c.logDebug,
		LogMessages:       &c.logMessages,
	}
}

func (c InfluxClientConfig) convertToRead() influxClientConfigRead {
	return influxClientConfigRead{
		Url:               c.url,
		Token:             c.token,
		Org:               c.org,
		Bucket:            c.bucket,
		WriteInterval:     c.writeInterval.String(),
		RetryInterval:     c.retryInterval.String(),
		AggregateInterval: c.aggregateInterval.String(),
		TimePrecision:     c.timePrecision.String(),
		ConnectTimeout:    c.connectTimeout.String(),
		BatchSize:         &c.batchSize,
		RetryQueueLimit:   &c.retryQueueLimit,
		LogDebug:          &c.logDebug,
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
		LogDebug:      &c.logDebug,
	}
}

func (c MqttTopicConfig) convertToRead() mqttTopicConfigRead {
	return mqttTopicConfigRead{
		Topic:  c.topic,
		Device: &c.device,
	}
}

func (c InfluxAuxiliaryTags) convertToRead() influxAuxiliaryTagsRead {
	return influxAuxiliaryTagsRead{
		Tag:       &c.tag,
		Equals:    c.equals,
		Matches:   c.matches,
		TagValues: c.tagValues,
	}
}
