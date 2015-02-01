// Package sparql contains functions and data structures for querying SPARQL
// endpoints and parsing the response into RDF terms, as well as other
// convenience functions for working with SPARQL queries.
package sparql

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/knakk/rdf"
	"github.com/knakk/rdf/xsd"
)

// DateFormat is the expected layout of the xsd:DateTime values. You can override
// it if your triple store uses a different layout.
var DateFormat = time.RFC3339

// Results holds the parsed results of a application/sparql-results+json response.
type Results struct {
	Head    header
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
		return rdf.Blank{ID: b.Value}, nil
	case "uri":
		return rdf.IRI{IRI: b.Value}, nil
	case "literal":
		// Untyped literals are typed as xsd:string
		if b.Lang != "" {
			return rdf.Literal{Val: b.Value, Lang: b.Lang, DataType: xsd.String}, nil
		}
		return rdf.Literal{Val: b.Value, DataType: xsd.String}, nil
	case "typed-literal":
		switch b.DataType {
		case xsd.String.IRI:
			return rdf.Literal{Val: b.Value, DataType: xsd.String}, nil
		case xsd.Integer.IRI:
			i, err := strconv.Atoi(b.Value)
			if err != nil {
				return rdf.Literal{Val: b.Value, DataType: xsd.String}, nil
			}
			return rdf.Literal{Val: i, DataType: xsd.Integer}, nil
		case xsd.Float.IRI:
			f, err := strconv.ParseFloat(b.Value, 64)
			if err != nil {
				return rdf.Literal{Val: b.Value, DataType: xsd.String}, nil
			}
			return rdf.Literal{Val: f, DataType: xsd.Float}, nil
		case xsd.Boolean.IRI:
			bo, err := strconv.ParseBool(b.Value)
			if err != nil {
				return rdf.Literal{Val: b.Value, DataType: xsd.String}, nil
			}
			return rdf.Literal{Val: bo, DataType: xsd.Boolean}, nil
		case xsd.DateTime.IRI:
			t, err := time.Parse(DateFormat, b.Value)
			if err != nil {
				return rdf.Literal{Val: b.Value, DataType: xsd.String}, nil
			}
			return rdf.Literal{Val: t, DataType: xsd.DateTime}, nil
		// TODO: other xsd dataypes
		// TODO: custom datatypes
		default:
			return rdf.Literal{Val: b.Value, DataType: xsd.String}, nil
		}
	default:
		return nil, errors.New("unknown term type")
	}
}
