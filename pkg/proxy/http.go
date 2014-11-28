package proxy

import (
	"net/http"
)

func ProxyHTTP(context *Context, client *Client) {
	
	
}

func proxyHTTPInternal(context *Context, client *Client) (int, error) {
	req := http.NewRequest(context.Request.Method, context.DestURL.String(),
		context.Request.Body)
	for header, value := range RequestHeaders(context) {
		req.Header[header] = value
	}
	res, err := client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	// TODO: write all this stuff
}