package sparql

import (
	"fmt"
	"io/ioutil"
	"testing"

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

func (_ testCache) Delete(_ string) {}

func TestWithCache(t *testing.T) {
	jsonBody, err := ioutil.ReadFile("testdata/sparql_cache_response.json")
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
			repo, _ := NewRepo("https://example.com/404", WithCache(test.cache))

			_, err := repo.Query("SELECT * WHERE { ?s ?p ?o } LIMIT 1")
			if test.expectError && err == nil {
				t.Errorf("expected error")
			} else if !test.expectError && err != nil {
				t.Errorf("unexpected error: %#v", err)
			}
		})
	}
}
