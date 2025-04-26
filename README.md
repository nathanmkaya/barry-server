# **Barry Server**

`barry-server` is the Go gRPC backend implementation for the Barry speed test system. It provides the necessary endpoints for clients (like the Barry KMP SDK) to perform network latency, download, and upload speed measurements.

This server is designed with separation of concerns, leveraging Go's standard library, gRPC, and Protocol Buffers.

## **Prerequisites**

* Go (version 1.18+ recommended)  
* Protocol Buffer Compiler (`protoc`) v3+  
* Go gRPC plugins:  
  * `protoc-gen-go`: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest  `
  * `protoc-gen-go-grpc`: `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest  `
* Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is in your system's `PATH`.  
* (Optional) Docker and Docker Compose for containerized running.

## **Project Structure**

The project follows a standard Go layout with separation between the application entry point (`cmd`), internal logic (`internal`), generated code (`proto`), and API definitions (`api`).

For a detailed breakdown, see [PROJECT\_STRUCTURE.md](docs/PROJECT_STRUCTURE.md). (You would create this file in a docs directory).

## **Getting Started**

**1\. Clone the Repository:**

```bash
git clone \<your-repository-url\>  
cd barry-server
```

2\. Initialize Module (if not already done):  
Replace `github.com/nathanmkaya/barry-server` with your actual module path. 
```bash
go mod init github.com/nathanmkaya/barry-server
```
3\. Generate Protobuf Code:  
This step compiles the .proto definitions into Go code used by the server and potentially clients. Run from the project root:  
```bash
protoc --go_out=. --go_opt=paths=source_relative \  
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \  
       api/proto/speedtest.proto
```

**4\. Install Dependencies:**

```bash
go mod tidy
```

## **Building**

To build the server executable:

```bash
go build \-o barry-server ./cmd/server
```

This creates an executable named `barry-server` in the project root.

## **Configuration**

The server is configured primarily through environment variables. See `internal/config/config.go` for details. Key variables include:

* `LISTEN_ADDRESS`: The address and port the gRPC server listens on (e.g., `:8080, 0.0.0.0:9090`). Default: `:8080`.  
* `SERVER_ID`: A unique identifier for this server instance (e.g., `barry-go-lon-1`). Default: `default-go-server`.  
* `PUBLIC_URL`: **Crucial:** The publicly accessible gRPC address (hostname/IP and port) that clients will use to connect back to *this specific server* for tests (e.g., `speedtest-lon-1.barry.example.com:443`). Default: `localhost:8080`.  
* `CHUNK_SIZE_BYTES`: Default chunk size for download data generation. Default: `65536` (64KB).  
* *(Add others like REGION, CITY, COUNTRY if implemented)*

## **Running**

**Locally:**

1. Set environment variables:
2. ```bash
   export LISTEN_ADDRESS=":8080"  
   export SERVER_ID="barry-go-dev-1"  
   export PUBLIC_URL="localhost:8080" # Adjust if running in different network setup  
   # Set other variables as needed
    ```

2. Run the executable: 
3. ```bash
   ./barry-server
    ```

**Using Docker:**

1. Build the Docker image: 
2. ```bash
   docker build \-t barry-server:latest .
    ```

2. Run the container:  
3. ```bash
   docker run -p 8080:8080 \  
          -e LISTEN_ADDRESS=":8080" \  
          -e SERVER\_ID="barry-go-docker-1" \  
          -e PUBLIC_URL="host.docker.internal:8080"  `# Or your machine's IP if needed by client` \  
          --name barry-srv \  
          barry-server:latest
    ```

   *(Note: `PUBLIC_URL` needs careful consideration depending on how your client running outside Docker will reach the server inside Docker).*

## **Testing**

The project includes unit and integration tests.

* **Unit Tests:** Test individual services and utilities. Mock dependencies where appropriate. Run using `go test ./...`.  
* **Integration Tests:** Test the gRPC handlers by injecting mock services and using an in-memory gRPC server (`bufconn`). See `internal/grpc/*_test.go`. Run using `go test ./...`.

## **Contributing**

*(Add contribution guidelines here if applicable)*

## **License**

*(Specify project license, e.g., MIT, Apache 2.0)*