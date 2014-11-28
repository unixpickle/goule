package proxy

import (
	"errors"
	"net"
	"net/http"
)

func ProxyWebsocket(context *Context) {
	if num, err := proxyWebsocketInternal(context); err != nil {
		res := context.Response
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(num)
		res.Write([]byte(err.Error()))
	}
}

func proxyWebsocketInternal(context *Context) (int, error) {
	if hj, ok := context.Response.(http.Hijacker); !ok {
		return http.StatusInternalServerError,
			errors.New("Response does not support hijacking")
	} else {
		conn, err := net.Dial("tcp", context.DestURL.Host)
		if err != nil {
			return http.StatusInternalServerError,
				errors.New("Cannot connect to proxy destination.")
		}
		defer conn.Close()

		// Send the request
		if err := context.Request.Write(conn); err != nil {
			return http.StatusInternalServerError, err
		}

		// Hijack the original connection
		origConn, bf, err := hj.Hijack()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		// Pipe conn and origConn
		Pipe(bf, conn)
		origConn.Close()

		return 0, nil
	}
}
