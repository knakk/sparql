package sparql

import (
	"bytes"
	"testing"
)

const testBank = `
# Comments will be ignored, excepts those tagging a query

# tag: q0
SELECT * WHERE { ?s ?p ?o }

# tag: q1
SELECT *
WHERE
 {
  ?s ?p ?o
  FILTER(?s = <{{.Subj}}>)
 }

# another comment

# tag: q2
SELECT ?s
WHERE { ?s ?p    ?o }
LIMIT {{.L}}
OFFSET {{.O}}

# tag: myq
SELECT *
WHERE {
	{ <{{.Res}}> ?p ?o }
	UNION
	{ ?s ?p <{{.Res}}> }
}
`

func TestLoadBank(t *testing.T) {
	f := bytes.NewBufferString(testBank)
	b := LoadBank(f)

	if len(b) != 4 {
		t.Errorf("len(bank) => %d, want 4", len(b))
	}
}

func TestBankPrepare(t *testing.T) {
	f := bytes.NewBufferString(testBank)
	b := LoadBank(f)

	q0, err := b.Prepare("q0")
	if err != nil {
		t.Fatal(err)
	}
	want := "SELECT * WHERE { ?s ?p ?o } "
	if q0 != want {
		t.Errorf("got %v, want %v", q0, want)
	}

	q1, err := b.Prepare("q1", struct{ Subj string }{"http://example.org/s1"})
	if err != nil {
		t.Fatal(err)
	}
	want = "SELECT * WHERE { ?s ?p ?o FILTER(?s = <http://example.org/s1>) } "
	if q1 != want {
		t.Errorf("got %v, want %v", q1, want)
	}

	q2, err := b.Prepare("q2", struct{ L, O int }{10, 33})
	if err != nil {
		t.Fatal(err)
	}
	want = "SELECT ?s WHERE { ?s ?p ?o } LIMIT 10 OFFSET 33 "
	if q2 != want {
		t.Errorf("got %v, want %v", q2, want)
	}

	q3, err := b.Prepare("myq", struct{ Res string }{"http://r.com"})
	if err != nil {
		t.Fatal(err)
	}
	want = "SELECT * WHERE { { <http://r.com> ?p ?o } UNION { ?s ?p <http://r.com> } } "
	if q3 != want {
		t.Errorf("got %v, want %v", q3, want)
	}

	_, err = b.Prepare("q3")
	if err == nil {
		t.Error("calling prepare() with a non-existing query should result in an error")
	}

}
