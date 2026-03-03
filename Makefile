.PHONY: build test clean web

build:
	go build -o bin/issue2md ./cmd/issue2md

web:
	go build -o bin/issue2mdweb ./cmd/issue2mdweb

test:
	go test ./...

clean:
	rm -rf bin/
