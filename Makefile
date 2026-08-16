.PHONY: all build build-tui clean run install-deps

BINARY_NAME=auror
BINARY_TUI=auror-tui
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

all: build build-tui

install-deps:
	$(GO) mod download
	$(GO) mod verify

build: install-deps
	mkdir -p build
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o build/$(BINARY_NAME) cmd/auror/main.go

build-tui: install-deps
	mkdir -p build
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o build/$(BINARY_TUI) cmd/auror-tui/main.go

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
