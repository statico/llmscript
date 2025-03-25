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

example1:
	make
	./bin/llmscript --debug --no-cache examples/hello-world

example2:
	make
	./bin/llmscript --debug --no-cache examples/product-downloader

.DEFAULT_GOAL := build