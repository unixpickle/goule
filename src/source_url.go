package goule

import (
	"net/url"
	"strings"
)

type SourceURL struct {
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Path     string `json:"path"`
}

// MatchesURL returns true if the receiving SourceURL matches a specified URL.
func (self SourceURL) MatchesURL(url *url.URL) bool {
	if url.Scheme != self.Protocol || url.Host != self.Hostname {
		return false
	}
	return strings.HasPrefix(url.Path, self.Path)
}

// SubpathForURL returns the path components that a given URL contains that the
// receiving SourceURL does not.
// This subpath can be appended to the proxy destination in some cases.
func (self SourceURL) SubpathForURL(url *url.URL) string {
	return url.Path[len(self.Path) : len(url.Path)]
}
