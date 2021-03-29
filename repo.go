package sparql

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Repo represent a RDF repository, assumed to be
// queryable via the SPARQL protocol over HTTP.
type Repo struct {
	endpoint string
	client   *http.Client
}

// NewRepo creates a new representation of a RDF repository. It takes a
// variadic list of functional options which can alter the configuration
// of the repository.
func NewRepo(addr string, options ...func(*Repo) error) (r *Repo, err error) {
	r = &Repo{
		endpoint: addr,
		client:   http.DefaultClient,
	}

	err = r.SetOption(options...)

	return
}

// SetOption takes one or more option function and applies them in order to Repo.
func (r *Repo) SetOption(options ...func(*Repo) error) error {
	for _, opt := range options {
		if err := opt(r); err != nil {
			return err
		}
	}
	return nil
}

// Query performs a SPARQL HTTP request to the Repo, and returns the
// parsed application/sparql-results+json response.
//
// These lookups are expected to be idempotent, and as such use http.MethodGet requests.
// See: Repo.Update for requests which use http.MethodPost
func (r Repo) Query(query string) (*Results, error) {
	return r.query(http.MethodGet, query)
}

// Update performs a SPARQL HTTP request to the Repo, and returns the
// parsed application/sparql-results+json response.
//
// These queries are made via an http.MethodPost and, so, are expected to
// change state.
//
// Functionally these requests differ very little from requests made via Repo.Query;
// and in fact  there's nothing that says a POST'd request *must* update state.
// The difference is purely to allow for the caching of GET requests
func (r Repo) Update(query string) (*Results, error) {
	return r.query(http.MethodPost, query)
}

// query performs the actual heavy lifting of querying an endpoint and parsing
// responses
func (r Repo) query(verb, query string) (res *Results, err error) {
	req, err := createRequest(r.endpoint, verb, query)
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		var msg string

		if err != nil {
			msg = "Failed to read response body"
		} else {
			if strings.TrimSpace(string(b)) != "" {
				msg = "Response body: \n" + string(b)
			}
		}

		return nil, fmt.Errorf("sparql request failed: %s. %s", resp.Status, msg)
	}

	return ParseJSON(resp.Body)
}

// createRequest creates an *http.Request from the endpoint url, the http verb,
// and the query.
//
// Where verb is http.MethodPost, the query is form encoded, whereas where the verb
// is http.MethodGet the query is encoded as a query parameter
//
// This function returns an error for other verbs; I don't /think/ SPARQL generic
// endpoints use any other verb, and PRs are welcomed
func createRequest(endpoint, verb, query string) (req *http.Request, err error) {
	form := url.Values{}
	form.Set("query", query)
	b := form.Encode()

	switch verb {
	case http.MethodGet:
		var u *url.URL

		u, err = url.Parse(endpoint)
		if err != nil {
			return
		}

		u.RawQuery = b
		endpoint = u.String()

	case http.MethodPost:
		// nop

	default:
		err = fmt.Errorf("unable to handle sparql queries with http verb %q", verb)

		return
	}

	req, err = http.NewRequest(
		verb,
		endpoint,
		bytes.NewBufferString(b))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))
	req.Header.Set("Accept", "application/sparql-results+json")

	return
}
