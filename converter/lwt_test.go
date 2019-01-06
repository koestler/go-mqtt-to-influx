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

func Test(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := converter_mock.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Name().Return("test-converter").AnyTimes()
	mockConfig.EXPECT().TargetMeasurement().Return("boolValue").MinTimes(1)

	stimuli := []struct {
		Topic         string
		Payload       string
		ExpectedLines []string
	}{
		{
			Topic:         "piegn/tele/software/srv1-go-ve-sensor/LWT",
			Payload:       "Online",
			ExpectedLines: []string{"boolValue,device=software/srv1-go-ve-sensor,field=Available value=true"},
		}, {
			Topic:         "piegn/tele/software/srv1-go-ve-sensor/LWT",
			Payload:       "Offline",
			ExpectedLines: []string{"boolValue,device=software/srv1-go-ve-sensor,field=Available value=false"},
		}, {
			Topic:         "piegn/tele/software/srv1-go-ve-sensor/LWT",
			Payload:       "invalid",
			ExpectedLines: []string{},
		}, {
			Topic:         "piegn/tele/software/srv1-go-ve-sensor/LWT-invalid-topic",
			Payload:       "Online",
			ExpectedLines: []string{},
		},
	}

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
		lwtHandler(mockConfig, mockInput, outputTestFunc)

		if !reflect.DeepEqual(s.ExpectedLines, responseLines) {
			t.Errorf("expected %v, but got %v", s.ExpectedLines, responseLines)
		}
	}

}
