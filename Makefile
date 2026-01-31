.PHONY: build clean install

BUILD_DIR := bin
BINARY := sse

build:
	go build -o $(BUILD_DIR)/$(BINARY) .

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY) ~/bin/$(BINARY)
