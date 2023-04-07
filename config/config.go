package config

import (
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"log"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const NameRegexp = "^[a-zA-Z0-9\\-]{1,32}$"

var nameMatcher = regexp.MustCompile(NameRegexp)

func ReadConfigFile(exe, source string) (config Config, err []error) {
	yamlStr, e := os.ReadFile(source)
	if e != nil {
		return config, []error{fmt.Errorf("cannot read configuration: %v; use see `%s --help`", err, exe)}
	}

	return ReadConfig(yamlStr)
}

func ReadConfig(yamlStr []byte) (config Config, err []error) {
	var configRead configRead

	yamlStr = []byte(os.ExpandEnv(string(yamlStr)))
	e := yaml.Unmarshal(yamlStr, &configRead)
	if e != nil {
		return config, []error{fmt.Errorf("cannot parse yaml: %s", e)}
	}

	return configRead.TransformAndValidate()
}

func (c Config) PrintConfig() (err error) {
	newYamlStr, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("cannot encode yaml again: %s", err)
	}

	log.Print("config: use the following config:")
	for _, line := range strings.Split(string(newYamlStr), "\n") {
		log.Print("config: ", line)
	}
	return nil
}

func (c configRead) TransformAndValidate() (ret Config, err []error) {
	if c.Version == nil {
		err = append(err, fmt.Errorf("version must be defined. Use Version=0"))
	} else {
		ret.version = *c.Version
		if ret.version != 0 {
			err = append(err, fmt.Errorf("version=%d is not supported", ret.Version()))
		}
	}

	var e []error

	ret.httpServer, e = c.HttpServer.TransformAndValidate()
	err = append(err, e...)

	ret.localDb, e = c.LocalDb.TransformAndValidate()
	err = append(err, e...)

	ret.statistics, e = c.Statistics.TransformAndValidate()
	err = append(err, e...)

	if c.LogConfig != nil && *c.LogConfig {
		ret.logConfig = true
	}

	if c.LogWorkerStart != nil && *c.LogWorkerStart {
		ret.logWorkerStart = true
	}

	ret.mqttClients, e = c.MqttClients.TransformAndValidate()
	err = append(err, e...)

	ret.influxClients, e = c.InfluxClients.TransformAndValidate()
	err = append(err, e...)

	ret.converters, e = c.Converters.TransformAndValidate(ret.mqttClients, ret.influxClients)
	err = append(err, e...)

	ret.influxAuxiliaryTags, e = c.InfluxAuxiliaryTags.TransformAndValidate()
	err = append(err, e...)

	return
}

func (c *httpServerConfigRead) TransformAndValidate() (ret HttpServerConfig, err []error) {
	ret.enabled = false
	ret.bind = "[::1]"
	ret.port = 8000

	if c == nil {
		return
	}

	ret.enabled = true

	if len(c.Bind) > 0 {
		ret.bind = c.Bind
	}

	if c.Port != nil {
		ret.port = *c.Port
	}

	if c.LogRequests != nil && *c.LogRequests {
		ret.logRequests = true
	}

	return
}

func (c *localDbConfigRead) TransformAndValidate() (ret LocalDbConfig, err []error) {
	// default values
	ret.enabled = false
	ret.path = "./go-mqtt-to-influx.db"

	if c == nil {
		return
	}

	ret.enabled = true

	if c.Path != nil {
		ret.path = *c.Path
	}

	return
}

func (c *statisticsConfigRead) TransformAndValidate() (ret StatisticsConfig, err []error) {
	// default values
	ret.enabled = false
	ret.historyResolution = 10 * time.Second
	ret.historyMaxAge = 10 * time.Minute

	if c == nil {
		return
	}

	ret.enabled = true

	if len(c.HistoryResolution) < 1 {
		// use default 1s
	} else if historyResolution, e := time.ParseDuration(c.HistoryResolution); e != nil {
		err = append(err, fmt.Errorf("Statistics->HistoryResolution='%s' parse error: %s",
			c.HistoryResolution, e,
		))
	} else if historyResolution <= 0 {
		err = append(err, fmt.Errorf("Statistics->HistoryResolution='%s' must be >0",
			c.HistoryResolution,
		))
	} else {
		ret.historyResolution = historyResolution
	}

	if len(c.HistoryMaxAge) < 1 {
		// use default 10min
	} else if historyMaxAge, e := time.ParseDuration(c.HistoryMaxAge); e != nil {
		err = append(err, fmt.Errorf("Statistics->HistoryMaxAge='%s' parse error: %s",
			c.HistoryMaxAge, e,
		))
	} else if historyMaxAge <= 0 {
		err = append(err, fmt.Errorf("Statistics->HistoryMaxAge='%s' must be >0",
			c.HistoryMaxAge,
		))
	} else {
		ret.historyMaxAge = historyMaxAge
	}

	return
}

func (c mqttClientConfigReadMap) getOrderedKeys() (ret []string) {
	ret = make([]string, len(c))
	i := 0
	for k := range c {
		ret[i] = k
		i++
	}
	sort.Strings(ret)
	return
}

func (c mqttClientConfigReadMap) TransformAndValidate() (ret []*MqttClientConfig, err []error) {
	if len(c) < 1 {
		return ret, []error{fmt.Errorf("MqttClients section must no be empty")}
	}

	ret = make([]*MqttClientConfig, len(c))
	j := 0
	for _, name := range c.getOrderedKeys() {
		r, e := c[name].TransformAndValidate(name)
		ret[j] = &r
		err = append(err, e...)
		j++
	}

	return
}

func (c mqttClientConfigRead) TransformAndValidate(name string) (ret MqttClientConfig, err []error) {
	ret = MqttClientConfig{
		name:        name,
		user:        c.User,
		password:    c.Password,
		topicPrefix: c.TopicPrefix,
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("MqttClientConfig->Name='%s' does not match %s", ret.name, NameRegexp))
	}

	if len(c.Broker) < 1 {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->Broker must not be empty", name))
	} else {
		if broker, e := url.ParseRequestURI(c.Broker); e != nil {
			err = append(err, fmt.Errorf("MqttClientConfig->%s->Broker invalid url: %s", name, e))
		} else if broker == nil {
			err = append(err, fmt.Errorf("MqttClientConfig->%s->Broker cannot parse broker", name))
		} else {
			ret.broker = broker
		}
	}

	if c.ProtocolVersion == nil {
		ret.protocolVersion = 5
	} else if *c.ProtocolVersion == 3 || *c.ProtocolVersion == 5 {
		ret.protocolVersion = *c.ProtocolVersion
	} else {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->Protocol=%d but must be 3 or 5", name, *c.ProtocolVersion))
	}

	if c.ClientId == nil {
		ret.clientId = "go-mqtt-to-influx-" + uuid.New().String()
	} else {
		ret.clientId = *c.ClientId
	}

	if c.Qos == nil {
		ret.qos = 1 // default qos is 1
	} else if *c.Qos == 0 || *c.Qos == 1 || *c.Qos == 2 {
		ret.qos = *c.Qos
	} else {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->Qos=%d but must be 0, 1 or 2", name, *c.Qos))
	}

	if len(c.KeepAlive) < 1 {
		// use default 60s
		ret.keepAlive = time.Minute
	} else if keepAlive, e := time.ParseDuration(c.KeepAlive); e != nil {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->KeepAlive='%s' parse error: %s",
			name, c.KeepAlive, e,
		))
	} else if keepAlive < time.Second {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->KeepAlive='%s' must be >=1s",
			name, c.KeepAlive,
		))
	} else if keepAlive%time.Second != 0 {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->KeepAlive='%s' must be a multiple of a second",
			name, c.KeepAlive,
		))
	} else {
		ret.keepAlive = keepAlive
	}

	if len(c.ConnectRetryDelay) < 1 {
		// use default 10s
		ret.connectRetryDelay = 10 * time.Second
	} else if connectRetryDelay, e := time.ParseDuration(c.ConnectRetryDelay); e != nil {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->ConnectRetryDelay='%s' parse error: %s",
			name, c.ConnectRetryDelay, e,
		))
	} else if connectRetryDelay < 100*time.Millisecond {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->ConnectRetryDelay='%s' must be >=100ms",
			name, c.ConnectRetryDelay,
		))
	} else {
		ret.connectRetryDelay = connectRetryDelay
	}

	if len(c.ConnectTimeout) < 1 {
		// use default 5s
		ret.connectTimeout = 5 * time.Second
	} else if connectTimeout, e := time.ParseDuration(c.ConnectTimeout); e != nil {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->ConnectTimeout='%s' parse error: %s",
			name, c.ConnectTimeout, e,
		))
	} else if connectTimeout < 100*time.Millisecond {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->ConnectTimeout='%s' must be >=100ms",
			name, c.ConnectTimeout,
		))
	} else {
		ret.connectTimeout = connectTimeout
	}

	if c.AvailabilityTopic == nil {
		// use default
		ret.availabilityTopic = "%Prefix%tele/%ClientId%/status"
	} else {
		ret.availabilityTopic = *c.AvailabilityTopic
	}

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	if c.LogMessages != nil && *c.LogMessages {
		ret.logMessages = true
	}

	return
}

func (c influxClientConfigReadMap) getOrderedKeys() (ret []string) {
	ret = make([]string, len(c))
	i := 0
	for k := range c {
		ret[i] = k
		i++
	}
	sort.Strings(ret)
	return
}

func (c influxClientConfigReadMap) TransformAndValidate() (ret []*InfluxClientConfig, err []error) {
	if len(c) < 1 {
		return ret, []error{fmt.Errorf("InfluxClients section must no be empty")}
	}

	ret = make([]*InfluxClientConfig, len(c))
	j := 0
	for _, name := range c.getOrderedKeys() {
		r, e := c[name].TransformAndValidate(name)
		ret[j] = &r
		err = append(err, e...)
		j++
	}
	return
}

func (c influxClientConfigRead) TransformAndValidate(name string) (ret InfluxClientConfig, err []error) {
	ret = InfluxClientConfig{
		name:   name,
		url:    c.Url,
		token:  c.Token,
		org:    c.Org,
		bucket: c.Bucket,
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("InfluxClientConfig->Name='%s' does not match %s", ret.name, NameRegexp))
	}

	if len(ret.url) < 1 {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->Url must not be empty", name))
	}

	if len(ret.token) < 1 {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->Token must not be empty", name))
	}

	if len(ret.org) < 1 {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->Org must not be empty", name))
	}

	if len(ret.bucket) < 1 {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->Bucket must not be empty", name))
	}

	if len(c.WriteInterval) < 1 {
		// use default 10s
		ret.writeInterval = 10 * time.Second
	} else if writeInterval, e := time.ParseDuration(c.WriteInterval); e != nil {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->WriteInterval='%s' parse error: %s",
			name, c.WriteInterval, e,
		))
	} else if writeInterval < 0 {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->WriteInterval='%s' must be positive",
			name, c.WriteInterval,
		))
	} else {
		ret.writeInterval = writeInterval
	}

	if len(c.RetryInterval) < 1 {
		// use default 10s
		ret.retryInterval = 10 * time.Second
	} else if retryInterval, e := time.ParseDuration(c.RetryInterval); e != nil {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->RetryInterval='%s' parse error: %s",
			name, c.RetryInterval, e,
		))
	} else if retryInterval < 0 {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->RetryInterval='%s' must be positive",
			name, c.RetryInterval,
		))
	} else {
		ret.retryInterval = retryInterval
	}

	if len(c.AggregateInterval) < 1 {
		// use default 60s
		ret.aggregateInterval = time.Minute
	} else if aggregateInterval, e := time.ParseDuration(c.AggregateInterval); e != nil {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->AggregateInterval='%s' parse error: %s",
			name, c.AggregateInterval, e,
		))
	} else if aggregateInterval < 0 {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->AggregateInterval='%s' must be positive",
			name, c.AggregateInterval,
		))
	} else {
		ret.aggregateInterval = aggregateInterval
	}

	if len(c.TimePrecision) < 1 {
		// use default 1s
		ret.timePrecision = time.Second
	} else if timePrecision, e := time.ParseDuration(c.TimePrecision); e != nil {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->TimePrecision='%s' parse error: %s",
			name, c.TimePrecision, e,
		))
	} else if timePrecision < 0 {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->TimePrecision='%s' must be positive",
			name, c.TimePrecision,
		))
	} else {
		ret.timePrecision = timePrecision
	}

	if len(c.ConnectTimeout) < 1 {
		// use default 5s
		ret.connectTimeout = 5 * time.Second
	} else if connectTimeout, e := time.ParseDuration(c.ConnectTimeout); e != nil {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->ConnectTimeout='%s' parse error: %s",
			name, c.ConnectTimeout, e,
		))
	} else if connectTimeout < time.Second {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->ConnectTimeout='%s' must be >=1s",
			name, c.ConnectTimeout,
		))
	} else {
		ret.connectTimeout = connectTimeout
	}

	if c.BatchSize == nil {
		// use default 5000
		ret.batchSize = 5000
	} else if b := *c.BatchSize; b > 0 {
		ret.batchSize = b
	} else {
		err = append(err, fmt.Errorf("InfluxClientConfig->%s->BatchSize='%d' must be positive", name, b))
	}

	if c.RetryQueueLimit == nil {
		// use default 20
		ret.retryQueueLimit = 20
	} else {
		ret.retryQueueLimit = *c.RetryQueueLimit
	}

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	return
}

func (c converterConfigReadMap) getOrderedKeys() (ret []string) {
	ret = make([]string, len(c))
	i := 0
	for k := range c {
		ret[i] = k
		i++
	}
	sort.Strings(ret)
	return
}

func (c converterConfigReadMap) TransformAndValidate(
	mqttClients []*MqttClientConfig,
	influxClients []*InfluxClientConfig,
) (ret []*ConverterConfig, err []error) {
	if len(c) < 1 {
		return ret, []error{fmt.Errorf("converters section must no be empty")}
	}

	ret = make([]*ConverterConfig, len(c))
	j := 0
	for _, name := range c.getOrderedKeys() {
		r, e := c[name].TransformAndValidate(name, mqttClients, influxClients)
		ret[j] = &r
		err = append(err, e...)
		j++
	}
	return
}

func (c converterConfigRead) TransformAndValidate(
	name string,
	mqttClients []*MqttClientConfig,
	influxClients []*InfluxClientConfig,
) (ret ConverterConfig, err []error) {
	ret = ConverterConfig{
		name:           name,
		implementation: c.Implementation,
		mqttClients:    c.MqttClients,
		influxClients:  c.InfluxClients,
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("Converters->Name='%s' does not match %s", ret.name, NameRegexp))
	}

	if len(ret.implementation) < 1 {
		err = append(err, fmt.Errorf("Converters->%s->Implementation='%s' is unkown", name, ret.implementation))
	}

	var e []error
	ret.mqttTopics, e = c.MqttTopics.TransformAndValidate()
	err = append(err, e...)

	// validate that all listed mqttClients exist
	for _, clientName := range ret.mqttClients {
		found := false
		for _, client := range mqttClients {
			if clientName == client.name {
				found = true
				break
			}
		}

		if !found {
			err = append(err, fmt.Errorf("Converters->%s->MqttClient='%s' is not defined", name, clientName))
		}
	}

	// validate that all listed influxClients exist
	for _, clientName := range ret.influxClients {
		found := false
		for _, client := range influxClients {
			if clientName == client.name {
				found = true
				break
			}
		}

		if !found {
			err = append(err, fmt.Errorf("Converters->%s->InfluxClient='%s' is not defined", name, clientName))
		}
	}

	if len(ret.mqttTopics) < 1 {
		err = append(err, fmt.Errorf("Converters->%s->MqttTopics must not be empty", name))
	}

	if c.LogHandleOnce != nil && *c.LogHandleOnce {
		ret.logHandleOnce = true
	}

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	return
}

func (c mqttTopicConfigReadList) TransformAndValidate() (ret []*MqttTopicConfig, err []error) {
	if len(c) < 1 {
		return ret, []error{fmt.Errorf("mqttTopics section must no be empty")}
	}

	ret = make([]*MqttTopicConfig, len(c))
	for i, t := range c {
		r, e := t.TransformAndValidate()
		ret[i] = &r
		err = append(err, e...)
	}
	return
}

func (c mqttTopicConfigRead) TransformAndValidate() (ret MqttTopicConfig, err []error) {
	ret = MqttTopicConfig{
		topic: c.Topic,
	}

	if c.Device == nil {
		ret.device = "+"
	} else {
		ret.device = *c.Device
	}

	// is dynamic
	if deviceDynamicMatcher.MatchString(ret.device) {
		// we have a dynamic device name
		if !strings.Contains(ret.topic, "%Device%") {
			err = append(err, fmt.Errorf("topic '%s' must contain '%%Device%%'", ret.topic))
		}
	} else {
		// we have a static device name, name must not contain +
		if strings.Contains(ret.device, "+") {
			err = append(err, fmt.Errorf("invalid device=%s", ret.device))
		}
	}

	return
}

type ApplyTopicReplaceFunc func(string) string

func (c MqttTopicConfig) ApplyTopicReplace(f ApplyTopicReplaceFunc) MqttTopicConfig {
	return MqttTopicConfig{
		topic:  f(c.topic),
		device: c.device,
	}
}
func (c influxAuxiliaryTagsReadList) TransformAndValidate() (ret []*InfluxAuxiliaryTags, err []error) {
	ret = make([]*InfluxAuxiliaryTags, len(c))
	for i, t := range c {
		r, e := t.TransformAndValidate()
		ret[i] = &r
		err = append(err, e...)
	}
	return
}

func (c influxAuxiliaryTagsRead) TransformAndValidate() (ret InfluxAuxiliaryTags, err []error) {
	ret = InfluxAuxiliaryTags{
		equals:    c.Equals,
		matches:   c.Matches,
		tagValues: c.TagValues,
	}

	if c.Tag == nil {
		ret.tag = "device"
	} else if len(*c.Tag) < 1 {
		err = append(err, fmt.Errorf("InfluxAuxiliaryTags->TagValues Tag must not by empty"))
	} else {
		ret.tag = *c.Tag
	}

	if len(c.TagValues) < 1 {
		err = append(err, fmt.Errorf("InfluxAuxiliaryTags->TagValues must not be empty"))
	}

	if c.Equals != nil && c.Matches == nil {
		// ok
	} else if c.Equals == nil && c.Matches != nil {
		expr := *c.Matches
		if m, e := regexp.Compile(expr); e != nil {
			err = append(err, fmt.Errorf("InfluxAuxiliaryTags: invalid regexp given by Matches='%s': %s", expr, e))
		} else {
			ret.matcher = m
		}
	} else {
		err = append(err, fmt.Errorf("InfluxAuxiliaryTags Equals xor Matches must be set"))
		return
	}

	return
}
