# bank

Package bank provides a query bank for sparql queries.

It works by processing data which looks like:

```go
# Comments will be ignored, except those tagging a query

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
```

These queries can then be called by the user specified 'tag', with template values inserted where necessary (via 'text/template')

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
