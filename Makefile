BINARY_NAME=main
TARGET ?= linux
ARCH ?= amd64
SRC=$(shell find . -type f -name '*.go' -not -path "./vendor/*")
DIR=$(shell pwd)

default: test

generate:
	@echo "== Go Generate =="
	go generate ./...

run: test
	@echo "== Run =="
	go run cmd/main.go

build: test
	@echo "== Build =="
	go build -o $(BINARY_NAME) -v cmd/main.go

clean:
	@echo "== Cleaning =="
	rm $(BINARY_NAME) || true
	rm concourse-sts-lambda.zip || true

release:
	@echo "== Release build =="
	CGO_ENABLED=0 GOOS=$(TARGET) GOARCH=$(ARCH) go build -o $(BINARY_NAME) -v cmd/main.go
	zip concourse-sts-lambda.zip main

test-code:
	@echo "== Test =="
	gofmt -s -l -w $(SRC)
	go vet -v ./...
	go test -race -v ./...

test: test-code
	@echo "== Terraform tests =="
	@cd terraform; \
	if ! terraform fmt -write=false -check=true >> /dev/null; then \
		echo "✗ terraform fmt (Some files need to be formatted, run 'terraform fmt' to fix.)"; \
		exit 1; \
	fi
	@echo "√ terraform fmt"
	@cd $(DIR)

	@for d in $$(find . -type f -name '*.tf' -path "./terraform/modules/*" -not -path "**/.terraform/*" -exec dirname {} \; | sort -u); do \
		cd $$d; \
		terraform init -backend=false >> /dev/null; \
		terraform validate -check-variables=false; \
		if [ $$? -eq 1 ]; then \
			echo "✗ terraform validate failed: $$d"; \
			exit 1; \
		fi; \
		cd $(DIR); \
	done
	@echo "√ terraform validate modules (not including variables)"; \

	@cd terraform; \
	terraform init -backend=false >> /dev/null; \
	terraform validate; \
	if [ $$? -eq 1 ]; then \
		echo "✗ terraform validate failed: $$d"; \
		exit 1; \
	fi
	@echo "√ terraform validate example"
	@cd $(DIR)

.PHONY: default build test release test-code generate
