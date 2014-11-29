package proxy

import "net/http"

// RequestHeaders adds "x-forwarded-*" headers to a request while removing
// hop-by-hop headers.
func RequestHeaders(context *Context) http.Header {
	result := http.Header{}

	// Copy all regular headers
	for header, values := range context.Request.Header {
		if !IsHopByHop(header, context.Settings) && !IsForwardedHeader(header) {
			result[header] = values
		}
	}

	// Set X-Forwarded-* headers.
	// If the incoming request already had one of these headers, commas are used
	// to add the new values to the existing ones.
	forwarded := map[string]string{"For": context.Request.RemoteAddr,
		"Host": context.ProxyURL.Host, "Proto": context.ProxyURL.Scheme}
	for _, suffix := range []string{"For", "Host", "Proto"} {
		header := "X-Forwarded-" + suffix
		var value string
		if existing, ok := context.Request.Header[header]; ok {
			value = existing[0] + ", " + forwarded[suffix]
		} else {
			value = forwarded[suffix]
		}
		result[header] = []string{value}
	}
	return result
}

// ResponseHeaders rewrites the response headers from an HTTP proxy target.
// In the future, this may rewrite the "Location" header.
// This removes hop-by-hop headers.
func ResponseHeaders(context *Context, headers http.Header) http.Header {
	result := http.Header{}

	// Copy all regular headers
	for header, values := range headers {
		if !IsHopByHop(header, context.Settings) {
			result[header] = values
		}
	}

	return result
}

// IsHopByHop checks if a header is a hop-by-hop header that should be removed
// from the proxied request.
// If websocket is true, the Connection and Upgrade headers won't be considered
// as hop-by-hop headers.
func IsHopByHop(header string, settings *Settings) bool {
	var hopByHop []string
	if settings.Websockets {
		hopByHop = []string{"Keep-Alive", "Proxy-Authenticate",
			"Proxy-Authorization", "Te", "Transfer-Encoding"}
	} else {
		hopByHop = []string{"Connection", "Keep-Alive", "Proxy-Authenticate",
			"Proxy-Authorization", "Te", "Transfer-Encoding", "Upgrade"}
	}
	for _, val := range hopByHop {
		if val == header {
			return true
		}
	}
	return false
}

// IsForwardedHeader returns true if and only if the passed header is
// "X-Forwarded-" + suffix where suffix is "For", "Host", or "Proto".
func IsForwardedHeader(header string) bool {
	return header == "X-Forwarded-For" || header == "X-Forwarded-Proto" ||
		header == "X-Forwarded-Host"
}
