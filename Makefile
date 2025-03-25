.PHONY: build test clean lint example

build:
	go build -o bin/llmscript cmd/llmscript/main.go

test:
	go test ./...

clean:
	rm -rf bin/
	go clean

lint:
	golangci-lint run

example:
	make
	./bin/llmscript --verbose examples/hello-world

.DEFAULT_GOAL := build