package proxy

import (
	"io"
	"sync"
)

func Pipe(a io.ReadWriter, b io.ReadWriter) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		io.Copy(b, a)
		wg.Add(-1)
	}()
	go func() {
		io.Copy(a, b)
		wg.Add(-1)
	}()
	wg.Wait()
}
