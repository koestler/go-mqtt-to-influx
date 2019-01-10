package statistics

type HierarchicalCount struct {
	Total     int
	Last10Res int
	LastMax   int
}

type HierarchicalCounts map[string]map[string]map[string]HierarchicalCount

type requestHierarchicalCounts struct {
	response chan HierarchicalCounts
}

func (s *InMemmoryStatistics) GetHierarchicalCounts() interface{} {
	return s.getHierarchicalCounts()
}

func (s *InMemmoryStatistics) getHierarchicalCounts() HierarchicalCounts {
	if !s.Enabled() {
		return nil
	}
	response := make(chan HierarchicalCounts)
	s.requestHierarchicalCounts <- requestHierarchicalCounts{
		response: response,
	}
	return <-response
}

func (s *InMemmoryStatistics) handleRequestHierarchicalCounts(request requestHierarchicalCounts) {
	// copy / restructure data
	ret := make(HierarchicalCounts)

	last10Res := s.getHistoricalCounts(s.config.HistoryResolution() * 10)
	lastMax := s.getHistoricalCounts(s.config.HistoryMaxAge())

	for desc, count := range s.total {
		if _, ok := ret[desc.module]; !ok {
			ret[desc.module] = make(map[string]map[string]HierarchicalCount)
		}
		if _, ok := ret[desc.module][desc.name]; !ok {
			ret[desc.module][desc.name] = make(map[string]HierarchicalCount)
		}
		ret[desc.module][desc.name][desc.field] = HierarchicalCount{
			Total:     count,
			Last10Res: last10Res[desc],
			LastMax:   lastMax[desc],
		}
	}

	request.response <- ret
	close(request.response)
}
