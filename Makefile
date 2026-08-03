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
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) cmd/auror/main.go

build-full: install-deps build-ui
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) cmd/auror/main.go

clean:
	rm -f $(BINARY_NAME)

clean-all: clean
	# Complete cleanup
	rm -f $(BINARY_NAME)
	rm -Rf reports

run: build
	./$(BINARY_NAME)

lint:
	golangci-lint run

fmt:
	$(GO) fmt ./...
	$(GO) vet ./...

.DEFAULT_GOAL := build
