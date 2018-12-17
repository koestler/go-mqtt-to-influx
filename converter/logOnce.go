package converter

import (
	"fmt"
	"log"
)

var loggedTopics = make(map[string]bool)

func logTopicOnce(converter, topic string) {
	id := fmt.Sprintf("%s,%s", converter, topic)
	_, ok := loggedTopics[id]
	if !ok {
		log.Printf("converter[%s]: handle topic='%s'", converter, topic)
		loggedTopics[id] = true
	}
}
