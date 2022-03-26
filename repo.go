package sparql

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/knakk/digest"
	"github.com/knakk/rdf"
)

// Provider provides an interface for the
// client to implement a generic request matching
// the sparql endpoint they are using
type Provider interface {
	GenRequest(endpoint string) (*http.Request, error)
}

// GenericCall contains the query to execute and fulfils the
// generic request provider interface thanks to the GenRequest function.
//
// To override the request format, the client must declare a
// new struct which fulfils the Provider interface
type GenericCall struct {
	Query string
}

// Repo represent a RDF repository, assumed to be
// queryable via the SPARQL protocol over HTTP.
type Repo struct {
	endpoint string
	client   *http.Client
}

// NewRepo creates a new representation of a RDF repository. It takes a
// variadic list of functional options which can alter the configuration
// of the repository.
func NewRepo(addr string, options ...func(*Repo) error) (*Repo, error) {
	r := Repo{
		endpoint: addr,
		client:   http.DefaultClient,
	}
	return &r, r.SetOption(options...)
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

// DigestAuth configures Repo to use digest authentication on HTTP requests.
func DigestAuth(username, password string) func(*Repo) error {
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

// BasicAuth configures Repo to use basic authentication on HTTP requests.
func BasicAuth(username, password string) func(*Repo) error {
	return func(r *Repo) error {
		r.client.Transport = basicAuthTransport{
			Username: username,
			Password: password,
		}
		return nil
	}
}

// Timeout instructs the underlying HTTP transport to timeout after given duration.
func Timeout(t time.Duration) func(*Repo) error {
	return func(r *Repo) error {
		r.client.Timeout = t
		return nil
	}
}

// GenRequest is the generic sparql api request
func (call GenericCall) GenRequest(endpoint string) (*http.Request, error) {
	form := url.Values{}
	form.Set("query", call.Query)
	b := form.Encode()

	// TODO make optional GET or Post, Query() should default GET (idempotent, cacheable)
	// maybe new for updates: func (r *Repo) Update(q string) using POST?
	req, err := http.NewRequest(
		"POST",
		endpoint,
		bytes.NewBufferString(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))
	req.Header.Set("Accept", "application/sparql-results+json")

	return req, err
}

// Query performs a SPARQL HTTP request to the Repo, and returns the
// parsed application/sparql-results+json response.
//
// The function accepts either a query in string format, or a struct
// which fulfils the Provider interface and matches the GenericCall struct
func (r *Repo) Query(queryProvider interface{}) (*Results, error) {

	var req *http.Request
	var err error

	if p, ok := queryProvider.(Provider); ok {
		req, err = p.GenRequest(r.endpoint)
	}

	if s, ok := queryProvider.(string); ok {
		var p GenericCall
		p.Query = s
		req, err = p.GenRequest(r.endpoint)
	}

	if err != nil {
		return nil, err
	}

	if req == nil {
		return nil, fmt.Errorf("Request failed: no http.Request provided. ")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("Query: SPARQL request failed: %s. "+msg, resp.Status)
	}
	results, err := ParseJSON(resp.Body)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// QueryWithoutParsing performs query and returns the unparsed reponse as an io.ReadCloser.
// Except on errors, it's the callers responsibility to close the response.
func (r *Repo) QueryWithoutParsing(queryProvider interface{}) (io.ReadCloser, error) {

	var req *http.Request
	var err error

	if p, ok := queryProvider.(Provider); ok {
		req, err = p.GenRequest(r.endpoint)
	}

	if s, ok := queryProvider.(string); ok {
		var p GenericCall
		p.Query = s
		req, err = p.GenRequest(r.endpoint)
	}

	if err != nil {
		return nil, err
	}

	if req == nil {
		return nil, fmt.Errorf("Request failed: no http.Request provided. ")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		var msg string
		if err != nil {
			msg = "Failed to read response body"
		} else {
			if strings.TrimSpace(string(b)) != "" {
				msg = "Response body: \n" + string(b)
			}
		}
		return nil, fmt.Errorf("Query: SPARQL request failed: %s. "+msg, resp.Status)
	}

	return resp.Body, nil
}

// Construct performs a SPARQL HTTP request to the Repo, and returns the
// result triples.
func (r *Repo) Construct(q string) ([]rdf.Triple, error) {
	form := url.Values{}
	form.Set("query", q)
	form.Set("format", "text/turtle")
	b := form.Encode()

	req, err := http.NewRequest(
		"POST",
		r.endpoint,
		bytes.NewBufferString(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))
	req.Header.Set("Accept", "text/turtle")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("Construct: SPARQL request failed: %s. "+msg, resp.Status)
	}
	dec := rdf.NewTripleDecoder(resp.Body, rdf.Turtle)
	return dec.DecodeAll()
}

// Update performs a SPARQL HTTP update request
func (r *Repo) Update(q string) error {
	form := url.Values{}
	form.Set("update", q)
	b := form.Encode()

	req, err := http.NewRequest(
		"POST",
		r.endpoint,
		bytes.NewBufferString(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		b, err := ioutil.ReadAll(resp.Body)
		var msg string
		if err != nil {
			msg = "Failed to read response body"
		} else {
			if strings.TrimSpace(string(b)) != "" {
				msg = "Response body: \n" + string(b)
			}
		}
		return fmt.Errorf("Update: SPARQL request failed: %s. "+msg, resp.Status)
	}
	return nil
}
