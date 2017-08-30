all: test cover lint

ci: test coveralls lint

test:
	go list -f '{{if len .TestGoFiles}}"go test -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' ./... | xargs -L 1 sh -c

gover:
	gover

cover: gover
	go tool cover -html=gover.coverprofile

coveralls: gover
  goveralls -coverprofile=gover.coverprofile -service=travis-ci

lint:
	go vet ./...
	golint ./...
	gofmt -d -s `go list -f '{{.Dir}}' ./...`

deps:
	go get github.com/modocache/gover
	go get github.com/mattn/goveralls
	go get github.com/golang/lint/golint

.PHONY: deps all ci test cover coveralls lint
