package goule

import (
	"net/http"
	"net/url"
)

type RouteRequest struct {
	Response   http.ResponseWriter
	Request    *http.Request
	URL        url.URL
	Overseer   *Overseer
	Authorized bool
	AdminPath  string
}

func NewRouteRequest(res http.ResponseWriter, req *http.Request,
	overseer *Overseer, scheme string) *RouteRequest {
	// The server won't know if TLS was used, so we need to specify manually.
	url := *req.URL
	url.Scheme = scheme
	url.Host = req.Host
	return &RouteRequest{res, req, url, overseer, false, ""}
}

func Route(req *RouteRequest) {
	if handleForwardRule(req) {
		return
	} else if handleAdminRule(req) {
		return
	}

	// No rules found. In the future, we might consider sending a meaningful 404
	// page.
	req.Response.Header().Set("Content-Type", "text/plain")
	req.Response.Write([]byte("No forward rule found."))
}

func handleForwardRule(req *RouteRequest) bool {
	// TODO: here, check services' forward rules
	return false
}

func handleAdminRule(req *RouteRequest) bool {
	for _, source := range req.Overseer.GetAdminSettings().Rules {
		if source.MatchesURL(&req.URL) {
			// Configure the administrative fields of the RouteRequest
			req.AdminPath = source.SubpathForURL(&req.URL)
			cookie, _ := req.Request.Cookie("goule_id")
			if cookie != nil {
				if req.Overseer.GetSessions().Validate(cookie.Value) {
					req.Authorized = true
				}
			}
			AdminHandler(req)
			return true
		}
	}
	return false
}
