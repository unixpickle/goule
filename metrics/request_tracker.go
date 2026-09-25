package metrics

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	WindowDuration = time.Minute * 5
	HistoryLength  = time.Hour * 24 * 7
)

type RequestTracker struct {
	hostsLock sync.RWMutex
	hosts     map[string]struct{}

	windows *TimeWindows

	// Maps from host string to *WindowCounter
	counters sync.Map
}

func NewRequestTracker() *RequestTracker {
	windows := NewTimeWindows(time.Now().Truncate(WindowDuration), WindowDuration, HistoryLength)
	return &RequestTracker{
		hosts:   map[string]struct{}{},
		windows: windows,
	}
}

func (b *RequestTracker) SetHosts(hosts []string) {
	hostSet := map[string]struct{}{}
	for _, h := range hosts {
		hostSet[h] = struct{}{}
	}
	b.hostsLock.Lock()
	b.hosts = hostSet
	b.hostsLock.Unlock()
}

func (b *RequestTracker) allowHost(h string) bool {
	b.hostsLock.RLock()
	defer b.hostsLock.RUnlock()
	_, ok := b.hosts[h]
	return ok
}

func (b *RequestTracker) Wrap(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if !b.allowHost(host) {
			host = ""
		}

		anyTracker, _ := b.counters.LoadOrStore(host, NewWindowCounter(b.windows))
		tracker := anyTracker.(*WindowCounter)
		tracker.Add(MetricKeyRequests, 1)

		writeCallback := func(n int) {
			tracker.Add(MetricKeyEgress, uint64(n))
		}
		readCallback := func(n int) {
			tracker.Add(MetricKeyIngress, uint64(n))
		}

		newW := TrackedResponseWriter(w, readCallback, writeCallback)
		newReq := TrackedRequest(r, readCallback)

		h.ServeHTTP(newW, newReq)
	}
}

// MetricsSince gets a per-host total usage breakdown since the start time for
// all available metrics.
//
// Hosts may be omitted if there was zero traffic for them, and keys
// with zero values may be omitted as well.
func (b *RequestTracker) MetricsSince(start time.Time) map[string]map[MetricKey]uint64 {
	results := map[string]map[MetricKey]uint64{}
	for anyHost, anyHostCounter := range b.counters.Range {
		host := anyHost.(string)
		hostCounter := anyHostCounter.(*WindowCounter)
		sums := hostCounter.SumSince(start)
		if len(sums) == 0 {
			continue
		}
		results[host] = sums
	}
	return results
}

func TrackedResponseWriter(w http.ResponseWriter, readCallback, writeCallback func(int)) http.ResponseWriter {
	wInitial := &wrappedWriter{
		ResponseWriter: w,
		writeCallback:  writeCallback,
		readCallback:   readCallback,
	}

	// Expose http.Hijacker and http.Flusher interfaces if they are present,
	// so that downstream users could fallback if the type cast fails.
	var newW http.ResponseWriter = wInitial
	if _, ok := w.(http.Hijacker); ok {
		hj := hijackWrappedWriter{wrappedWriter: *wInitial}
		if flusher, ok := w.(http.Flusher); ok {
			newW = &flusherHijackWrappedWriter{
				hijackWrappedWriter: hj,
				Flusher:             flusher,
			}
		} else {
			newW = &hj
		}
	} else if flusher, ok := w.(http.Flusher); ok {
		newW = &flusherWrappedWriter{
			wrappedWriter: *wInitial,
			Flusher:       flusher,
		}
	}

	return newW
}

func TrackedRequest(r *http.Request, readCallback func(int)) *http.Request {
	newReq := *r
	newReq.Body = countReadCloser{
		countReader: countReader{
			r:        r.Body,
			callback: readCallback,
		},
		Closer: r.Body,
	}
	return &newReq
}

type wrappedWriter struct {
	http.ResponseWriter
	writeCallback func(int)
	readCallback  func(int)
}

func (w *wrappedWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.writeCallback(n)
	return n, err
}

type hijackWrappedWriter struct {
	wrappedWriter
}

func (h *hijackWrappedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, readWrite, err := h.ResponseWriter.(http.Hijacker).Hijack()
	if err != nil {
		return conn, readWrite, err
	}
	newReadWrite := &bufio.ReadWriter{
		Writer: bufio.NewWriter(&countWriter{w: readWrite.Writer, callback: h.writeCallback}),
		Reader: bufio.NewReader(&countReader{r: readWrite.Reader, callback: h.readCallback}),
	}
	return conn, newReadWrite, err
}

type flusherWrappedWriter struct {
	wrappedWriter
	http.Flusher
}

type flusherHijackWrappedWriter struct {
	hijackWrappedWriter
	http.Flusher
}

type countWriter struct {
	w        *bufio.Writer
	callback func(int)
}

func (c countWriter) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	c.callback(n)
	if err == nil {
		// We take over buffering outside of the writer, so we don't
		// want to allow data to become stuck forever in a duplex
		// communication channel.
		err = c.w.Flush()
	}
	return n, err
}

type countReader struct {
	r        io.Reader
	callback func(int)
}

func (c countReader) Read(p []byte) (n int, err error) {
	n, err = c.r.Read(p)
	c.callback(n)
	return n, err
}

type countReadCloser struct {
	countReader
	io.Closer
}
