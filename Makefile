# Makefile for barry-server

# Variables
BINARY_NAME=barry-server
CMD_PATH=./cmd/server
PROTO_SRC_DIR=api/proto
PROTO_OUT_DIR=proto
PROTO_FILES=$(wildcard $(PROTO_SRC_DIR)/*.proto)
GO_FILES=$(shell find . -name '*.go' -not -path "./proto/*") # Source files excluding generated proto

# Go parameters
GO=go
GO_BUILD=$(GO) build
GO_CLEAN=$(GO) clean
GO_TEST=$(GO) test
GO_FMT=$(GO) fmt
GO_RUN=$(GO) run

# Build flags
# -w suppresses DWARF debugging information
# -s strips the symbol table and debugging information
LDFLAGS = -ldflags="-w -s"

# Protoc parameters
PROTOC=protoc
PROTOC_GO_OUT_FLAG=--go_out=. --go_opt=paths=source_relative
PROTOC_GRPC_OUT_FLAG=--go-grpc_out=. --go-grpc_opt=paths=source_relative

# Docker parameters
DOCKER=docker
IMAGE_NAME=barry-server
IMAGE_TAG=latest

# Default target executed when you run `make`
.DEFAULT_GOAL := build

# Targets
.PHONY: all proto build run test fmt clean docker-build docker-run help

all: build

# Generate Go code from .proto files
# This target runs protoc on all .proto files found in PROTO_SRC_DIR
# The generated files in PROTO_OUT_DIR depend on the source .proto files
$(PROTO_OUT_DIR)/%.pb.go $(PROTO_OUT_DIR)/%.pb.grpc.go: $(PROTO_FILES)
	@echo "Generating Go code from Protobuf definitions..."
	@mkdir -p $(PROTO_OUT_DIR)
	$(PROTOC) $(PROTOC_GO_OUT_FLAG) $(PROTOC_GRPC_OUT_FLAG) $(PROTO_FILES)
	@echo "Protobuf code generation complete."

proto: $(PROTO_OUT_DIR)/speedtest.pb.go $(PROTO_OUT_DIR)/speedtest_grpc.pb.go ## Generate Go code from proto files

# Build the Go binary
# Depends on the Go source files and the generated proto files
$(BINARY_NAME): $(GO_FILES) $(PROTO_OUT_DIR)/speedtest.pb.go $(PROTO_OUT_DIR)/speedtest_grpc.pb.go
	@echo "Building Go binary..."
	$(GO_BUILD) $(LDFLAGS) -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BINARY_NAME)"

build: $(BINARY_NAME) ## Build the Go application binary

# Run the server locally
# Requires environment variables to be set (see README.md)
run: build ## Run the server locally (requires environment variables)
	@echo "Running $(BINARY_NAME) locally..."
	@echo "Ensure required environment variables (e.g., PUBLIC_URL) are set."
	./$(BINARY_NAME)

# Run Go tests
test: ## Run Go unit and integration tests
	@echo "Running tests..."
	$(GO_TEST) -v ./...

# Format Go code
fmt: ## Format Go source code
	@echo "Formatting Go code..."
	$(GO_FMT) ./...

# Clean build artifacts and generated code
clean: ## Remove build artifacts and generated code
	@echo "Cleaning..."
	$(GO_CLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(PROTO_OUT_DIR) # Remove generated proto directory
	@echo "Clean complete."

# Build the Docker image
docker-build: ## Build the Docker image
	@echo "Building Docker image $(IMAGE_NAME):$(IMAGE_TAG)..."
	$(DOCKER) build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "Docker image built."

# Run the application inside a Docker container
# Requires environment variables to be passed using -e flags
docker-run: ## Run the application in a Docker container
	@echo "Running Docker container $(IMAGE_NAME):$(IMAGE_TAG)..."
	@echo "Ensure required environment variables (e.g., PUBLIC_URL) are passed via -e flags."
	$(DOCKER) run --rm -p 8080:8080 --name barry-server-container \
		-e LISTEN_ADDRESS=":8080" \
		-e SERVER_ID="barry-go-docker-make" \
		-e PUBLIC_URL="host.docker.internal:8080" \
		$(IMAGE_NAME):$(IMAGE_TAG)

# Display help message
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

