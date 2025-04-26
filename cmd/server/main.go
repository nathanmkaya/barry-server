package main

import (
	"barry-server-go/internal/config"
	grpcHandler "barry-server-go/internal/grpc"
	"barry-server-go/internal/service"
	pb "barry-server-go/proto/speedtest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting server with config: %+v", cfg)

	// --- Create Service Instances ---
	serverProvider := service.NewSimpleServerProvider(cfg)
	downloadStreamer := service.NewDefaultDownloadStreamer(cfg)
	uploadHandler := service.NewDefaultUploadHandler()
	ipDetector := service.NewDefaultIPDetector()
	// --- End Service Instances ---

	// Create TCP listener
	lis, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", cfg.ListenAddress, err)
	}
	defer lis.Close()
	log.Printf("Server listening on %s", cfg.ListenAddress)

	// Create gRPC server instance
	grpcServer := grpc.NewServer()

	// Create and register service implementation, injecting dependencies
	speedTestServer := grpcHandler.NewSpeedTestServer(
		cfg,
		serverProvider,
		downloadStreamer,
		uploadHandler,
		ipDetector,
	)
	pb.RegisterSpeedTestServiceServer(grpcServer, speedTestServer)

	// Optional: Enable server reflection
	reflection.Register(grpcServer)
	log.Println("Server reflection enabled.")

	// Start gRPC server in a separate goroutine
	go func() {
		log.Println("Starting gRPC server...")
		if err := grpcServer.Serve(lis); err != nil {
			// Avoid Fatalf in goroutine if graceful shutdown is also logging
			log.Printf("ERROR: Failed to serve gRPC: %v", err)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("Server gracefully stopped.")
}
