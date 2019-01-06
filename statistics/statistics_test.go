package statistics

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influxdb/statistics/mock"
	"testing"
	"time"
)

//go:generate mockgen -destination=mock/statistics_mock.go -package mock github.com/koestler/go-mqtt-to-influxdb/statistics Config

func Test(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mock.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Enabled().Return(true).AnyTimes()
	mockConfig.EXPECT().HistoryResolution().Return(10 * time.Millisecond).AnyTimes()
	mockConfig.EXPECT().HistoryMaxAge().Return(50 * time.Millisecond).AnyTimes()

	s := Run(mockConfig)
	// t = 0ms

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 3)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 4)
	incrementN(s, "mqtt", "foobar", "piegn/tele/bar/SENSOR", 9)
	time.Sleep(10 * time.Millisecond)

	// t = 10ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 1)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 2)
	time.Sleep(10 * time.Millisecond)

	// t = 20ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 3)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 4)
	time.Sleep(20 * time.Millisecond)

	// t = 40ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 5)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 6)
	incrementN(s, "influxDb", "foobar", "piegn/tele/bar/SENSOR", 3)
	time.Sleep(10 * time.Millisecond)

	// t = 50ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 3)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 2)
	time.Sleep(20 * time.Millisecond)

	// t = 70ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 3)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 2)
	time.Sleep(10 * time.Millisecond)

	// t = 80ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 1)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 1)
	time.Sleep(10 * time.Millisecond)

	// t = 90ms
	printHistorical(t, s)

	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR", 4)
	incrementN(s, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR", 3)
	time.Sleep(10 * time.Millisecond)

	// t = 100ms
	printHistorical(t, s)

	if c := getHistorical(s, 25*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR"); c != 5 {
		t.Errorf("expect getHistorical(25ms) == 4+1 for foo1; got=%d", c)
	}
	if c := getHistorical(s, 25*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR"); c != 4 {
		t.Errorf("expect getHistorical(25ms) == 3+1 for foo2; got=%d", c)
	}
	if c := getHistorical(s, 55*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR"); c != 11 {
		t.Errorf("expect getHistorical(55ms) == 4+1+3+3 for foo1; got=%d", c)
	}
	if c := getHistorical(s, 55*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo2/SENSOR"); c != 8 {
		t.Errorf("expect getHistorical(55ms) == 3+1+2+2 for foo2; got=%d", c)
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

	time.Sleep(100 * time.Millisecond)
	// t = 200ms
	printHistorical(t, s)

	if c := getHistorical(s, 25*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR"); c != 0 {
		t.Errorf("expect getHistorical(25ms) == 0 for foo1; got=%d", c)
	}
	if c := getHistorical(s, 55*time.Millisecond, "mqtt", "0-piegn-mosquitto", "piegn/tele/foo1/SENSOR"); c != 0 {
		t.Errorf("expect getHistorical(55ms) == 0 for foo1; got=%d", c)
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

func incrementN(s *Statistics, module, name, field string, n int) {
	for i := 0; i < n; i += 1 {
		s.IncrementOne(module, name, field)
	}
}

func getHistorical(s *Statistics, duration time.Duration, module, name, field string) int {
	return s.getHistoricalCounts(duration)[Desc{module, name, field}]
}

func printHistorical(t *testing.T, s *Statistics) {
	t.Log("printHistorical:")

	for e := s.historical.Front(); e != nil; e = e.Next() {
		c := e.Value.(*HistoricalCount)
		t.Logf("c.NewerThan=%s", c.NewerThan)
		for desc, count := range c.Count {
			t.Logf("  desc=%v, count=%d", desc, *count)
		}
	}
	t.Log("")
}
