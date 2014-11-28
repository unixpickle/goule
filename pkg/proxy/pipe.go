package proxy

import (
	"io"
	"sync"
)

func Pipe(a io.ReadWriter, b io.ReadWriter) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		io.Copy(bf, conn)
		wg.Add(-1)
	}
	go func() {
		io.Copy(conn, bf)
		wg.Add(-1)
	}
	wg.Wait()
}
