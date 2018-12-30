package statistics

import "sync"

type Statistics struct {
	counts      Counts
	countsMutex sync.RWMutex

	incrementOne chan Desc
}

type Desc struct {
	module string
	field  string
}

type Counts struct {
	// 1. dim: module
	// 2. dim: field
	Total map[Desc]int
}

func Run() (stats *Statistics) {
	stats = &Statistics{
		counts: Counts{
			Total: make(map[Desc]int),
		},
		incrementOne: make(chan Desc, 128),
	}

	// start incrementer routine
	go stats.incerementer()

	return stats
}

func (s *Statistics) incerementer() {
	s.countsMutex.Lock()
	defer s.countsMutex.Unlock()

}

func (s *Statistics) GetTotalPerModule() (ret map[string]map[string]int) {
	s.countsMutex.RLock()
	defer s.countsMutex.RUnlock()

	// copy / restructe data
	ret = make(map[string]map[string]int)
	for desc, tot := range s.counts.Total {
		if _, ok := ret[desc.module]; !ok {
			ret[desc.module] = make(map[string]int)
		}
		ret[desc.module][desc.field] = tot
	}

	return
}
