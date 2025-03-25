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
	./bin/llmscript --debug examples/hello-world

.DEFAULT_GOAL := build