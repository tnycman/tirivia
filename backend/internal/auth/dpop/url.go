package dpop

import (
	"net/url"
	"strings"
)

// sameURL compares scheme, host and path per RFC 3986 §6.2.2–6.2.3, ignoring query and fragment.
func sameURL(a, b string) bool {
	ua, err1 := url.Parse(a)
	ub, err2 := url.Parse(b)
	if err1 != nil || err2 != nil || ua.Scheme == "" || ua.Host == "" {
		return false
	}
	return normalize(ua) == normalize(ub)
}

func normalize(u *url.URL) string {
	scheme := strings.ToLower(u.Scheme)
	host, port := strings.ToLower(u.Hostname()), u.Port()
	if (scheme == "https" && port == "443") || (scheme == "http" && port == "80") {
		port = ""
	}
	if port != "" {
		host += ":" + port
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	return scheme + "://" + host + path
}
