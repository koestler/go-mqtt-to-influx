package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influxdb/converter/mock"
	"github.com/koestler/go-mqtt-to-influxdb/influxDbClient"
	"reflect"
	"strings"
	"testing"
	"time"
)

//go:generate mockgen -destination=mock/converter_mock.go -package converter_mock github.com/koestler/go-mqtt-to-influxdb/converter Config,Statistics,Input,Output

const epsilon = time.Millisecond

func checkNow(t time.Time) bool {
	now := time.Now()
	return t.After(now.Add(-epsilon)) && t.Before(now.Add(epsilon))
}

func getLineWoTime(line string) string {
	return strings.Join(strings.Split(line, " ")[0:2], " ")
}

type TestStimuliResponse []struct {
	Topic         string
	Payload       string
	ExpectedLines []string
}

func testStimuliResponse(
	t *testing.T,
	mockCtrl *gomock.Controller,
	config Config,
	dut HandleFunc,
	stimuli TestStimuliResponse,
	) {
	for _, s := range stimuli {
		t.Logf("stimuli: Topic='%s'", s.Topic)
		t.Logf("stimuli: Payload='%s'", s.Payload)
		t.Logf("stimuli: ExpectedLines='%s'", s.ExpectedLines)

		mockInput := converter_mock.NewMockInput(mockCtrl)
		mockInput.EXPECT().Topic().Return(s.Topic).MinTimes(1)
		mockInput.EXPECT().Payload().Return([]byte(s.Payload)).AnyTimes() // must no be called when topic is invalid

		outputTestFuncCounter := 0
		responseLines := make([]string, 0, len(s.ExpectedLines))
		outputTestFunc := func(output Output) {
			outputTestFuncCounter += 1

			point, err := influxDbClient.ToInfluxPoint(output)
			if err != nil {
				t.Errorf("expect no error, got: %v", err)
			}

			response := getLineWoTime(point.String())
			t.Logf("response: '%s'", response)

			responseLines = append(responseLines, response)

			if !checkNow(output.Time()) {
				t.Errorf("expect timestamp to be now")
			}
		}
		dut(config, mockInput, outputTestFunc)

		if !reflect.DeepEqual(s.ExpectedLines, responseLines) {
			t.Errorf("expected %v, but got %v", s.ExpectedLines, responseLines)
		}
	}
}
