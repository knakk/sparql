// Package sparql contains functions and data structures for querying SPARQL
// endpoints and parsing the response into RDF terms, as well as other
// convenience functions for working with SPARQL queries.
package sparql

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/knakk/rdf"
)

// DateFormat is the expected layout of the xsd:DateTime values. You can override
// it if your triple store uses a different layout.
var DateFormat = time.RFC3339

var xsdString rdf.IRI

func init() {
	xsdString, _ = rdf.NewIRI("http://www.w3.org/2001/XMLSchema#string")
}

// Results holds the parsed results of a application/sparql-results+json response.
type Results struct {
	Head    header
	Boolean bool
	Results results
}

type header struct {
	Link []string
	Vars []string
}

type results struct {
	Distinct bool
	Ordered  bool
	Bindings []map[string]binding
}

type binding struct {
	Type     string // "uri", "literal", "typed-literal" or "bnode"
	Value    string
	Lang     string `json:"xml:lang"`
	DataType string
}

// ParseJSON takes an application/sparql-results+json response and parses it
// into a Results struct.
func ParseJSON(r io.Reader) (*Results, error) {
	var res Results
	err := json.NewDecoder(r).Decode(&res)

	return &res, err
}

// Bindings returns a map of the bound variables in the SPARQL response, where
// each variable points to one or more RDF terms.
func (r *Results) Bindings() map[string][]rdf.Term {
	rb := make(map[string][]rdf.Term)
	for _, v := range r.Head.Vars {
		for _, b := range r.Results.Bindings {
			t, err := termFromJSON(b[v])
			if err == nil {
				rb[v] = append(rb[v], t)
			}
		}
	}

	return rb
}

// Solutions returns a slice of the query solutions, each containing a map
// of all bindings to RDF terms.
func (r *Results) Solutions() []map[string]rdf.Term {
	var rs []map[string]rdf.Term

	for _, s := range r.Results.Bindings {
		solution := make(map[string]rdf.Term)
		for k, v := range s {
			term, err := termFromJSON(v)
			if err == nil {
				solution[k] = term
			}
		}
		rs = append(rs, solution)
	}

	return rs
}

// termFromJSON converts a SPARQL json result binding into a rdf.Term. Any
// parsing errors on typed-literal will result in a xsd:string-typed RDF term.
// TODO move this functionality to package rdf?
func termFromJSON(b binding) (rdf.Term, error) {
	switch b.Type {
	case "bnode":
		return rdf.NewBlank(b.Value)
	case "uri":
		return rdf.NewIRI(b.Value)
	case "literal":
		// Untyped literals are typed as xsd:string
		if b.Lang != "" {
			return rdf.NewLangLiteral(b.Value, b.Lang)
		}
		if b.DataType != "" {
			dt, _ := rdf.NewIRI(b.DataType)
			return rdf.NewTypedLiteral(b.Value, dt), nil
		}
		return rdf.NewTypedLiteral(b.Value, xsdString), nil
	case "typed-literal":
		iri, err := rdf.NewIRI(b.DataType)
		if err != nil {
			return nil, err
		}
		return rdf.NewTypedLiteral(b.Value, iri), nil
	default:
		return nil, errors.New("unknown term type")
	}
}
