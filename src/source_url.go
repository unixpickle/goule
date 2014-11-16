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

func (self SourceURL) MatchesURL(url *url.URL) bool {
	if url.Scheme != self.Protocol || url.Host != self.Hostname {
		return false
	}
	return strings.HasPrefix(url.Path, self.Path)
}
