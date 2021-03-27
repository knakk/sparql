package sparql

import (
	"github.com/gregjones/httpcache"
)

// WithCache takes a httpcache.Cache, thus providing a way of caching sparql
// queries which may be otherwise slow, or where the data returned changes
// infrequently.
//
// It may be used as:
//
//    cache := httpcache.NewMemoryCache()
//    repo := sparql.NewRepo("localhost:8080/sparql", sparql.WithCache(cache))
//
// This uses the default httpcache in-memory cache in requests
func WithCache(c httpcache.Cache) func(*Repo) error {
	return func(r *Repo) error {
		r.client = httpcache.NewTransport(c).Client()

		return nil
	}
}
