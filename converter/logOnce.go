package converter

import (
	"fmt"
	"log"
	"sync"
)

var (
	loggedTopics     = make(map[string]bool)
	loggedTopicMutex sync.Mutex
)

func LogTopicOnce(converter string, input Input) {
	id := fmt.Sprintf("%s,%s", converter, input.Topic())

	loggedTopicMutex.Lock()
	defer loggedTopicMutex.Unlock()

	if _, ok := loggedTopics[id]; !ok {
		log.Printf("converter[%s]: handle topic='%s', payload='%s'", converter, input.Topic(), input.Payload())
		loggedTopics[id] = true
	}
}
