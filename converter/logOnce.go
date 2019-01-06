package converter

import (
	"fmt"
	"log"
	"sync"
)

var (
	loggedTopics     = make(map[string]bool)
	loggedTopicMutex sync.RWMutex
)

func logTopicOnce(converter string, input Input) {
	id := fmt.Sprintf("%s,%s", converter, input.Topic())

	loggedTopicMutex.RLock()
	_, ok := loggedTopics[id]
	loggedTopicMutex.RUnlock()

	if !ok {
		log.Printf("converter[%s]: handle topic='%s', payload='%s'", converter, input.Topic(), input.Payload())
		loggedTopicMutex.Lock()
		loggedTopics[id] = true
		loggedTopicMutex.Unlock()
	}
}
