BINARY_NAME=main
TARGET ?= linux
ARCH ?= amd64
SRC=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

default: test

run: test
	@echo "== Run =="
	go run main.go

build: test
	@echo "== Build =="
	go build -o $(BINARY_NAME) -v

test:
	@echo "== Test =="
	gofmt -s -l -w $(SRC)
	go vet -v ./...
	go test -race -v ./...

clean:
	@echo "== Cleaning =="
	rm main
	rm concourse-sts-lambda.zip

lint:
	@echo "== Lint =="
	golint

release: build-release
	@echo "== Release build =="
	zip concourse-sts-lambda.zip main

build-release: test
	CGO_ENABLED=0 GOOS=$(TARGET) GOARCH=$(ARCH) go build -o $(BINARY_NAME) -v

.PHONY: default build test release build-release
