.PHONY: build test lint all

all: build test lint

build:
	go build ./...

test:
	go test -race ./...

lint:
	golangci-lint run ./...
