[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/anglo-korean/sparql)
[![Go Report Card](https://goreportcard.com/badge/.)](https://goreportcard.com/report/github.com/anglo-korean/sparql)
# sparql

Package sparql provides a series of parsers for turning SPARQL JSON
(mime type: application/sparql-results+json only- this package doesn't
touch xml) into useful go data types.

Within this package there are two subpackages:

1. github.com/anglo-korean/sparql/bank - a sparql 'query bank' (though realistically I guess this may work for other well-formed data) with `text/template` support
2. github.com/anglo-korean/sparql/repo - a sparql respository client, with various helpers such as auth, and caching

This package is simple to use, stable, and relatively quick. It accepts some json, in `string`, `[]byte`, or `io.Reader` form, and returns some rdf terms:

```go
package main

import (
    "fmt"

    "github.com/anglo-korean/sparql"
)

var data = `{"head":{"vars":["item","itemLabel"]},"results":{"bindings":[{"item":{"type":"uri","value":"[http://www.wikidata.org/entity/Q378619](http://www.wikidata.org/entity/Q378619)"},"itemLabel":{"xml:lang":"en","type":"literal","value":"CC"}},{"item":{"type":"uri","value":"[http://www.wikidata.org/entity/Q498787](http://www.wikidata.org/entity/Q498787)"},"itemLabel":{"xml:lang":"en","type":"literal","value":"Muezza"}}]}}`

func main() {
    res, err := sparql.ParseString(data)
    if err != nil {
        return
    }

    fmt.Printf("%!(NOVERB)v\n", res.Solutions())
}
```

Or, to query data using the provided client:

```go
package main

import (
    "fmt"

    "github.com/anglo-korean/sparql/repo"
)

func main() {
    client, err := repo.New("[https://example.com](https://example.com)")
    if err != nil {
        return
    }

    res, err := client.Query("SELECT * WHERE { ?s ?p ?o } LIMIT 1")
    if err != nil {
        return
    }

    fmt.Printf("%!(NOVERB)v\n", res.Solutions())
}
```

Note that the client provided in repo parses returned json autiomatically.
(Further documentation for both the repo client and the query bank may be found in those specific packages)

Parsed results are parsed into a sparql.Results type, which contains two functions used for accessing data:

1. `res.Bindings()` -> `map[string][]rdf.Term`
2. `res.Solutions()`  -> `[]map[string]rdf.Term`

An example of working with this data may be found in the examples directory

## Sub Packages

* [bank](./bank): Package bank provides a query bank for sparql queries.

* [repo](./repo): Package repo provides a simple http client for interacting with sparql endpoints.

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
