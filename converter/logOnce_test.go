package converter

import (
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influx/converter/mock"
	"log"
	"sync"
	"testing"
)

func TestLogOnce(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)

	var wg sync.WaitGroup

	const numbTopics = 42
	const numbPayloads = 11
	const repeatPayload = 100

	// use LogTopicOnce as much in parallel as possible to check proper sync handling
	for topic := 0; topic < numbTopics; topic += 1 {
		go func(topic string) {
			for payload := 0; payload < numbPayloads; payload += 1 {
				wg.Add(1)
				go func(payload string) {
					defer wg.Done()
					for repeat := 0; repeat < repeatPayload; repeat += 1 {
						mockInput := converter_mock.NewMockInput(mockCtrl)
						mockInput.EXPECT().Topic().Return(topic).MinTimes(1)
						mockInput.EXPECT().Payload().Return([]byte(payload)).MinTimes(0)

						LogTopicOnce("test-converter", mockInput)
					}
				}(fmt.Sprintf("payload=%d", payload))
			}
		}(fmt.Sprintf("/a/%d/foo", topic))
	}

	wg.Wait()

	if cnt := bytes.Count(logBuffer.Bytes(), []byte{'\n'}); cnt != numbTopics {
		t.Errorf("expected number of lines to match numbTopics=%d, but go %d lines", numbTopics, cnt)
	}
}
