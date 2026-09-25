package metrics

import (
	"sort"
	"sync"
	"time"
)

type MetricKey string

var (
	MetricKeyEgress   MetricKey = "egress"
	MetricKeyIngress  MetricKey = "ingress"
	MetricKeyRequests MetricKey = "requests"
)

type WindowCounts struct {
	Window *TimeWindow
	Values map[MetricKey]uint64
}

type WindowCounter struct {
	lock    sync.RWMutex
	windows *TimeWindows
	counts  *ringBuffer[*WindowCounts]
}

func NewWindowCounter(windows *TimeWindows) *WindowCounter {
	return &WindowCounter{windows: windows, counts: newRingBuffer[*WindowCounts]()}
}

func (w *WindowCounter) Add(metric MetricKey, amount uint64) {
	first, last := w.windows.StartAndEnd()
	w.lock.Lock()
	defer w.lock.Unlock()

	// Remove old windows
	for w.counts.Len() != 0 && w.counts.First().Window.ID < first.ID {
		w.counts.PopFirst()
	}

	var lastCounts *WindowCounts
	if w.counts.Len() != 0 {
		lastCounts = w.counts.Last()
	}

	// Another Add may have recorded a newer window after we obtained
	// last but before we acquired the lock. In that case, add to the
	// newer window rather than append an older one.
	if lastCounts == nil || lastCounts.Window.ID < last.ID {
		lastCounts = &WindowCounts{Window: last}
		w.counts.Push(lastCounts)
	}

	if lastCounts.Values == nil {
		lastCounts.Values = map[MetricKey]uint64{}
	}
	curValue, _ := lastCounts.Values[metric]
	lastCounts.Values[metric] = curValue + amount
}

func (w *WindowCounter) SumSince(start time.Time) map[MetricKey]uint64 {
	w.lock.RLock()
	defer w.lock.RUnlock()
	startIdx := sort.Search(w.counts.Len(), func(idx int) bool {
		return w.counts.At(idx).Window.End.After(start)
	})
	total := map[MetricKey]uint64{}
	for i := startIdx; i < w.counts.Len(); i++ {
		for k, v := range w.counts.At(i).Values {
			cur, _ := total[k]
			cur += v
			total[k] = cur
		}
	}
	return total
}
