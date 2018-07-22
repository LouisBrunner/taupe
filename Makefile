all: test cover lint

build:
	go build ./cmd/taupe

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
	golint `go list -f '{{.Dir}}' ./...`
	gofmt -d -s `go list -f '{{.Dir}}' ./...`

deps:
	go get -u github.com/modocache/gover github.com/mattn/goveralls github.com/golang/lint/golint

.PHONY: all ci test gover cover coveralls lint deps
