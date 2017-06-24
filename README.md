# sparql

Go package that contains functions and data structures for querying SPARQL endpoints and parsing the response into RDF terms, as well as other convenience functions for working with SPARQL queries.

## Interacting with SPARQL endpoints



This snippet creates a RDF repository instance to interact with a SPARQL endpoint running on localhost, set up with HTTP digest authentication and a timeout of 1.5s:

```go
repo, err := sparql.NewRepo("http://localhost:8890/sparql-auth",
  sparql.DigestAuth("dba", "dba"),
  sparql.Timeout(time.Millisecond*1500),
)
if err != nil {
	log.Fatal(err)
}
```

Issue a SPARQL query to the repository:
```go
res, err := repo.Query("SELECT * WHERE { ?s ?p ?o } LIMIT 1")
if err != nil {
	log.Fatal(err)
}
```

See also the section below on using a query bank.

## Working with SPARQL result sets

The results returned by `Query` is a struct corresponding to the [`application/sparql-results+json`](http://www.w3.org/TR/rdf-sparql-json-res/)-data as returned by the SPARQL endpoint. To further work with the result set in [`rdf.Term`](https://github.com/knakk/rdf) format you can call either of theese two methods on the results, `res` being the result returned by `Query`:

- `res.Results.Bindings()` -> `map[string][]rdf.Term`

  `Bindings` returns a map of the bound variables in the SPARQL response, where each variable points to one or more RDF terms.

- `res.Results.Solutions()`  -> `[]map[string]rdf.Term`

  `Solutions` returns a slice of the query solutions, each containing a map of all bindings to RDF terms.

## Query bank

The package includes a query bank implementation. Write all your query templates in string or in a separate file if you like, and tag each query with a name. You can then easily prepare queries by using the `Prepare` method along with an anonymous struct with variables to interpolate into the query.

Example usage:

```go
const queries = `
# Comments are ignored, except those tagging a query.

# tag: my-query
SELECT *
WHERE {
  ?s ?p ?o
} LIMIT {{.Limit}} OFFSET {{.Offset}}
`

f := bytes.NewBufferString(queries)
bank := LoadBank(f)

q, err := bank.Prepare("my-query", struct{ Limit, Offset int }{10, 100})
if err != nil {
	log.Fatal(err)
}


fmt.Println(q)
// Will print: "SELECT * WHERE { ?s ?p ?o } LIMIT 10 OFFSET 100"

```


