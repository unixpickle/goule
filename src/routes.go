package goule

import (
	"net/http"
	"net/url"
)

func Route(res http.ResponseWriter, req *http.Request, scheme string,
	overseer *Overseer) {

	// The server won't know if TLS was used, so we need to specify manually.
	url := *req.URL
	url.Scheme = scheme
	url.Host = req.Host

	if handleForwardRule(res, req, url, overseer) {
		return
	} else if handleAdminRule(res, req, url, overseer) {
		return
	}

	// No rules found. In the future, we might consider sending a meaningful 404
	// page.
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte("No forward rule found."))
}

func handleForwardRule(res http.ResponseWriter, req *http.Request, url url.URL,
	overseer *Overseer) bool {
	// TODO: here, check services' forward rules
	return false
}

func handleAdminRule(res http.ResponseWriter, req *http.Request, url url.URL,
	overseer *Overseer) bool {
	for _, source := range overseer.GetAdminSettings().Rules {
		if source.MatchesURL(&url) {
			// Handle the administrative request
			subpath := source.SubpathForURL(&url)
			AdminHandler(res, req, subpath)
			return true
		}
	}
	return false
}
