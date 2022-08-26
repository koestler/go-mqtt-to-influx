package statistics

import (
	"strings"
	"time"
)

type HierarchicalCounts map[string]map[string]map[string]map[string]int

type requestHierarchicalCounts struct {
	response chan HierarchicalCounts
}

func (s *InMemoryStatistics) GetHierarchicalCountsStructless() interface{} {
	return s.GetHierarchicalCounts()
}

func (s *InMemoryStatistics) GetHierarchicalCounts() HierarchicalCounts {
	if !s.Enabled() {
		return nil
	}
	response := make(chan HierarchicalCounts)
	s.requestHierarchicalCounts <- requestHierarchicalCounts{
		response: response,
	}
	return <-response
}

func (s *InMemoryStatistics) handleRequestHierarchicalCounts(request requestHierarchicalCounts) {
	// copy / restructure data
	ret := make(HierarchicalCounts)

	d0 := s.config.HistoryResolution() * 10
	d1 := s.config.HistoryMaxAge()
	last0 := s.getHistoricalCounts(d0)
	last1 := s.getHistoricalCounts(d1)
	d0Str := "last" + shortDur(d0)
	d1Str := "last" + shortDur(d1)

	for desc, count := range s.total {
		if _, ok := ret[desc.module]; !ok {
			ret[desc.module] = make(map[string]map[string]map[string]int)
		}
		if _, ok := ret[desc.module][desc.name]; !ok {
			ret[desc.module][desc.name] = make(map[string]map[string]int)
		}
		ret[desc.module][desc.name][desc.field] = map[string]int{
			"total": count,
			d0Str:   last0[desc],
			d1Str:   last1[desc],
		}
	}

	request.response <- ret
	close(request.response)
}

func shortDur(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}
