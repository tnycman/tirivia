package main

import (
	"fmt"
	"net/url"
)

// origin requires an absolute http(s) origin with no path, query, fragment or userinfo; it returns scheme://host.
func origin(name, v string) (string, error) {
	u, err := url.Parse(v)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil ||
		u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", fmt.Errorf("%s must be an http(s) origin such as https://api.example.com, got %q", name, v)
	}
	return u.Scheme + "://" + u.Host, nil
}
