package statistics

import (
	"container/list"
	"time"
)

type Statistics interface {
	Enabled() bool
	IncrementOne(module, name, field string)
	GetHierarchicalCountsStructless() interface{}
	GetHierarchicalCounts() HierarchicalCounts
}

type InMemmoryStatistics struct {
	config Config

	total      map[Desc]int
	historical *list.List

	// input channels
	incrementOne chan Desc

	// output channels
	requestHierarchicalCounts chan requestHierarchicalCounts
}

type DisabledStatistics struct{}

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
	Count     map[Desc]int
}

func Run(config Config) (stats Statistics) {
	if !config.Enabled() {
		return &DisabledStatistics{}
	}

	return RunInMemory(config)
}

func RunInMemory(config Config) (stats *InMemmoryStatistics) {
	inMemoryStats := &InMemmoryStatistics{
		config:                    config,
		total:                     make(map[Desc]int),
		historical:                list.New(),
		incrementOne:              make(chan Desc, 1024),
		requestHierarchicalCounts: make(chan requestHierarchicalCounts),
	}

	// start incrementer routine
	go inMemoryStats.countWorker()

	return inMemoryStats
}

func (s InMemmoryStatistics) Enabled() bool {
	return true
}

func (s *InMemmoryStatistics) IncrementOne(module, name, field string) {
	s.incrementOne <- Desc{
		module: module,
		name:   name,
		field:  field,
	}
}

func (s DisabledStatistics) Enabled() bool {
	return false
}

func (s *DisabledStatistics) IncrementOne(module, name, field string) {}

func (s *DisabledStatistics) GetHierarchicalCountsStructless() interface{} {
	return struct{}{}
}

func (s *DisabledStatistics) GetHierarchicalCounts() HierarchicalCounts {
	return nil
}
