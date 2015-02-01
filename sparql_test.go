package sparql

import (
	"bytes"
	"testing"
	"time"

	"github.com/knakk/rdf"
	"github.com/knakk/rdf/xsd"
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
	DateFormat = "2006-01-02T15:04:05-07:00"
	rdf.DateFormat = DateFormat
	loc, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		got  rdf.Term
		want rdf.Term
	}{
		{s[0]["x"], rdf.Blank{ID: "r1"}},
		{s[0]["hpage"], rdf.IRI{IRI: "http://work.example.org/alice/"}},
		{s[0]["name"], rdf.Literal{Val: "Alice", DataType: xsd.String}},
		{s[1]["name"], rdf.NewLangLiteral("Bob", "en")},
		{s[0]["age"], rdf.Literal{Val: 17, DataType: xsd.Integer}},
		{s[0]["score"], rdf.Literal{Val: 0.2, DataType: xsd.Float}},
		{s[0]["z"], rdf.Literal{Val: true, DataType: xsd.Boolean}},
		{s[1]["z"], rdf.Literal{Val: false, DataType: xsd.Boolean}},
		{s[0]["updated"], rdf.Literal{Val: time.Date(2014, time.July, 21, 04, 0, 40, 0, loc), DataType: xsd.DateTime}},
	}

	for _, tt := range tests {
		if tt.got.String() != tt.want.String() {
			t.Errorf("Got \"%v\", want \"%v\"", tt.got, tt.want)
		}
	}
}
