.PHONY: build test clean lint

build:
	go build -o bin/llmscript cmd/llmscript/main.go

test:
	go test ./...

clean:
	rm -rf bin/
	go clean

lint:
	golangci-lint run

.DEFAULT_GOAL := build