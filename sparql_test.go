package sparql

import (
	"bytes"
	"testing"

	"github.com/knakk/rdf"
)

const testResults = `
{
   "head": {
       "link": [],
       "vars": [ "x", "hpage", "name", "mbox", "age", "friend", "score", "z", "updated" ]
       },
   "results": {
       "bindings": [
               {
                   "x" : { "type": "bnode", "value": "r1" },
                   "hpage" : { "type": "uri", "value": "http://work.example.org/alice/" },
                   "name" : {  "type": "literal", "value": "Alice" } ,
                   "friend" : { "type": "bnode", "value": "r2" },
                   "age": { "type": "typed-literal", "datatype": "http://www.w3.org/2001/XMLSchema#integer", "value": "17" },
                   "score": { "type": "typed-literal", "datatype": "http://www.w3.org/2001/XMLSchema#float", "value": "0.2" },
                   "z": { "type": "typed-literal", "datatype": "http://www.w3.org/2001/XMLSchema#boolean", "value": "true" },
                   "updated": {
                    "type": "typed-literal",
                    "datatype": "http://www.w3.org/2001/XMLSchema#dateTime",
                    "value": "2014-07-21T04:00:40+02:00"
                  }
               },
               {
                   "x" : { "type": "bnode", "value": "r2" },
                   "hpage" : { "type": "uri", "value": "http://work.example.org/bob/" },
                   "name" : { "type": "literal", "value": "Bob", "xml:lang": "en" },
                   "mbox" : { "type": "uri", "value": "mailto:bob@work.example.org" },
                   "friend" : { "type": "bnode", "value": "r1" },
                   "age": { "type": "typed-literal", "datatype": "http://www.w3.org/2001/XMLSchema#integer", "value": "43" },
                   "score": { "type": "typed-literal", "datatype": "http://www.w3.org/2001/XMLSchema#float", "value": "11.93" },
                   "z": { "type": "typed-literal", "datatype": "http://www.w3.org/2001/XMLSchema#boolean", "value": "false" }
               }
           ]
       }
}`

func TestParseJSON(t *testing.T) {
	b := bytes.NewBufferString(testResults)
	r, err := ParseJSON(b)
	if err != nil {
		t.Fatal(err)
	}

	if len(r.Results.Bindings) != 2 {
		t.Errorf("Got %d solutions, want 2", len(r.Results.Bindings))
	}

	if len(r.Head.Vars) != 9 {
		t.Errorf("Got %d vars in head, want 9", len(r.Head.Vars))
	}
}

func TestBindings(t *testing.T) {
	b := bytes.NewBufferString(testResults)
	r, err := ParseJSON(b)
	if err != nil {
		t.Fatal(err)
	}

	bi := r.Bindings()

	if len(bi) != 9 {
		t.Errorf("Got %d bound variables, want 9", len(bi))
	}

	if len(bi["x"]) != 2 {
		t.Errorf("Got %d bindings for x; want 2", len(bi["x"]))
	}

	if len(bi["updated"]) != 1 {
		t.Errorf("Got %d bindings for updated; want 1", len(bi["updated"]))
	}
}

func TestSolutions(t *testing.T) {
	b := bytes.NewBufferString(testResults)
	r, err := ParseJSON(b)
	if err != nil {
		t.Fatal(err)
	}

	s := r.Solutions()

	blankR1, _ := rdf.NewBlank("r1")
	iriAlice, _ := rdf.NewIRI("http://work.example.org/alice/")
	litAlice, _ := rdf.NewLiteral("Alice")
	litBob, _ := rdf.NewLangLiteral("Bob", "en")
	xsdInteger, _ := rdf.NewIRI("http://www.w3.org/2001/XMLSchema#integer")
	xsdFloat, _ := rdf.NewIRI("http://www.w3.org/2001/XMLSchema#float")
	xsdBoolean, _ := rdf.NewIRI("http://www.w3.org/2001/XMLSchema#boolean")
	xsdDateTime, _ := rdf.NewIRI("http://www.w3.org/2001/XMLSchema#dateTime")

	var tests = []struct {
		got  rdf.Term
		want rdf.Term
	}{
		{s[0]["x"], blankR1},
		{s[0]["hpage"], iriAlice},
		{s[0]["name"], litAlice},
		{s[1]["name"], litBob},
		{s[0]["age"], rdf.NewTypedLiteral("17", xsdInteger)},
		{s[0]["score"], rdf.NewTypedLiteral("0.2", xsdFloat)},
		{s[0]["z"], rdf.NewTypedLiteral("true", xsdBoolean)},
		{s[1]["z"], rdf.NewTypedLiteral("false", xsdBoolean)},
		{s[0]["updated"], rdf.NewTypedLiteral("2014-07-21T04:00:40+02:00", xsdDateTime)},
	}

	for _, tt := range tests {
		if !rdf.TermsEqual(tt.got, tt.want) {
			t.Errorf("Got \"%v\", want \"%v\"", tt.got, tt.want)
		}
	}
}
