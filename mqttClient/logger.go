package mqttClient

import (
	"log"
)

type logger struct {
	prefix string
}

func (l logger) Println(v ...interface{}) {
	log.Println(append([]interface{}{l.prefix}, v...)...)
}

func (l logger) Printf(format string, v ...interface{}) {
	if len(format) > 0 && format[len(format)-1] != '\n' {
		format = format + "\n" // some log calls in paho do not add \n
	}
	log.Printf(l.prefix+format, v...)
}
