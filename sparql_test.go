package sparql

import (
	"bytes"
	"testing"
	"time"

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
	rdf.DateFormat = "2006-01-02T15:04:05-07:00"
	loc, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		got  rdf.Term
		want rdf.Term
	}{
		{s[0]["x"], rdf.NewBlankUnsafe("r1")},
		{s[0]["hpage"], rdf.NewURIUnsafe("http://work.example.org/alice/")},
		{s[0]["name"], rdf.NewLiteralUnsafe("Alice")},
		{s[1]["name"], rdf.NewLangLiteral("Bob", "en")},
		{s[0]["age"], rdf.NewLiteralUnsafe(17)},
		{s[0]["score"], rdf.NewLiteralUnsafe(0.2)},
		{s[0]["z"], rdf.NewLiteralUnsafe(true)},
		{s[1]["z"], rdf.NewLiteralUnsafe(false)},
		{s[0]["updated"], rdf.NewLiteralUnsafe(
			time.Date(2014, time.July, 21, 04, 0, 40, 0, loc))},
	}

	for _, tt := range tests {
		if !tt.got.Eq(tt.want) {
			t.Errorf("Got \"%v\", want \"%v\"", tt.got, tt.want)
		}
	}
}
