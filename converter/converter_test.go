package converter

import (
	"bytes"
	"github.com/golang/mock/gomock"
	influxdb2Write "github.com/influxdata/influxdb-client-go/v2/api/write"
	converter_mock "github.com/koestler/go-mqtt-to-influx/v2/converter/mock"
	"github.com/koestler/go-mqtt-to-influx/v2/influxClient"
	"log"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

//go:generate mockgen -destination=mock/converter_mock.go -package converter_mock github.com/koestler/go-mqtt-to-influx/v2/converter Config,Input,Output
//go:generate mockgen -destination=mock/topicMatcher_mock.go -package converter_mock github.com/koestler/go-mqtt-to-influx/v2/converter TopicMatcherConfig

const epsilon = 10 * time.Millisecond

func checkTimeStamp(expected, response time.Time) bool {
	return response.After(expected.Add(-epsilon)) && response.Before(expected.Add(epsilon))
}

func getLineWoTime(line string) string {
	parts := strings.Split(line, " ")
	return strings.Join(parts[0:len(parts)-1], " ")
}

func pointToLine(output Output) string {
	point := influxClient.ToInfluxPoint(output)
	return influxdb2Write.PointToLineProtocol(point, time.Second)
}

type TestStimuliResponse []struct {
	Topic             string
	Payload           string
	ExpectedTimeStamp time.Time
	ExpectedLines     []string
}

func testStimuliResponse(
	t *testing.T,
	mockCtrl *gomock.Controller,
	config Config,
	tmConfig TopicMatcherConfig,
	dut HandleFunc,
	stimuli TestStimuliResponse,
) {
	tm, err := CreateTopicMatcher(tmConfig)
	if err != nil {
		t.Error(err)
	}

	for _, s := range stimuli {
		t.Logf("stimuli: Topic='%s'", s.Topic)
		t.Logf("stimuli: Payload='%s'", s.Payload)
		t.Logf("stimuli: ExpectedLines='%s'", s.ExpectedLines)

		mockInput := converter_mock.NewMockInput(mockCtrl)
		mockInput.EXPECT().Topic().Return(s.Topic).MinTimes(1)
		mockInput.EXPECT().Payload().Return([]byte(s.Payload)).AnyTimes() // must not be called when topic is invalid

		outputTestFuncCounter := 0
		responseLines := make([]string, 0, len(s.ExpectedLines))
		outputTestFunc := func(output Output) {
			outputTestFuncCounter += 1

			response := getLineWoTime(pointToLine(output))
			t.Logf("response: '%s'", response)

			responseLines = append(responseLines, response)

			expectedTime := s.ExpectedTimeStamp
			if strings.HasPrefix(response, "timeValue,") {
				expectedTime = time.Now()
			}
			if !checkTimeStamp(expectedTime, output.Time()) {
				t.Errorf("expect timestamp to %s but got %s", s.ExpectedTimeStamp, output.Time())
			}
		}
		dut(config, tm, mockInput, outputTestFunc)

		// sort strings before comparison
		sort.Strings(s.ExpectedLines)
		sort.Strings(responseLines)
		if !reflect.DeepEqual(s.ExpectedLines, responseLines) {
			t.Errorf("expected lines do not match response lines:")

			t.Errorf("  expected: %d lines:", len(s.ExpectedLines))
			for _, l := range s.ExpectedLines {
				t.Errorf("    %s", l)
			}
			t.Errorf("  got: %d lines:", len(responseLines))
			for _, l := range responseLines {
				t.Errorf("    %s", l)
			}
		}
	}
}

func TestInvalidHandler(t *testing.T) {
	_, err := GetHandler("unknown")
	if err == nil {
		t.Errorf("expected an error")
	}
}

func TestRegisterTwice(t *testing.T) {
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected a panic when registering the same handler twice")
		}
	}()

	registerHandler("empty", func(c Config, tm TopicMatcher, input Input, outputFunc OutputFunc) {})
	registerHandler("empty", func(c Config, tm TopicMatcher, input Input, outputFunc OutputFunc) {})
}
