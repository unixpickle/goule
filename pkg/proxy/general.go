package proxy

import "net/http"

func ProxyRequest(context *Context, client *http.Client) {
	if context.Request.Header.Get("Upgrade") == "websocket" {
		ProxyWebsocket(context)
	} else {
		ProxyHTTP(context, client)
	}
}