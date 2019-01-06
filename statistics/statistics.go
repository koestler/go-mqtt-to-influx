package statistics

import (
	"container/list"
	"time"
)

type Statistics struct {
	config Config

	total      map[Desc]*int
	historical *list.List

	// input channels
	incrementOne chan Desc

	// output channels
	requestHierarchicalCounts chan requestHierarchicalCounts
}

type Config interface {
	Enabled() bool
	HistoryResolution() time.Duration
	HistoryMaxAge() time.Duration
}

type Desc struct {
	module string
	name   string
	field  string
}

type HistoricalCount struct {
	NewerThan time.Time
	Count     map[Desc]*int
}

func Run(config Config) (stats *Statistics) {
	if !config.Enabled() {
		return &Statistics{
			config: config,
		}
	}

	stats = &Statistics{
		config:                    config,
		total:                     make(map[Desc]*int),
		historical:                list.New(),
		incrementOne:              make(chan Desc, 128),
		requestHierarchicalCounts: make(chan requestHierarchicalCounts),
	}

	// start incrementer routine
	go stats.countWorker()

	return stats
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
