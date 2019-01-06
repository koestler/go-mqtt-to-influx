package statistics

type HierarchicalCount struct {
	Total int
	Last1 int
}

type HierarchicalCounts map[string]map[string]map[string]HierarchicalCount

type requestHierarchicalCounts struct {
	response chan HierarchicalCounts
}

func (s *Statistics) GetHierarchicalCounts() (interface{}) {
	return s.getHierarchicalCounts()
}

func (s *Statistics) getHierarchicalCounts() (HierarchicalCounts) {
	if !s.Enabled() {
		return nil
	}
	response := make(chan HierarchicalCounts)
	s.requestHierarchicalCounts <- requestHierarchicalCounts{
		response: response,
	}
	return <-response
}

func (s *Statistics) handleRequestHierarchicalCounts(request requestHierarchicalCounts) {
	// copy / restructure data
	ret := make(HierarchicalCounts)

	for desc, count := range s.total {
		if _, ok := ret[desc.module]; !ok {
			ret[desc.module] = make(map[string]map[string]HierarchicalCount)
		}
		if _, ok := ret[desc.module][desc.name]; !ok {
			ret[desc.module][desc.name] = make(map[string]HierarchicalCount)
		}
		ret[desc.module][desc.name][desc.field] = HierarchicalCount{
			Total: *count,
		}
	}

	request.response <- ret
	close(request.response)
}
