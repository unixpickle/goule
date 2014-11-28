package proxy

import (
	"net/http"
	"net/url"
)

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	ProxyURL *url.URL
	DestURL  *url.URL
	Settings *Settings
}