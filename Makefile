TRAVIS_TAG ?= $(shell git describe --tags --candidates=1 --dirty 2>/dev/null || echo "dev")
BINARY      = main
RELEASE     = $(TRAVIS_TAG).zip
TARGET     ?= linux
ARCH       ?= amd64

SRC = $(filter-out vendor/*, $(wildcard *.go))
DIR = $(shell pwd)

export GO111MODULE=on

default: test

clean:
	@echo "== Cleaning =="
	rm -rf build

build: build/$(RELEASE)
build/$(RELEASE): $(SRC)
	@echo "== Build =="
	CGO_ENABLED=0 GOOS=$(TARGET) GOARCH=$(ARCH) go build -o build/$(BINARY) -v cmd/main.go
	zip -mj build/$(RELEASE) build/$(BINARY)
	
	# Create a copy of the zip with a static filename for uploading to github releases
	cp build/$(RELEASE) build/concourse-sts-lambda.zip

generate: $(SRC)
	@echo "== Go Generate =="
	go generate ./...

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

test-code:
	@echo "== Test =="
	gofmt -s -l -w $(SRC)
	go vet -v ./...
	go test -race -v ./...

.PHONY: default build test test-code generate
