package config

import (
	"net/url"
	"testing"
)

func TestMatchesURL(t *testing.T) {
	source := SourceURL{"http", "localhost", "/foo"}
	ensureMatch(t, "http://localhost/foo", source)
	ensureMatch(t, "http://localhost/foo/bar", source)
	ensureMatch(t, "http://localhost/foo/", source)
	ensureMatch(t, "http://localhost:1337/foo/", source)
	ensureMismatch(t, "http://localhost/foobar", source)
	ensureMismatch(t, "https://localhost/foo", source)
	ensureMismatch(t, "http://localhost1/foo", source)
	
	source = SourceURL{"https", "aqnichol.com", ""}
	ensureMatch(t, "https://aqnichol.com", source)
	ensureMatch(t, "https://aqnichol.com/", source)
	ensureMatch(t, "https://aqnichol.com/foo/bar", source)
	ensureMatch(t, "https://aqnichol.com:1337/foo/bar", source)
	ensureMismatch(t, "http://aqnichol.com", source)
	ensureMismatch(t, "https://www.aqnichol.com", source)
	ensureMismatch(t, "http://aqnichol.net", source)
	
	source = SourceURL{"https", "aqnichol.com", "/"}
	ensureMismatch(t, "http://aqnichol.com", source)
}

func TestSubpathForURL(t *testing.T) {
	source := SourceURL{"http", "localhost", "/foo"}
	ensureSubpath(t, "http://localhost/foo", source, "")
	ensureSubpath(t, "http://localhost/foo/", source, "/")
	ensureSubpath(t, "http://localhost/foo/bar", source, "/bar")
	
	source = SourceURL{"http", "aqnichol.com", ""}
	ensureSubpath(t, "http://aqnichol.com", source, "")
	ensureSubpath(t, "http://aqnichol.com/", source, "/")
	ensureSubpath(t, "http://aqnichol.com/foo", source, "/foo")
	
	source = SourceURL{"http", "aqnichol.com", "/"}
	ensureSubpath(t, "http://aqnichol.com/", source, "")
	ensureSubpath(t, "http://aqnichol.com/foo", source, "foo")
}

func TestApply(t *testing.T) {
	source := SourceURL{"https", "aqnichol.com", "/foo"}
	dest := DestinationURL{"http", "localhost", 1337, ""}
	rule := ForwardRule{source, dest}
	ensureApply(t, "https://aqnichol.com/foo", rule, "http://localhost:1337")
	ensureApply(t, "https://aqnichol.com/foo/", rule, "http://localhost:1337/")
	ensureApply(t, "https://aqnichol.com/foo/bar", rule,
		"http://localhost:1337/bar")
	rule.From = SourceURL{"https", "aqnichol.com", ""}
	rule.To = DestinationURL{"http", "localhost", 80, "/foo"}
	ensureApply(t, "https://aqnichol.com", rule, "http://localhost:80/foo")
	ensureApply(t, "https://aqnichol.com/", rule, "http://localhost:80/foo/")
	ensureApply(t, "https://aqnichol.com/a", rule, "http://localhost:80/foo/a")
}

func ensureMatch(t *testing.T, urlStr string, source SourceURL) {
	if parsed, err := url.Parse(urlStr); err == nil {
		matches := source.MatchesURL(parsed)
		if !matches {
			t.Error("URL does not match: '" + urlStr + "'")
		}
	} else {
		t.Error("Failed to parse '" + urlStr + "'")
	}
}

func ensureMismatch(t *testing.T, urlStr string, source SourceURL) {
	if parsed, err := url.Parse(urlStr); err == nil {
		matches := source.MatchesURL(parsed)
		if matches {
			t.Error("URL matches: '" + urlStr + "'")
		}
	} else {
		t.Error("Failed to parse '" + urlStr + "'")
	}
}

func ensureSubpath(t *testing.T, urlStr string, source SourceURL,
	expect string) {
	if parsed, err := url.Parse(urlStr); err == nil {
		if sp := source.SubpathForURL(parsed); sp != expect {
			t.Error("Expected subpath '" + expect + "', got '" + sp + "'")
		}
	} else {
		t.Error("Failed to parse '" + urlStr + "'")
	}
}

func ensureApply(t *testing.T, urlStr string, rule ForwardRule, expect string) {
	if parsed, err := url.Parse(urlStr); err == nil {
		res := rule.Apply(parsed)
		if res == nil {
			t.Error("Failed to apply rule to '" + urlStr + "'")
		} else {
			check, err := url.Parse(expect)
			if err != nil {
				t.Error("Failed to parse '" + expect + "'")
				return
			}
			if check.Path != res.Path {
				t.Error("Expected path '" + check.Path + "', got '" + res.Path +
					"'")
			}
			if check.Scheme != res.Scheme {
				t.Error("Expected scheme '" + check.Scheme + "', got '" +
					res.Scheme + "'")
			}
			if check.Host != res.Host {
				t.Error("Expected host '" + check.Host + "', got '" + res.Host +
					"'")
			}
		}
	} else {
		t.Error("Failed to parse '" + urlStr + "'")
	}
}