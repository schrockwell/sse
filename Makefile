.PHONY: build clean install all test notarize-darwin

BUILD_DIR := bin
BINARY := sse
VERSION := $(shell grep 'var Version' cmd/root.go | cut -d'"' -f2)

# Code signing / notarization (Darwin). The identity is matched by substring
# against the keychain, so the team ID suffix isn't needed. Override on the
# command line if needed:
#   make build-darwin-arm64 CODESIGN_IDENTITY="Developer ID Application: ..."
# One-time setup for notarization credentials:
#   xcrun notarytool store-credentials $(NOTARY_PROFILE) --apple-id <apple-id> --team-id <team-id>
CODESIGN_IDENTITY := Developer ID Application: Rockwell Schrock
NOTARY_PROFILE := sse-notary

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
	codesign --force --sign "$(CODESIGN_IDENTITY)" --options runtime --timestamp $(BUILD_DIR)/darwin-amd64/$(BINARY)

build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/darwin-arm64/$(BINARY) .
	codesign --force --sign "$(CODESIGN_IDENTITY)" --options runtime --timestamp $(BUILD_DIR)/darwin-arm64/$(BINARY)

# Notarize the signed darwin binaries so Gatekeeper allows them on any Mac.
# Requires stored credentials (see NOTARY_PROFILE setup above). Bare binaries
# can't be stapled, so Gatekeeper verifies the ticket online on first run.
notarize-darwin: build-darwin-amd64 build-darwin-arm64
	ditto -c -k $(BUILD_DIR)/darwin-amd64/$(BINARY) $(BUILD_DIR)/$(BINARY)-darwin-amd64.zip
	ditto -c -k $(BUILD_DIR)/darwin-arm64/$(BINARY) $(BUILD_DIR)/$(BINARY)-darwin-arm64.zip
	xcrun notarytool submit $(BUILD_DIR)/$(BINARY)-darwin-amd64.zip --keychain-profile "$(NOTARY_PROFILE)" --wait
	xcrun notarytool submit $(BUILD_DIR)/$(BINARY)-darwin-arm64.zip --keychain-profile "$(NOTARY_PROFILE)" --wait

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/linux-amd64/$(BINARY) .

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/linux-arm64/$(BINARY) .

# Build all platforms
all: build-windows-amd64 build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64
