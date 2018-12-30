package converter

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
)

type HandleFunc func(c Config, oup Output, msg mqtt.Message)

var converterImplementations = make(map[string]HandleFunc)

func registerHandler(implementation string, h HandleFunc) {
	if _, ok := converterImplementations[implementation]; ok {
		log.Fatalf("converter: implementation='%s' registered twice", implementation)
	}

	converterImplementations[implementation] = h
}

func getHandler(implementation string) (h HandleFunc, err error) {
	h, ok := converterImplementations[implementation]
	if !ok {
		return nil, fmt.Errorf("unknown implementation='%s'", implementation)
	}
	return h, nil
}