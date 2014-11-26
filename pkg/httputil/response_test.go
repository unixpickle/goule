package httputil

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	writer := NewDummyWriter()
	RespondJSON(writer, 404, "hey there bro")
	if writer.status != 404 {
		t.Error("Unexpected status code.")
	}
	var strObj string
	if err := json.Unmarshal(writer.data, &strObj); err != nil {
		t.Error("Unable to unmarshal encoded string")
	}
	if strObj != "hey there bro" {
		t.Error("Unmarshaled string does not match.")
	}

	// Attempt to send a function (which can't be marshaled)
	writer = NewDummyWriter()
	unsendable := func() {}
	RespondJSON(writer, 404, unsendable)
	if writer.status != 404 {
		t.Error("Unexpected status code.")
	}
	writer = NewDummyWriter()
	RespondJSON(writer, 200, unsendable)
	if writer.status != http.StatusInternalServerError {
		t.Error("Unexpected status code.")
	}
}

type DummyWriter struct {
	status int
	data   []byte
	header http.Header
}

func NewDummyWriter() *DummyWriter {
	res := new(DummyWriter)
	res.data = []byte{}
	res.header = http.Header{}
	return res
}

func (self *DummyWriter) Header() http.Header {
	return self.header
}

func (self *DummyWriter) Write(data []byte) (int, error) {
	self.data = append(self.data, data...)
	return len(data), nil
}

func (self *DummyWriter) WriteHeader(code int) {
	self.status = code
}
