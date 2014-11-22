package goule

import (
	"net/http"
	"net/url"
)

const SessionIdCookie = "goule_id"

type RouteRequest struct {
	Response   http.ResponseWriter
	Request    *http.Request
	URL        url.URL
	Overseer   *Overseer
	Authorized bool
	AdminPath  string
	AdminRule  SourceURL
}

func NewRouteRequest(res http.ResponseWriter, req *http.Request,
	overseer *Overseer, scheme string) *RouteRequest {
	// The server won't know if TLS was used, so we need to specify manually.
	url := *req.URL
	url.Scheme = scheme
	url.Host = req.Host
	return &RouteRequest{res, req, url, overseer, false, "", SourceURL{}}
}

func Route(req *RouteRequest) {
	if !handleAdminRule(req) {
		if !handleForwardRule(req) {
			// TODO: send a nice 404 page here.
			req.Response.Header().Set("Content-Type", "text/plain")
			req.Response.Write([]byte("No forward rule found."))
		}
	}
}

func handleForwardRule(req *RouteRequest) bool {
	// TODO: here, check services' forward rules
	return false
}

func handleAdminRule(req *RouteRequest) bool {
	for _, source := range req.Overseer.GetAdminSettings().Rules {
		if source.MatchesURL(&req.URL) {
			// Configure the administrative fields of the RouteRequest
			req.AdminRule = source
			req.AdminPath = source.SubpathForURL(&req.URL)
			cookie, _ := req.Request.Cookie(SessionIdCookie)
			if cookie != nil {
				if req.Overseer.GetSessions().Validate(cookie.Value) {
					req.Authorized = true
					http.SetCookie(req.Response, cookie)
				}
			}
			if !RouteAdminSite(req) {
				RouteAdminAPI(req)
			}
			return true
		}
	}
	return false
}
