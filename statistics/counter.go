package statistics

import (
	"log"
	"time"
)

func (s *Statistics) countWorker() {
	ticker := time.Tick(s.config.HistoryResolution())

	s.historical.PushBack(&HistoricalCount{
		NewerThan: time.Now().Add(s.config.HistoryResolution()),
		Count:     make(map[Desc]*int),
	})

	for {
		select {
		case desc := <-s.incrementOne:
			s.handleIncrementOne(desc)
		case now := <-ticker:
			s.handleHistoryTick(now)
		case request := <-s.requestHierarchicalCounts:
			s.handleRequestHierarchicalCounts(request)
		}
	}
}

func (s *Statistics) handleIncrementOne(desc Desc) {
	log.Printf("handleIncrementOne(%v)", desc)

	// handle total
	if t := s.total[desc]; t == nil {
		i := 1
		s.total[desc] = &i
	} else {
		// existing element -> increment
		*t += 1
	}

	// historical data: increment one to newest entry
	newest := s.historical.Back().Value.(*HistoricalCount)
	if t := newest.Count[desc]; t == nil {
		i := 1
		newest.Count[desc] = &i
	} else {
		// existing element -> increment
		*t += 1
	}
}

func (s *Statistics) handleHistoryTick(now time.Time) {
	// create new historical list entry if newest one is outdated
	if back := s.historical.Back(); back == nil || back.Value.(*HistoricalCount).NewerThan.Before(now) {
		if back != nil && len(back.Value.(*HistoricalCount).Count) < 1 {
			// newest entry is empty -> update it
			back.Value.(*HistoricalCount).NewerThan = now
		} else {
			// list is empty or newest entry is too old -> create new one
			s.historical.PushBack(&HistoricalCount{
				NewerThan: now.Add(s.config.HistoryResolution()),
				Count:     make(map[Desc]*int),
			})
		}
	}

	// remove too old historical data but keep at least one
	oldest := now.Add(-s.config.HistoryMaxAge())
	for e := s.historical.Front(); e != nil && s.historical.Len() > 0; e = e.Next() {
		c := e.Value.(*HistoricalCount)
		if c.NewerThan.After(oldest) {
			break
		}
		// remove element
		s.historical.Remove(e)
	}
}

func (s *Statistics) getHistoricalCounts(duration time.Duration) (ret map[Desc]int) {
	ret = make(map[Desc]int)

	limit := time.Now().Add(-duration)

	// iterate reverse over history
	for e := s.historical.Back(); e != nil; e = e.Prev() {
		c := e.Value.(*HistoricalCount)

		// until NewerThan is reached
		if c.NewerThan.Before(limit) {
			break;
		}

		// accumulate counts
		for desc, i := range c.Count {
			ret[desc] += *i
		}
	}

	return
}
