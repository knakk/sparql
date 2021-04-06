package repo

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gregjones/httpcache"
)

// testCache implements the httpcache.Cache interface (https://pkg.go.dev/github.com/gregjones/httpcache#Cache)
type testCache struct {
	resp []byte
}

func (c testCache) Get(key string) (responseBytes []byte, ok bool) {
	return c.resp, true
}

func (c *testCache) Set(key string, responseBytes []byte) {
	c.resp = responseBytes
}

func (testCache) Delete(_ string) {}

type dummyRT struct {
	req      *http.Request
	status   int
	err      bool
	reqCount int
	body     []byte
}

func (r *dummyRT) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	r.reqCount += 1
	resp = new(http.Response)
	resp.StatusCode = r.status

	if r.body != nil && len(r.body) > 0 {
		resp.Body = io.NopCloser(bytes.NewReader(r.body))
	}

	r.req = req

	if r.err {
		err = fmt.Errorf("an error")
	}

	return
}

func TestWithCache(t *testing.T) {
	jsonBody, err := ioutil.ReadFile("../testdata/sparql_cache_response.json")
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}

	happyPathCache := &testCache{
		resp: []byte(fmt.Sprintf(`HTTP/1.1 200
content-type: application/sparql-results+json
X-From-Cache: 1
Expires: Thu, 06 Dec 9999 15:30:07 UTC
date: Sat, 27 Mar 2021 08:35:34 GMT

%s
`, jsonBody))}

	for _, test := range []struct {
		name        string
		cache       httpcache.Cache
		expectError bool
	}{
		{"Request exists in cache", happyPathCache, false},
		{"Not in cache, request 404s", httpcache.NewMemoryCache(), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo, _ := New("https://example.com/404", WithCache(test.cache))

			_, err := repo.Query("SELECT * WHERE { ?s ?p ?o } LIMIT 1")
			if test.expectError && err == nil {
				t.Errorf("expected error")
			} else if !test.expectError && err != nil {
				t.Errorf("unexpected error: %#v", err)
			}
		})
	}
}

func TestWithHeader(t *testing.T) {
	repo, _ := New("https://example.com/404")

	rt := &dummyRT{status: 200}
	repo.client.Transport = rt

	repo.SetOption(WithHeader("hello", "world"))

	repo.Query("SELECT * WHERE { ?s ?p ?o } LIMIT 1")

	got := rt.req.Header.Get("hello")

	if got != "world" {
		t.Errorf(`expected "world", recived %q`, got)
	}
}

func TestWithCacheAndWithHeader(t *testing.T) {
	jsonBody, err := ioutil.ReadFile("../testdata/sparql_cache_response.json")
	if err != nil {
		t.Fatalf("unexpected error: %#v", err)
	}

	happyPathCache := &testCache{
		resp: []byte(fmt.Sprintf(`HTTP/1.1 200
content-type: application/sparql-results+json
Cache-Control: max-age=30
date: %s

%s
`, time.Now().Add(0-(time.Hour*24)).Format(http.TimeFormat), jsonBody))}

	// Requests should be made, with a max age of 1 week. This should
	// still serve requests from the cache, even though the cached response
	// has a max age of 30 seconds, and is an hour old.
	//
	// Let's see.....
	repo, _ := New("https://example.com/", WithCache(happyPathCache), WithHeader("cache-control", "max-age=86400"))
	_, err = repo.Query("hahaha this is not real sparql <3")

	// if we receive an error, which we will if the above query actually hits
	// a real server (at the very least we expect a 404 from that example.com
	// url) then we can assume our headers don't work- we should always get a
	// valid response from the cache
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
}
