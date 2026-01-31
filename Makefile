.PHONY: build clean install all test

BUILD_DIR := bin
BINARY := sse
VERSION := $(shell grep 'var Version' cmd/root.go | cut -d'"' -f2)

build:
	go build -o $(BUILD_DIR)/$(BINARY) .

test:
	go test -v -race ./...

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY) ~/bin/$(BINARY)

# Cross-compilation targets (static linking, no external dependencies)
build-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/windows-amd64/$(BINARY).exe .

build-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/darwin-amd64/$(BINARY) .

build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/darwin-arm64/$(BINARY) .

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/linux-amd64/$(BINARY) .

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/linux-arm64/$(BINARY) .

# Build all platforms
all: build-windows-amd64 build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64
