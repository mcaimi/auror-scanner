.PHONY: all build clean run install-deps

BINARY_NAME=auror
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

all: build

install-deps:
	$(GO) mod download
	$(GO) mod verify

build: install-deps
	mkdir -p build
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o build/$(BINARY_NAME) cmd/auror/main.go

clean:
	rm -Rf build

clean-all: clean
	# Complete cleanup
	rm -Rf build
	rm -Rf reports

run: build
	./$(BINARY_NAME)

lint:
	golangci-lint run

fmt:
	$(GO) fmt ./...
	$(GO) vet ./...

.DEFAULT_GOAL := build
