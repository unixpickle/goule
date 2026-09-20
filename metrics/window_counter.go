package metrics

import "sync"

var (
	CountsKeyEgress   = "egress"
	CountsKeyIngress  = "ingress"
	CountsKeyRequests = "requests"
)

type WindowCounts struct {
	Window *TimeWindow
	Values map[string]uint64
}

type WindowCounter struct {
	lock    sync.RWMutex
	windows *TimeWindows
	counts  *ringBuffer[*WindowCounts]
}

func NewWindowCounter(windows *TimeWindows) *WindowCounter {
	return &WindowCounter{windows: windows, counts: newRingBuffer[*WindowCounts]()}
}

func (w *WindowCounter) Add(metric string, amount uint64) {
	first, last := w.windows.StartAndEnd()
	w.lock.Lock()
	defer w.lock.Unlock()
	for w.counts.First().Window.ID < first.ID {
		w.counts.PopFirst()
	}
	lastCounts := w.counts.Last()
	if lastCounts.Window.ID > last.ID {
		panic("ID should never decrease")
	}
	if lastCounts.Window.ID != last.ID {
		lastCounts = &WindowCounts{Window: last}
		w.counts.Push(lastCounts)
	}
	curValue, _ := lastCounts.Values[metric]
	lastCounts.Values[metric] = curValue + amount
}
