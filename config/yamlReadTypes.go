package config

// The following structs are used to read the configuration from the yml file.

type configRead struct {
	Version             *int                        `yaml:"Version"`
	HttpServer          *httpServerConfigRead       `yaml:"HttpServer"`
	LocalDb             *localDbConfigRead          `yaml:"LocalDb"`
	Statistics          *statisticsConfigRead       `yaml:"Statistics"`
	LogConfig           *bool                       `yaml:"LogConfig"`
	LogWorkerStart      *bool                       `yaml:"LogWorkerStart"`
	MqttClients         mqttClientConfigReadMap     `yaml:"MqttClients"`
	InfluxClients       influxClientConfigReadMap   `yaml:"InfluxClients"`
	Converters          converterConfigReadMap      `yaml:"Converters"`
	InfluxAuxiliaryTags influxAuxiliaryTagsReadList `yaml:"InfluxAuxiliaryTags"`
}

type httpServerConfigRead struct {
	Bind        string `yaml:"Bind"`
	Port        *int   `yaml:"Port"`
	LogRequests *bool  `yaml:"LogRequests"`
}

type localDbConfigRead struct {
	Path *string `yaml:"Path"`
}

type statisticsConfigRead struct {
	HistoryResolution string `yaml:"HistoryResolution"`
	HistoryMaxAge     string `yaml:"HistoryMaxAge"`
}

type mqttClientConfigRead struct {
	Broker            string  `yaml:"Broker"`
	ProtocolVersion   *int    `yaml:"ProtocolVersion"`
	User              string  `yaml:"User"`
	Password          string  `yaml:"Password"`
	ClientId          *string `yaml:"ClientId"`
	Qos               *byte   `yaml:"Qos"`
	KeepAlive         string  `yaml:"KeepAlive"`
	ConnectRetryDelay string  `yaml:"ConnectRetryDelay"`
	ConnectTimeout    string  `yaml:"ConnectTimeout"`
	AvailabilityTopic *string `yaml:"AvailabilityTopic"`
	TopicPrefix       string  `yaml:"TopicPrefix"`
	LogDebug          *bool   `yaml:"LogDebug"`
	LogMessages       *bool   `yaml:"LogMessages"`
}

type mqttClientConfigReadMap map[string]mqttClientConfigRead

type influxClientConfigRead struct {
	Url               string `yaml:"Url"`
	Token             string `yaml:"Token"`
	Org               string `yaml:"Org"`
	Bucket            string `yaml:"Bucket"`
	WriteInterval     string `yaml:"WriteInterval"`
	RetryInterval     string `yaml:"RetryInterval"`
	AggregateInterval string `yaml:"AggregateInterval"`
	TimePrecision     string `yaml:"TimePrecision"`
	ConnectTimeout    string `yaml:"ConnectTimeout"`
	BatchSize         *uint  `yaml:"BatchSize"`
	RetryQueueLimit   *uint  `yaml:"RetryQueueLimit"`
	LogDebug          *bool  `yaml:"LogDebug"`
}

type influxClientConfigReadMap map[string]influxClientConfigRead

type converterConfigRead struct {
	Implementation string                  `yaml:"Implementation"`
	MqttTopics     mqttTopicConfigReadList `yaml:"MqttTopics"`
	MqttClients    []string                `yaml:"MqttClients"`
	InfluxClients  []string                `yaml:"InfluxClients"`
	LogHandleOnce  *bool                   `yaml:"LogHandleOnce"`
	LogDebug       *bool                   `yaml:"LogDebug"`
}

type converterConfigReadMap map[string]converterConfigRead

type mqttTopicConfigRead struct {
	Topic  string  `yaml:"Topic"`
	Device *string `yaml:"Device"`
}

type mqttTopicConfigReadList []mqttTopicConfigRead

type influxAuxiliaryTagsRead struct {
	Tag       *string           `yaml:"Tag"`
	Equals    *string           `yaml:"Equals"`
	Matches   *string           `yaml:"Matches"`
	TagValues map[string]string `yaml:"TagValues"`
}

type influxAuxiliaryTagsReadList []influxAuxiliaryTagsRead
