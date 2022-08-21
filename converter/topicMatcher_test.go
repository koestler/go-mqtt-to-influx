package converter

//go:generate mockgen -destination=mock/topicMatcher_mock.go -package converter_mock github.com/koestler/go-mqtt-to-influx/converter TopicMatcherConfig
