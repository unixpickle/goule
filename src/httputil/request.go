package httputil

import (
	"errors"
	"net/http"
)

// ReadRequest reads an http.Request and returns the buffer.
// If the specified limit is exceeded, an error will be returned.
func ReadRequest(req *http.Request, limit int) ([]byte, error) {
	data := []byte{}
	for {
		next := make([]byte, 0x1000)
		num, err := req.Body.Read(next)
		data = append(data, next[0:num]...)
		if len(data) > limit {
			return nil, errors.New("Request exceeded limit.")
		}
		if err != nil {
			break
		}
	}
	return data, nil
}
