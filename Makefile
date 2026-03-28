.PHONY: build run test test-race lint

build:
	go build -o bin/go-propagator ./cmd/go-propagator

run:
	go run ./cmd/go-propagator

test:
	go test ./...

test-race:
	go test -race ./...

lint:
	go vet ./...
