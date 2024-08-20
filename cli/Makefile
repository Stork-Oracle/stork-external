# Makefile

# Variables
OUTPUT_DIR := ./build
OUTPUT_BIN := $(OUTPUT_DIR)/stork

# Targets
.PHONY: all build clean

# The default target that will be executed if no target is specified
all: build

# Build the Go project and place the binary in the build directory
build:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(OUTPUT_BIN)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(OUTPUT_BIN)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_BIN)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o $(OUTPUT_BIN)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -o $(OUTPUT_BIN)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -o $(OUTPUT_BIN)-windows-arm64.exe .
	@echo "Build complete. Binaries are located in the $(OUTPUT_DIR) directory."


# Clean the build directory
clean:
	rm -rf $(OUTPUT_DIR)
