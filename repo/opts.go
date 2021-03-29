package repo

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/anglo-korean/digest"
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

// WithDigestAuth configures Repo to use digest authentication on HTTP requests.
func WithDigestAuth(username, password string) func(*Repo) error {
	return func(r *Repo) error {
		r.client.Transport = digest.NewTransport(username, password)
		return nil
	}
}

type basicAuthTransport struct {
	Username string
	Password string
}

func (bat basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
			bat.Username, bat.Password)))))
	return http.DefaultTransport.RoundTrip(req)
}

// WithBasicAuth configures Repo to use basic authentication on HTTP requests.
func WithBasicAuth(username, password string) func(*Repo) error {
	return func(r *Repo) error {
		r.client.Transport = basicAuthTransport{
			Username: username,
			Password: password,
		}
		return nil
	}
}

// WithTimeout instructs the underlying HTTP transport to timeout after given duration.
func WithTimeout(t time.Duration) func(*Repo) error {
	return func(r *Repo) error {
		r.client.Timeout = t
		return nil
	}
}
