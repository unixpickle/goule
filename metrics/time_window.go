package metrics

import (
	"sync"
	"time"
)

// A TimeWindow is chunk of time in [start, end), where ID increments for
// successive chunks of time within a TimeWindows context.
type TimeWindow struct {
	ID    uint64
	Start time.Time
	End   time.Time
}

// TimeWindows manages shared, time-aligned instances of *TimeWindow.
type TimeWindows struct {
	lock     sync.RWMutex
	rb       *ringBuffer[*TimeWindow]
	interval time.Duration
	limit    time.Duration
}

func NewTimeWindows(start time.Time, interval, limit time.Duration) *TimeWindows {
	rb := newRingBuffer[*TimeWindow]()
	firstWindow := &TimeWindow{
		ID:    0,
		Start: start,
		End:   start.Add(interval),
	}
	rb.Push(firstWindow)
	return &TimeWindows{
		rb:       rb,
		interval: interval,
		limit:    limit,
	}
}

// StartAndEnd returns the current first and last time window, possibly
// updating the state if necessary.
func (t *TimeWindows) StartAndEnd() (start, end *TimeWindow) {
	start, end = t.currentStartAndEnd()
	if time.Since(end.End) < 0 {
		return start, end
	}

	// Create new windows as necessary, noting that there could
	// have been an arbitrary delay since the last call.
	t.lock.Lock()
	defer t.lock.Unlock()
	for {
		latest := t.rb.Last()
		if time.Since(latest.End) < 0 {
			return t.rb.First(), latest
		}
		newBlock := &TimeWindow{
			ID:    latest.ID + 1,
			Start: latest.End,
			End:   latest.End.Add(t.interval),
		}
		t.rb.Push(newBlock)
		for t.rb.Len() > 1 {
			if time.Since(t.rb.First().End) < t.limit {
				break
			}
			t.rb.PopFirst()
		}
	}
}

func (t *TimeWindows) currentStartAndEnd() (start, end *TimeWindow) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.rb.First(), t.rb.Last()
}
