package bank

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"text/template"
)

// Bank is a map of SPARQL queries.
type Bank map[string]string

var (
	commentMatcher = regexp.MustCompile(`^#`)
	tagMatcher     = regexp.MustCompile(`^#\s*tag:\s+([\w-]+)\s*$`)
	spaceMatcher   = regexp.MustCompile(`\s{2,}`)
)

// Load takes an io.Reader, parses its input and extracts SPARQL queries
// and stores them in a bank. Any query must be preceded by a comment which
// tags the query with a name.
func Load(r io.Reader) (b Bank) {
	b = make(map[string]string)

	s := bufio.NewScanner(r)

	var (
		keepLineState = false
		buf           bytes.Buffer
		key           string
	)

	for s.Scan() {
		line := s.Text()

		if tagMatcher.MatchString(line) {
			keyCandidate := tagMatcher.FindStringSubmatch(line)[1]
			if keyCandidate != key && keepLineState {
				b[key] = buf.String()
				buf.Reset()
			}

			key = keyCandidate
			keepLineState = true
		}

		if keepLineState && !commentMatcher.MatchString(line) {
			buf.WriteString(line)
			buf.WriteString(" ") // s.Scan() strips newlines, so ensure we got a whitespace instead
			b[key] = buf.String()
		}
	}

	for k := range b {
		b[k] = stripLine(b[k])
	}

	return
}

// Prepare returns the query string given a key, and optionally a struct with
// exported fields to be interpolated as variables into the query.
func (b Bank) Prepare(key string, i ...interface{}) (string, error) {
	if q, ok := b[key]; ok {
		if len(i) == 0 {
			return q, nil
		}

		t, err := template.New("query").Parse(q)
		if err != nil {
			return "", err
		}

		var b bytes.Buffer
		err = t.Execute(&b, i[0])

		if err != nil {
			return "", err
		}

		return b.String(), nil
	}

	return "", fmt.Errorf("no query with key %v", key)
}

// stripLine strips excessive whitespace from a string
func stripLine(l string) string {
	return spaceMatcher.ReplaceAllString(l, " ")
}
