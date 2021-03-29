package main

import (
	"bytes"
	"flag"
	"fmt"

	"github.com/anglo-korean/rdf"
	"github.com/anglo-korean/sparql/bank"
	"github.com/anglo-korean/sparql/repo"
	"github.com/gregjones/httpcache"
)

var (
	name = flag.String("n", "Bill Gates", "Name of the person to lookup")
)

const queries = `
# tag: byName
SELECT ?companyLabel ?officerNameLabel ?starttime ?endtime
WHERE {
  ?company ( p:P169 | p:P488 ) ?officer.

  ?officer ( ps:P169 | ps:P488 ) ?officerName.
  ?officerName rdfs:label '{{ .Name }}'@en.

  OPTIONAL {?officer pq:P580 ?starttime.}
  OPTIONAL {?officer pq:P582 ?endtime.}


  SERVICE wikibase:label { bd:serviceParam wikibase:language "en". }
}
`

func main() {
	flag.Parse()

	f := bytes.NewBufferString(queries)
	queryBank := bank.Load(f)

	q, err := queryBank.Prepare("byName", map[string]string{"Name": *name})
	if err != nil {
		panic(err)
	}

	cache := httpcache.NewMemoryCache()
	repo, err := repo.New("https://query.wikidata.org/bigdata/namespace/wdq/sparql", repo.WithCache(cache))
	if err != nil {
		panic(err)
	}

	res, err := repo.Query(q)
	if err != nil {
		panic(err)
	}

	for idx, solution := range res.Solutions() {
		fmt.Printf("%d: %q, from %s to %s\n",
			idx, valStr(solution, "companyLabel"), valStr(solution, "starttime"), valStr(solution, "endtime"),
		)
	}
}

func valStr(m map[string]rdf.Term, k string) string {
	i, ok := m[k]
	if !ok {
		return ""
	}

	return i.String()
}
