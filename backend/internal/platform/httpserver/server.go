// Package httpserver runs HTTP servers with explicit timeouts, health checks and graceful drain.
package httpserver

import (
	"net/http"
	"time"
)

// Timeouts (sources: stack/03-backend-go "Server timeouts").
const (
	ReadHeaderTimeout = 5 * time.Second
	ReadTimeout       = 10 * time.Second
	WriteTimeout      = 15 * time.Second
	IdleTimeout       = 120 * time.Second
	MaxHeaderBytes    = 16 << 10
	DrainTimeout      = 25 * time.Second
)

// New returns a server speaking HTTP/1.1 and h2c.
func New(addr string, h http.Handler) *http.Server {
	var p http.Protocols
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)
	return &http.Server{
		Addr: addr, Handler: h, Protocols: &p,
		ReadHeaderTimeout: ReadHeaderTimeout, ReadTimeout: ReadTimeout, WriteTimeout: WriteTimeout,
		IdleTimeout: IdleTimeout, MaxHeaderBytes: MaxHeaderBytes,
	}
}
