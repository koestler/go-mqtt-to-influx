package statistics

type Statistics struct {
	config Config

	counts map[Desc]*Counts

	// input channels
	incrementOne chan Desc

	// output channels
	requestHierarchicalCounts chan requestHierarchicalCounts
}

type Config interface {
	Enabled() bool
}

type HierarchicalCounts map[string]map[string]map[string]Counts

type requestHierarchicalCounts struct {
	response chan HierarchicalCounts
}

type Desc struct {
	module string
	name   string
	field  string
}

type Counts struct {
	Total int
}

func Run(config Config) (stats *Statistics) {
	if !config.Enabled() {
		return &Statistics{
			config: config,
		}
	}

	stats = &Statistics{
		config:                    config,
		counts:                    make(map[Desc]*Counts),
		incrementOne:              make(chan Desc, 128),
		requestHierarchicalCounts: make(chan requestHierarchicalCounts),
	}

	// start incrementer routine
	go stats.worker()

	return stats
}

func (s *Statistics) worker() {
	for {
		select {
		case d := <-s.incrementOne:
			if s.counts[d] == nil {
				s.counts[d] = &Counts{
					Total: 1,
				}
			} else {
				s.counts[d].Total += 1
			}
		case request := <-s.requestHierarchicalCounts:
			// copy / restructure data
			ret := make(HierarchicalCounts)
			for desc, count := range s.counts {
				if _, ok := ret[desc.module]; !ok {
					ret[desc.module] = make(map[string]map[string]Counts)
				}
				if _, ok := ret[desc.module][desc.name]; !ok {
					ret[desc.module][desc.name] = make(map[string]Counts)
				}
				ret[desc.module][desc.name][desc.field] = *count
			}
			request.response <- ret
		}
	}

}

func (s *Statistics) Enabled() bool {
	return s.config.Enabled()
}

func (s *Statistics) IncrementOne(module, name, field string) {
	if !s.Enabled() {
		return
	}

	s.incrementOne <- Desc{
		module: module,
		name:   name,
		field:  field,
	}
}

func (s *Statistics) GetHierarchicalCounts() (ret interface{}) {
	if !s.Enabled() {
		return
	}
	response := make(chan HierarchicalCounts)
	s.requestHierarchicalCounts <- requestHierarchicalCounts{
		response: response,
	}
	return <-response
}
