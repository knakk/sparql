package sparql

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	// TODO find the perfect http client, needs digest auth
	"github.com/mreiferson/go-httpclient"
)

// Repo inteface for RDF repositories
type Repo interface {
	Query(q string) ([]byte, error)
}

// RemoteRepo queries via a remote SPARQL endpoint
type RemoteRepo struct {
	endpoint  string
	transport *httpclient.Transport
}

// Query the remote endpoint via a HTTP request
func (r RemoteRepo) Query(q string) ([]byte, error) {
	client := &http.Client{Transport: r.transport}

	payload := url.Values{}
	payload.Set("query", q)
	payload.Add("format", "application/sparql-results+json")

	req, err := http.NewRequest("POST", r.endpoint,
		bytes.NewBufferString(payload.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// NewRemoteRepo instanciates a remote RDF repository
func NewRemoteRepo(endpoint string) *RemoteRepo {
	return &RemoteRepo{
		endpoint: endpoint,
		transport: &httpclient.Transport{
			ConnectTimeout:        1 * time.Second,
			RequestTimeout:        3 * time.Second,
			ResponseHeaderTimeout: 2 * time.Second,
		}}
}
