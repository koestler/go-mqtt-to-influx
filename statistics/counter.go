package statistics

import (
	"time"
)

func (s *InMemoryStatistics) countWorker() {
	ticker := time.NewTicker(s.config.HistoryResolution())
	defer ticker.Stop()

	s.historical.PushBack(&HistoricalCount{
		NewerThan: time.Now(),
		Count:     make(map[Desc]int),
	})

	for {
		select {
		case desc := <-s.incrementOne:
			s.handleIncrementOne(desc)
		case now := <-ticker.C:
			s.handleHistoryTick(now)
		case request := <-s.requestHierarchicalCounts:
			s.handleRequestHierarchicalCounts(request)
		}
	}
}

func (s *InMemoryStatistics) handleIncrementOne(desc Desc) {
	// handle total
	if count, ok := s.total[desc]; !ok {
		s.total[desc] = 1
	} else {
		// existing element -> increment
		s.total[desc] = count + 1
	}

	// historical data: increment one to newest entry
	newest := s.historical.Back().Value.(*HistoricalCount)
	if count, ok := newest.Count[desc]; !ok {
		newest.Count[desc] = 1
	} else {
		// existing element -> increment
		newest.Count[desc] = count + 1
	}
}

func (s *InMemoryStatistics) handleHistoryTick(now time.Time) {
	// create new historical list entry if newest one is outdated
	back := s.historical.Back()
	h := back.Value.(*HistoricalCount)
	if len(h.Count) < 1 {
		// newest entry is empty -> update it
		h.NewerThan = now
	} else {
		// newest entry is too old and not empty -> create new one
		// use 1.2 times last used count size as capacity hint
		s.historical.PushBack(&HistoricalCount{
			NewerThan: now,
			Count:     make(map[Desc]int, int(1.2*float64(len(h.Count)))),
		})
	}

	// remove too old historical data but keep at least one
	// - s.config.HistoryResolution() is used to accommodate tick jitter
	oldest := now.Add(-s.config.HistoryMaxAge() - s.config.HistoryResolution())
	for e := s.historical.Front(); e != nil && s.historical.Len() > 0; e = e.Next() {
		if e.Value.(*HistoricalCount).NewerThan.After(oldest) {
			break
		}
		// remove element
		s.historical.Remove(e)
	}
}

func (s *InMemoryStatistics) getHistoricalCounts(duration time.Duration) (ret map[Desc]int) {
	if !s.Enabled() {
		return
	}

	if s.historical.Back() == nil {
		// list empty
		return
	}

	ret = make(map[Desc]int)

	// use 1.5 Resolution to compensate for tick jitter
	limit := time.Now().Add(-duration - s.config.HistoryResolution() - s.config.HistoryResolution()/2)

	// iterate reverse over history
	// having:
	// tN - tN-1 = HistoryResolution
	// duration = 3 * HistoryResolution
	// t0... t1... t2... t3... t4... t5... t6..
	//                   xxxxxxxxxxxxxxxxx        used
	//                     xxxxxxxxxxxxxxxxx      now - duration ... now
	// use t3, t4, t5, ignore partial t6, ignore older

	// start with second last (do not use active element)
	for e := s.historical.Back().Prev(); e != nil; e = e.Prev() {
		c := e.Value.(*HistoricalCount)

		// if the one we counted is too old already, stop now
		if c.NewerThan.Before(limit) {
			break
		}

		// accumulate counts
		for desc, i := range c.Count {
			ret[desc] += i
		}
	}

	return
}
