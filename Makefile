.PHONY: build clean install all

BUILD_DIR := bin
BINARY := sse
VERSION := $(shell grep 'var Version' cmd/root.go | cut -d'"' -f2)

build:
	go build -o $(BUILD_DIR)/$(BINARY) .

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY) ~/bin/$(BINARY)

# Cross-compilation targets
build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/windows-amd64/$(BINARY).exe .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/darwin-amd64/$(BINARY) .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/darwin-arm64/$(BINARY) .

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/linux-amd64/$(BINARY) .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/linux-arm64/$(BINARY) .

# Build all platforms
all: build-windows-amd64 build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64
