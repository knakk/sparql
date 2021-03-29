default: README.md

# The following tool may be installed with: go get github.com/posener/goreadme/cmd/goreadme
README.md: *.go
	cat .doc/head.md > $@
	goreadme >> $@
