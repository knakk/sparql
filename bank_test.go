package sparql

import (
	"bytes"
	"testing"

	"github.com/knakk/specs"
)

const testBank = `
# Some comment; should be ignored
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
`

func TestLoadBank(t *testing.T) {
	s := specs.New(t)

	f := bytes.NewBufferString(testBank)
	b := LoadBank(f)

	s.Expect(3, len(b))
}

func TestBankPrepare(t *testing.T) {
	s := specs.New(t)

	f := bytes.NewBufferString(testBank)
	b := LoadBank(f)

	q0, err := b.Prepare("q0")
	s.ExpectNil(err)
	s.Expect("SELECT * WHERE { ?s ?p ?o } ", q0)

	q1, err := b.Prepare("q1", struct{ Subj string }{"http://example.org/s1"})
	s.ExpectNil(err)
	s.Expect("SELECT * WHERE { ?s ?p ?o FILTER(?s = <http://example.org/s1>) } ", q1)

	q2, err := b.Prepare("q2", struct{ L, O int }{10, 33})
	s.ExpectNil(err)
	s.Expect("SELECT ?s WHERE { ?s ?p ?o } LIMIT 10 OFFSET 33 ", q2)

	_, err = b.Prepare("q3")
	s.ExpectNotNil(err)
}
