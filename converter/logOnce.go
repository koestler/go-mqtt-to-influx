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

func logTopicOnce(converter, topic string) {
	id := fmt.Sprintf("%s,%s", converter, topic)

	loggedTopicMutex.RLock()
	_, ok := loggedTopics[id]
	loggedTopicMutex.RUnlock()

	if !ok {
		log.Printf("converter[%s]: handle topic='%s'", converter, topic)
		loggedTopicMutex.Lock()
		loggedTopics[id] = true
		loggedTopicMutex.Unlock()
	}
}
