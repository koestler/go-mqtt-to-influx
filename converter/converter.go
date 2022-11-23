package converter

import (
	"fmt"
	"time"
)

type Config interface {
	Name() string
	Implementation() string
	InfluxClients() []string
	LogHandleOnce() bool
	LogDebug() bool
}

type Input interface {
	Topic() string
	Payload() []byte
}

type Output interface {
	Measurement() string
	Tags() map[string]string
	Fields() map[string]interface{}
	Time() time.Time
}

type OutputFunc func(output Output)
type HandleFunc func(c Config, topicMatcher TopicMatcher, input Input, outputFunc OutputFunc)

var converterImplementations = make(map[string]HandleFunc)

func registerHandler(implementation string, h HandleFunc) {
	if _, ok := converterImplementations[implementation]; ok {
		panic(fmt.Sprintf("converter: implementation='%s' registered twice; ignore second", implementation))
	}

	converterImplementations[implementation] = h
}

func GetHandler(implementation string) (h HandleFunc, err error) {
	h, ok := converterImplementations[implementation]
	if !ok {
		return nil, fmt.Errorf("unknown implementation='%s'", implementation)
	}
	return h, nil
}
