package metrics

import (
	"sync"
	"time"
)

type TimeWindow struct {
	ID    uint64
	Start time.Time
	End   time.Time
}

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

func (t *TimeWindows) StartAndEnd() (start, end *TimeWindow) {
	t.lock.RLock()
	latest := t.rb.Last()
	t.lock.RUnlock()
	if time.Since(latest.End) < 0 {
		return t.rb.First(), latest
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
