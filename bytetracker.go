package main

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

type ByteTracker struct {
	hostsLock sync.RWMutex
	hosts     map[string]struct{}

	// Maps from host string to *SingleUsageHistory
	writes sync.Map
	reads  sync.Map
}

func (b *ByteTracker) SetHosts(hosts []string) {
	hostSet := map[string]struct{}{}
	for _, h := range hosts {
		hostSet[h] = struct{}{}
	}
	b.hostsLock.Lock()
	b.hosts = hostSet
	b.hostsLock.Unlock()
}

func (b *ByteTracker) allowHost(h string) bool {
	b.hostsLock.RLock()
	defer b.hostsLock.RUnlock()
	_, ok := b.hosts[h]
	return ok
}

func (b *ByteTracker) Wrap(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if !b.allowHost(host) {
			host = ""
		}
		writeCallback := func(n int) {
			tracker, _ := b.writes.LoadOrStore(host, &SingleUsageHistory{})
			tracker.(*SingleUsageHistory).RecordBytes(n)
		}
		readCallback := func(n int) {
			tracker, _ := b.reads.LoadOrStore(host, &SingleUsageHistory{})
			tracker.(*SingleUsageHistory).RecordBytes(n)
		}

		wInitial := &wrappedWriter{
			ResponseWriter: w,
			writeCallback:  writeCallback,
			readCallback:   readCallback,
		}
		var newW http.ResponseWriter = wInitial
		if _, ok := w.(http.Hijacker); ok {
			// Only advertise hijacking capabilities if they are
			// truly there.
			newW = &hijackWrappedWriter{wrappedWriter: *wInitial}
		}

		newReq := *r
		newReq.Body = countReadCloser{
			countReader: countReader{
				r:        r.Body,
				callback: readCallback,
			},
			Closer: r.Body,
		}
		h.ServeHTTP(newW, &newReq)
	}
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
	newReadWrite := &bufio.ReadWriter{
		Writer: bufio.NewWriter(&countWriter{w: readWrite.Writer, callback: h.writeCallback}),
		Reader: bufio.NewReader(&countReader{r: readWrite.Reader, callback: h.readCallback}),
	}
	return conn, newReadWrite, err
}

type countWriter struct {
	w        *bufio.Writer
	callback func(int)
}

func (c countWriter) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	c.callback(n)
	if err != nil {
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

type HistoryWindow struct {
	Start time.Time
	End   time.Time
	Usage uint64
}

// TODO: implement a single host usage tracker
type SingleUsageHistory struct {
	lock    sync.RWMutex
	history []HistoryWindow
	current *HistoryWindow
}

func (h *SingleUsageHistory) RecordBytes(n int) {
	// TODO: this.
}
