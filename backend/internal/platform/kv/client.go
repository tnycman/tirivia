// Package kv opens the Valkey client (ElastiCache for Valkey; stack/03).
package kv

import "github.com/valkey-io/valkey-go"

// Open parses a valkey:// or valkeys:// (TLS) URL; client-side caching is off because no caller uses DoCache.
func Open(url string) (valkey.Client, error) {
	opt, err := valkey.ParseURL(url)
	if err != nil {
		return nil, err
	}
	opt.DisableCache = true
	return valkey.NewClient(opt)
}
