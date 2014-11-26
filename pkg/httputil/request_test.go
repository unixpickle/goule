package httputil

import (
	"bytes"
	"net/http"
	"testing"
)

type Buff struct {
	*bytes.Reader
}

func (self Buff) Close() error {
	return nil
}

func TestReadRequest(t *testing.T) {
	buff := make([]byte, 0x1800)
	for i := 0; i < len(buff); i++ {
		buff[i] = byte(i % 17)
	}
	reader := Buff{bytes.NewReader(buff)}
	req := new(http.Request)
	req.Body = reader

	// Exactly enough room
	data, err := ReadRequest(req, len(buff))
	if !bytes.Equal(data, buff) {
		t.Error("Data differs with exact length")
	}

	// More than enough room
	reader.Seek(0, 0)
	data, err = ReadRequest(req, len(buff)+1)
	if !bytes.Equal(data, buff) {
		t.Error("Data differs with extra length")
	}

	// Not enough room
	reader.Seek(0, 0)
	data, err = ReadRequest(req, len(buff)-1)
	if data != nil || err == nil {
		t.Error("Unexpected result from underflow")
	}
}
