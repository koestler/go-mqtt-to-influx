package statistics

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influxdb/statistics/mock"
	"testing"
	"time"
)

//go:generate mockgen -destination=mock/config_mock.go -package statistics_mock github.com/koestler/go-mqtt-to-influxdb/statistics Config

func TestEnabled(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := statistics_mock.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Enabled().Return(true).AnyTimes()
	mockConfig.EXPECT().HistoryResolution().Return(100 * time.Millisecond).AnyTimes()
	mockConfig.EXPECT().HistoryMaxAge().Return(600 * time.Millisecond).AnyTimes()

	s := Run(mockConfig)
	simulationCase0(t, s)

	if c := getHistorical(s, 225*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR"); c != 5 {
		t.Errorf("expect getHistorical(225ms) == 4+1 for foo1; got=%d", c)
	}
	if c := getHistorical(s, 225*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR"); c != 4 {
		t.Errorf("expect getHistorical(225ms) == 3+1 for foo2; got=%d", c)
	}
	if c := getHistorical(s, 525*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR"); c != 11 {
		t.Errorf("expect getHistorical(525ms) == 4+1+3+3 for foo1; got=%d", c)
	}
	if c := getHistorical(s, 525*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR"); c != 8 {
		t.Errorf("expect getHistorical(525ms) == 3+1+2+2 for foo2; got=%d", c)
	}

	{
		counts := s.getHierarchicalCounts()
		if r := counts["mqtt"]["0-piegn-mosquitto"]["piegn/tele/foo1/SENSOR"].Total; r != 23 {
			t.Errorf("expect counts.[mqtt][0-piegn-mosquitto][piegn/tele/foo1/SENSOR].Total == 23, got=%v", r)
		}
		if r := counts["mqtt"]["0-piegn-mosquitto"]["piegn/tele/foo2/SENSOR"].Total; r != 24 {
			t.Errorf("expect counts.[mqtt][0-piegn-mosquitto][piegn/tele/foo2/SENSOR].Total == 24, got=%v", r)
		}
	}

	time.Sleep(1000 * time.Millisecond)
	// t = 200ms
	printHistorical(t, s)

	if c := getHistorical(s, 225*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR"); c != 0 {
		t.Errorf("expect getHistorical(225ms) == 0 for foo1; got=%d", c)
	}
	if c := getHistorical(s, 525*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR"); c != 0 {
		t.Errorf("expect getHistorical(525ms) == 0 for foo1; got=%d", c)
	}

	{
		counts := s.getHierarchicalCounts()
		if r := counts["mqtt"]["0-piegn-mosquitto"]["piegn/tele/foo1/SENSOR"].Total; r != 23 {
			t.Errorf("expect counts.[mqtt][0-piegn-mosquitto][piegn/tele/foo1/SENSOR].Total == 23, got=%v", r)
		}
		if r := counts["mqtt"]["0-piegn-mosquitto"]["piegn/tele/foo2/SENSOR"].Total; r != 24 {
			t.Errorf("expect counts.[mqtt][0-piegn-mosquitto][piegn/tele/foo2/SENSOR].Total == 24, got=%v", r)
		}
	}
}

func TestDisabled(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := statistics_mock.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Enabled().Return(false).AnyTimes()

	s := Run(mockConfig)
	simulationCase0(t, s)

	// must not crash even if module disabled
	getHistorical(s, time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR")
	s.getHierarchicalCounts()
}

func simulationCase0(t *testing.T, s *Statistics) {

	// t = 0ms

	time.Sleep(50 * time.Millisecond)
	// t = 50ms

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 3)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 4)
	incrementN(s, "mqtt", "foobar", "piegn/tele/bar/SENSOR", 9)
	time.Sleep(100 * time.Millisecond)

	// t = 150ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 1)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 2)
	time.Sleep(100 * time.Millisecond)

	// t = 250ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 3)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 4)
	time.Sleep(200 * time.Millisecond)

	// t = 450ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 5)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 6)
	incrementN(s, "influxDb", "foobar", "piegn/tele/bar/SENSOR", 3)
	time.Sleep(100 * time.Millisecond)

	// t = 550ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 3)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 2)
	time.Sleep(200 * time.Millisecond)

	// t = 750ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 3)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 2)
	time.Sleep(100 * time.Millisecond)

	// t = 850ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 1)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 1)
	time.Sleep(100 * time.Millisecond)

	// t = 950ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 4)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 3)
	time.Sleep(100 * time.Millisecond)

	// t = 1050ms
	printHistorical(t, s)

	time.Sleep(100 * time.Millisecond)

	// t = 1150ms
	printHistorical(t, s)

}

func incrementN(s *Statistics, module, name, field string, n int) {
	for i := 0; i < n; i += 1 {
		s.IncrementOne(module, name, field)
	}
}

func getHistorical(s *Statistics, duration time.Duration, module, name, field string) int {
	return s.getHistoricalCounts(duration)[Desc{module, name, field}]
}

func printHistorical(t *testing.T, s *Statistics) {
	if !s.Enabled() {
		return
	}

	t.Log("printHistorical:")

	for e := s.historical.Front(); e != nil; e = e.Next() {
		c := e.Value.(*HistoricalCount)
		t.Logf("c.NewerThan=%s", c.NewerThan)
		for desc, count := range c.Count {
			t.Logf("  desc=%v, count=%d", desc, count)
		}
	}
	t.Log("")
}
