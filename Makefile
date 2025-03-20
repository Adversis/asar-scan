.PHONY: all clean build-darwin-amd64 build-darwin-arm64 build-windows-amd64 dist

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR = build
DIST_DIR = dist

all: clean build dist

clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	
build: build-darwin-amd64 build-darwin-arm64 build-windows-amd64

build-darwin-amd64:
	@echo "Building for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/asarscan-darwin-amd64 cmd/asarscan/*.go

build-darwin-arm64:
	@echo "Building for macOS (arm64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/asarscan-darwin-arm64 cmd/asarscan/*.go

build-windows-amd64:
	@echo "Building for Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/asarscan-windows-amd64.exe cmd/asarscan/*.go

dist: build
	@echo "Creating distribution packages..."
	@mkdir -p $(DIST_DIR)
	cd $(BUILD_DIR) && tar -czf ../$(DIST_DIR)/asarscan-darwin-amd64-$(VERSION).tar.gz asarscan-darwin-amd64
	cd $(BUILD_DIR) && tar -czf ../$(DIST_DIR)/asarscan-darwin-arm64-$(VERSION).tar.gz asarscan-darwin-arm64
	cd $(BUILD_DIR) && zip -q ../$(DIST_DIR)/asarscan-windows-amd64-$(VERSION).zip asarscan-windows-amd64.exe

# For local testing
run:
	go build -o asarscan cmd/asarscan/*.go
	./asarscan
