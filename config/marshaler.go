package config

func (c Config) MarshalYAML() (interface{}, error) {
	return configRead{
		Version: &c.version,
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
		InfluxAuxiliaryTags: func() []influxAuxiliaryTagsRead {
			influxAuxiliaryTags := make([]influxAuxiliaryTagsRead, len(c.influxAuxiliaryTags))
			for i, c := range c.influxAuxiliaryTags {
				influxAuxiliaryTags[i] = c.convertToRead()
			}
			return influxAuxiliaryTags
		}(),
		Converters: func() converterConfigReadMap {
			converters := make(converterConfigReadMap, len(c.converters))
			for _, c := range c.converters {
				converters[c.Name()] = c.convertToRead()
			}
			return converters
		}(),
		HttpServer: func() *httpServerConfigRead {
			if !c.httpServer.Enabled() {
				return nil
			}
			r := c.httpServer.convertToRead()
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
	}, nil
}

func (c MqttClientConfig) convertToRead() mqttClientConfigRead {
	return mqttClientConfigRead{
		Broker:            c.broker.String(),
		ProtocolVersion:   &c.protocolVersion,
		User:              c.user,
		Password:          c.password,
		ClientId:          c.clientId,
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
		Url:           c.url,
		Token:         c.token,
		Org:           c.org,
		Bucket:        c.bucket,
		WriteInterval: c.writeInterval.String(),
		TimePrecision: c.timePrecision.String(),
		LogDebug:      &c.logDebug,
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
