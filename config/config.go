package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"
)

const NameRegexp = "^[a-zA-Z0-9\\-]{1,32}$"

var nameMatcher = regexp.MustCompile(NameRegexp)

func ReadConfigFile(exe, source string) (config Config, err []error) {
	yamlStr, e := ioutil.ReadFile(source)
	if e != nil {
		return config, []error{fmt.Errorf("cannot read configuration: %v; use see `%s --help`", err, exe)}
	}

	return ReadConfig(yamlStr)
}

func ReadConfig(yamlStr []byte) (config Config, err []error) {
	var configRead configRead

	e := yaml.Unmarshal(yamlStr, &configRead)
	if e != nil {
		return config, []error{fmt.Errorf("cannot parse yaml: %s", err)}
	}

	return configRead.TransformAndValidate()
}

func (config Config) PrintConfig() (err error) {
	newYamlStr, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("cannot encode yaml again: %s", err)
	}

	log.Print("config: use the following config:")
	for _, line := range strings.Split(string(newYamlStr), "\n") {
		log.Print("config: ", line)
	}
	return nil
}

func (i configRead) TransformAndValidate() (ret Config, err []error) {
	var e []error
	ret.MqttClients, e = i.MqttClients.TransformAndValidate()
	err = append(err, e...)

	ret.InfluxDbClients, e = i.InfluxDbClients.TransformAndValidate()
	err = append(err, e...)

	ret.Converters, e = i.Converters.TransformAndValidate(ret.MqttClients, ret.InfluxDbClients)
	err = append(err, e...)

	ret.HttpServer, e = i.HttpServer.TransformAndValidate()
	err = append(err, e...)

	ret.Statistics, e = i.Statistics.TransformAndValidate()
	err = append(err, e...)

	if i.Version == nil {
		err = append(err, fmt.Errorf("Version must be defined; use Version=0"))
	} else {
		ret.Version = *i.Version
		if ret.Version != 0 {
			err = append(err, fmt.Errorf("Version=%d not supported", ret.Version))
		}
	}

	if i.LogConfig != nil && *i.LogConfig {
		ret.LogConfig = true
	}

	if i.LogWorkerStart != nil && *i.LogWorkerStart {
		ret.LogWorkerStart = true
	}

	if i.LogMqttDebug != nil && *i.LogMqttDebug {
		ret.LogMqttDebug = true
	}

	return
}

func (i *httpServerConfigRead) TransformAndValidate() (ret HttpServerConfig, err []error) {
	ret.enabled = false
	ret.bind = "[::1]"
	ret.port = 8042

	if i == nil {
		return
	}

	ret.enabled = true

	if len(i.Bind) > 0 {
		ret.bind = i.Bind
	}

	if i.Port != nil {
		ret.port = *i.Port
	}

	if i.LogRequests != nil && *i.LogRequests {
		ret.logRequests = true
	}

	return
}

func (i *statisticsConfigRead) TransformAndValidate() (ret StatisticsConfig, err []error) {
	ret.Enabled = false
	if i == nil {
		return
	}

	ret.Enabled = true
	return
}

func (m mqttClientConfigReadMap) getOrderedKeys() (ret []string) {
	ret = make([]string, len(m))
	i := 0
	for k, _ := range (m) {
		ret[i] = k
		i++
	}
	sort.Strings(ret)
	return
}

func (i mqttClientConfigReadMap) TransformAndValidate() (ret []MqttClientConfig, err []error) {
	if len(i) < 1 {
		return ret, []error{fmt.Errorf("MqttClients section must no be empty")}
	}

	ret = make([]MqttClientConfig, len(i))
	j := 0
	for _, name := range i.getOrderedKeys() {
		var e []error
		ret[j], e = i[name].TransformAndValidate(name)
		err = append(err, e...)
		j++
	}
	return
}

func (i mqttClientConfigRead) TransformAndValidate(name string) (ret MqttClientConfig, err []error) {
	ret = MqttClientConfig{
		Name:        name,
		Broker:      i.Broker,
		User:        i.User,
		Password:    i.Password,
		ClientId:    i.ClientId,
		TopicPrefix: i.TopicPrefix,
	}

	if !nameMatcher.MatchString(ret.Name) {
		err = append(err, fmt.Errorf("MqttClientConfig->Name='%s' does not match %s", ret.Name, NameRegexp))
	}

	if len(ret.Broker) < 1 {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->Broker must not be empty", name))
	}
	if len(ret.ClientId) < 1 {
		ret.ClientId = "go-mqtt-to-influxdb"
	}
	if i.Qos == nil {
		ret.Qos = 0
	} else if *i.Qos == 0 || *i.Qos == 1 || *i.Qos == 2 {
		ret.Qos = *i.Qos
	} else {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->Qos=%d but must be 0, 1 or 2", name, *i.Qos))
	}

	if i.AvailabilityTopic == nil {
		// use default
		ret.AvailabilityTopic = "%Prefix%tele/%ClientId%/LWT"
	} else {
		ret.AvailabilityTopic = *i.AvailabilityTopic
	}

	if i.LogMessages != nil && *i.LogMessages {
		ret.LogMessages = true
	}

	return
}

func (m influxDbClientConfigReadMap) getOrderedKeys() (ret []string) {
	ret = make([]string, len(m))
	i := 0
	for k, _ := range (m) {
		ret[i] = k
		i++
	}
	sort.Strings(ret)
	return
}

func (i influxDbClientConfigReadMap) TransformAndValidate() (ret []InfluxDbClientConfig, err []error) {
	if len(i) < 1 {
		return ret, []error{fmt.Errorf("InfluxDbClients section must no be empty")}
	}

	ret = make([]InfluxDbClientConfig, len(i))
	j := 0
	for _, name := range i.getOrderedKeys() {
		var e []error
		ret[j], e = i[name].TransformAndValidate(name)
		err = append(err, e...)
		j++
	}
	return
}

func (i influxDbClientConfigRead) TransformAndValidate(name string) (ret InfluxDbClientConfig, err []error) {
	ret = InfluxDbClientConfig{
		Name:     name,
		Address:  i.Address,
		User:     i.User,
		Password: i.Password,
		Database: i.Database,
	}

	if !nameMatcher.MatchString(ret.Name) {
		err = append(err, fmt.Errorf("InfluxDbClientConfig->Name='%s' does not match %s", ret.Name, NameRegexp))
	}

	if len(ret.Address) < 1 {
		err = append(err, fmt.Errorf("InfluxDbClientConfig->%s->Address must not be empty", name))
	}

	if len(ret.Database) < 1 {
		ret.Database = "go-mqtt-to-influxdb"
	}

	if len(i.WriteInterval) < 1 {
		// use default 0
		ret.WriteInterval = 200 * time.Millisecond
	} else if writeInterval, e := time.ParseDuration(i.WriteInterval); e != nil {
		err = append(err, fmt.Errorf("InfluxDbClientConfig->%s->WriteInterval='%s' parse error: %s",
			name, i.WriteInterval, e,
		))
	} else if writeInterval < 0 {
		err = append(err, fmt.Errorf("InfluxDbClientConfig->%s->WriteInterval='%s' must be positive",
			name, i.WriteInterval,
		))
	} else {
		ret.WriteInterval = writeInterval
	}

	if len(i.TimePrecision) < 1 {
		// use default 1s
		ret.TimePrecision = time.Second
	} else if timePrecision, e := time.ParseDuration(i.TimePrecision); e != nil {
		err = append(err, fmt.Errorf("InfluxDbClientConfig->%s->TimePrecision='%s' parse error: %s",
			name, i.TimePrecision, e,
		))
	} else if timePrecision < 0 {
		err = append(err, fmt.Errorf("InfluxDbClientConfig->%s->TimePrecision='%s' must be positive",
			name, i.TimePrecision,
		))
	} else {
		ret.TimePrecision = timePrecision
	}

	if i.LogLineProtocol != nil && *i.LogLineProtocol {
		ret.LogLineProtocol = true
	}

	return
}

func (m converterReadMap) getOrderedKeys() (ret []string) {
	ret = make([]string, len(m))
	i := 0
	for k, _ := range (m) {
		ret[i] = k
		i++
	}
	sort.Strings(ret)
	return
}

func (i converterReadMap) TransformAndValidate(
	mqttClients []MqttClientConfig,
	influxDbClients []InfluxDbClientConfig,
) (ret []ConverterConfig, err []error) {
	if len(i) < 1 {
		return ret, []error{fmt.Errorf("Converters section must no be empty")}
	}

	ret = make([]ConverterConfig, len(i))
	j := 0
	for _, name := range i.getOrderedKeys() {
		var e []error
		ret[j], e = i[name].TransformAndValidate(name, mqttClients, influxDbClients)
		err = append(err, e...)
		j++
	}
	return
}

var implementationsAndDefaultMeasurement = map[string]string{
	"go-ve-sensor":   "floatValue",
	"lwt":            "boolValue",
	"tasmota-state":  "boolValue",
	"tasmota-sensor": "floatValue",
}

func (i converterConfigRead) TransformAndValidate(
	name string,
	mqttClients []MqttClientConfig,
	influxDbClients []InfluxDbClientConfig,
) (ret ConverterConfig, err []error) {
	ret = ConverterConfig{
		Name:              name,
		Implementation:    i.Implementation,
		TargetMeasurement: i.TargetMeasurement,
		MqttTopics:        i.MqttTopics,
		MqttClients:       i.MqttClients,
		InfluxDbClients:   i.InfluxDbClients,
	}

	if !nameMatcher.MatchString(ret.Name) {
		err = append(err, fmt.Errorf("Converters->Name='%s' does not match %s", ret.Name, NameRegexp))
	}

	if def, ok := implementationsAndDefaultMeasurement[ret.Implementation]; !ok {
		err = append(err, fmt.Errorf("Converters->%s->Implementation='%s' is unkown", name, ret.Implementation))
	} else if len(ret.TargetMeasurement) < 1 {
		ret.TargetMeasurement = def
	}

	// validate that all listed MqttClients exist
	for _, clientName := range ret.MqttClients {
		found := false
		for _, client := range mqttClients {
			if clientName == client.Name {
				found = true
				break
			}
		}

		if !found {
			err = append(err, fmt.Errorf("Converters->%s->MqttClient='%s' is not defined", name, clientName))
		}
	}

	// validate that all listed InfluxDbClients exist
	for _, clientName := range ret.InfluxDbClients {
		found := false
		for _, client := range influxDbClients {
			if clientName == client.Name {
				found = true
				break
			}
		}

		if !found {
			err = append(err, fmt.Errorf("Converters->%s->InfluxDbClient='%s' is not defined", name, clientName))
		}
	}

	if len(ret.MqttTopics) < 1 {
		err = append(err, fmt.Errorf("Converters->%s->MqttTopics must not be empty", name))
	}

	if i.LogHandleOnce != nil && *i.LogHandleOnce {
		ret.LogHandleOnce = true
	}

	return
}
