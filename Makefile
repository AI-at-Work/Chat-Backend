# Makefile

# Variables
PROTO_DIR := ./proto
PROTO_FILE := $(PROTO_DIR)/*.proto
PROTO_GEN_DIR := ./pb

# Check for protoc
PROTOC := $(shell command -v protoc 2> /dev/null)

# Go specific variables
GO := go
PROTOC_GEN_GO := $(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
PROTOC_GEN_GO_GRPC := $(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: check-protoc clean proto

all: check-protoc clean proto

run: all

# Clean generated files and virtual environment
clean:
	rm -f $(PROTO_GEN_DIR)/*.pb.go

# Check if protoc is installed
check-protoc:
	@if [ -z "$(PROTOC)" ]; then \
		echo "Error: protoc is not installed or not in PATH"; \
		echo "Please install protoc before proceeding:"; \
		echo "  - On Ubuntu/Debian: sudo apt-get install protobuf-compiler"; \
		echo "  - On macOS with Homebrew: brew install protobuf"; \
		echo "  - For other systems, visit: https://grpc.io/docs/protoc-installation/"; \
		exit 1; \
	fi

# Generate Protocol Buffer code
proto:
	@echo "Generating Go gRPC code..."
	$(PROTOC_GEN_GO)
	$(PROTOC_GEN_GO_GRPC)
	protoc -I$(PROTO_DIR) --go_out=$(PROTO_GEN_DIR) --go-grpc_out=$(PROTO_GEN_DIR) $(PROTO_FILE)
