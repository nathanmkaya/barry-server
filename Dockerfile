# Dockerfile for barry-server (Go gRPC)

# Stage 1: Build the application binary
# Use a specific Go version matching your development environment
FROM golang:1.22-alpine AS builder

# Set necessary environment variables
ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /app

# Copy module files first to leverage Docker build cache
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the application source code
# Copy api/proto for completeness, though generation often happens outside build
COPY api ./api
COPY cmd ./cmd
COPY internal ./internal
# Copy generated proto code if it's checked into VCS (optional, depends on workflow)
# COPY proto ./proto

# Build the Go application
# -ldflags="-w -s" reduces binary size by removing debug symbols
RUN go build -ldflags="-w -s" -o /barry-server ./cmd/server

# Stage 2: Create the final minimal image
# Use scratch for the smallest possible image size.
# Alternatively, use gcr.io/distroless/static-debian11 if you need CA certificates, timezone data, etc.
FROM scratch

# Copy the static binary built in the previous stage
COPY --from=builder /barry-server /barry-server

# Copy any other necessary non-code assets (e.g., config files if not using env vars)
# COPY --from=builder /app/config.yaml /config.yaml

# Expose the default gRPC port the application listens on
# Make sure this matches the LISTEN_ADDRESS port configured (e.g., :8080)
EXPOSE 8080

# Set the entrypoint for the container to run the server binary
ENTRYPOINT ["/barry-server"]

# Optional: Define default command arguments if your app uses flags
# CMD ["--config", "/config.yaml"]
