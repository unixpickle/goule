package goule

import (
	"net/url"
	"strconv"
	"strings"
)

type SourceURL struct {
	Scheme   string `json:"scheme"`
	Hostname string `json:"hostname"`
	Path     string `json:"path"`
}

type DestinationURL struct {
	Scheme   string `json:"scheme"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Path     string `json:"path"`
}

type ForwardRule struct {
	From SourceURL      `json:"from"`
	To   DestinationURL `json:"to"`
}

// MatchesURL returns true if the receiving SourceURL matches a specified URL.
func (self SourceURL) MatchesURL(url *url.URL) bool {
	if url.Scheme != self.Scheme || url.Host != self.Hostname {
		return false
	}
	return strings.HasPrefix(url.Path, self.Path)
}

// SubpathForURL returns the path components that a given URL contains that the
// receiving SourceURL does not.
// This subpath can be appended to the proxy destination in some cases.
func (self SourceURL) SubpathForURL(url *url.URL) string {
	return url.Path[len(self.Path):len(url.Path)]
}

// Apply attempts to apply a forward rule to a given URL.
// The return value is the destination to forward to, or nil if the forward
// rule is not applicable.
func (self ForwardRule) Apply(url *url.URL) *URL {
	if !self.From.MatchesURL(url) {
		return nil
	}
	result := *url
	result.Scheme = self.To.Scheme
	result.Host = self.To.Hostname + ":" + strconv.Itoa(self.To.Port)
	result.Path = self.To.Path + self.From.SubpathForURL(url)
	return result
}
