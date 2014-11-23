package goule

import (
	"net/url"
	"strconv"
	"strings"
)

// SourceURL represents a URL which goule receives a request on.
type SourceURL struct {
	Scheme   string `json:"scheme"`
	Hostname string `json:"hostname"`
	Path     string `json:"path"`
}

// DestinationURL represents a URL which goule can proxy to.
type DestinationURL struct {
	Scheme   string `json:"scheme"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Path     string `json:"path"`
}

// ForwardRule specifies a SourceURL to forward to a DestinationURL.
type ForwardRule struct {
	From SourceURL      `json:"from"`
	To   DestinationURL `json:"to"`
}

// MatchesURL returns true if and only if the receiving SourceURL matches the
// specified URL.
func (self SourceURL) MatchesURL(url *url.URL) bool {
	hostname := url.Host
	idx := strings.Index(hostname, ":")
	if idx != -1 {
		hostname = hostname[0:idx]
	}
	if url.Scheme != self.Scheme || hostname != self.Hostname {
		return false
	}
	return strings.HasPrefix(url.Path, self.Path)
}

// SubpathForURL returns the subpath that a given URL contains that a SourceURL
// does not.
func (self SourceURL) SubpathForURL(url *url.URL) string {
	return url.Path[len(self.Path):]
}

// Apply attempts to apply a forward rule to a given URL.
// The return value is the destination URL if the forward rule matches the
// specified URL. Otherwise, Apply returns nil.
func (self ForwardRule) Apply(url *url.URL) *url.URL {
	if !self.From.MatchesURL(url) {
		return nil
	}
	result := *url
	result.Scheme = self.To.Scheme
	result.Host = self.To.Hostname + ":" + strconv.Itoa(self.To.Port)
	result.Path = self.To.Path + self.From.SubpathForURL(url)
	return &result
}
