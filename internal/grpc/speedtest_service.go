package grpc

import (
	"barry-server-go/internal/config"
	"barry-server-go/internal/service"
	pb "barry-server-go/proto/speedtest"
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

// server implements the SpeedTestServiceServer interface using injected services.
type server struct {
	pb.UnimplementedSpeedTestServiceServer
	config         *config.Config
	serverProvider service.ServerProvider
	downloader     service.DownloadStreamer
	uploader       service.UploadHandler
	ipDetector     service.IPDetector
}

// NewSpeedTestServer creates a new server instance with injected services.
func NewSpeedTestServer(
	cfg *config.Config,
	sp service.ServerProvider,
	ds service.DownloadStreamer,
	uh service.UploadHandler,
	ip service.IPDetector,
) pb.SpeedTestServiceServer {
	return &server{
		config:         cfg,
		serverProvider: sp,
		downloader:     ds,
		uploader:       uh,
		ipDetector:     ip,
	}
}

// GetServers delegates to the ServerProvider service.
func (s *server) GetServers(ctx context.Context, req *pb.GetServersRequest) (*pb.GetServersResponse, error) {
	log.Printf("Handler: Received GetServers request. Limit: %d", req.Limit)
	servers, err := s.serverProvider.GetServers(ctx, req.Limit)
	if err != nil {
		log.Printf("Handler: Error from ServerProvider: %v", err)
		// Map domain errors to gRPC status codes if needed
		return nil, status.Errorf(codes.Internal, "failed to get servers: %v", err)
	}
	return &pb.GetServersResponse{Servers: servers}, nil
}

// Ping remains simple, handled directly.
func (s *server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	// Minimal logic, stays in handler
	serverTimestamp := time.Now().UnixNano() // Example: Use a time source if testing needed
	log.Printf("Handler: Received Ping request from client timestamp: %d", req.ClientTimestampUnixNano)
	return &pb.PingResponse{
		ClientTimestampUnixNano: req.ClientTimestampUnixNano,
		ServerTimestampUnixNano: serverTimestamp,
	}, nil
}

// Download delegates streaming logic to the DownloadStreamer service.
func (s *server) Download(req *pb.DownloadRequest, stream pb.SpeedTestService_DownloadServer) error {
	log.Printf("Handler: Received Download request. Chunk size hint: %d", req.ChunkSizeHintBytes)

	chunkSize := s.config.ChunkSizeBytes
	if req.ChunkSizeHintBytes > 0 && req.ChunkSizeHintBytes < (1024*1024) {
		chunkSize = int(req.ChunkSizeHintBytes)
	}

	// Create a writer that sends data over the gRPC stream
	streamWriter := &grpcStreamWriter{stream: stream}

	// Delegate the core streaming logic
	err := s.downloader.StreamDownloadData(stream.Context(), streamWriter, chunkSize)
	if err != nil {
		log.Printf("Handler: Error from DownloadStreamer: %v", err)
		// The service should ideally return gRPC status errors already
		if _, ok := status.FromError(err); ok {
			return err // Return status error directly
		}
		// Fallback for non-status errors
		return status.Errorf(codes.Internal, "download failed: %v", err)
	}
	log.Println("Handler: Download stream completed successfully.")
	return nil
}

// Upload delegates stream handling to the UploadHandler service.
func (s *server) Upload(stream pb.SpeedTestService_UploadServer) error {
	log.Println("Handler: Received Upload stream request.")

	// Create a reader from the gRPC stream
	streamReader := &grpcStreamReader{stream: stream}

	// Delegate the core upload handling
	bytesReceived, err := s.uploader.HandleUploadStream(stream.Context(), streamReader)

	if err != nil && !errors.Is(err, io.EOF) { // EOF from Recv() is handled by SendAndClose in service
		log.Printf("Handler: Error from UploadHandler: %v", err)
		// Attempt to send response even on error
		_ = stream.SendAndClose(&pb.UploadResponse{BytesReceived: bytesReceived})
		if _, ok := status.FromError(err); ok {
			return err // Return status error directly
		}
		return status.Errorf(codes.Internal, "upload failed: %v", err)
	}

	log.Printf("Handler: Upload stream finished. Total bytes received: %d", bytesReceived)
	// Send the final response
	return stream.SendAndClose(&pb.UploadResponse{BytesReceived: bytesReceived})
}

// GetClientIP delegates to the IPDetector service.
func (s *server) GetClientIP(ctx context.Context, req *pb.GetClientIPRequest) (*pb.GetClientIPResponse, error) {
	log.Println("Handler: Received GetClientIP request.")
	ip, city, country, err := s.ipDetector.DetectClientIP(ctx)
	if err != nil {
		log.Printf("Handler: Error from IPDetector: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to detect client IP: %v", err)
	}
	log.Printf("Handler: Reporting client IP: %s", ip)
	return &pb.GetClientIPResponse{
		ClientIp: ip,
		City:     city,
		Country:  country,
	}, nil
}

// --- Helper types to adapt gRPC streams to io.Writer/io.Reader ---

// grpcStreamWriter adapts SpeedTestService_DownloadServer to io.Writer
type grpcStreamWriter struct {
	stream pb.SpeedTestService_DownloadServer
}

func (w *grpcStreamWriter) Write(p []byte) (n int, err error) {
	err = w.stream.Send(&pb.DataChunk{Data: p})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// grpcStreamReader adapts SpeedTestService_UploadServer to io.Reader
type grpcStreamReader struct {
	stream pb.SpeedTestService_UploadServer
	buffer []byte // Leftover data from previous Recv
}

func (r *grpcStreamReader) Read(p []byte) (n int, err error) {
	// Use leftover data first
	if len(r.buffer) > 0 {
		copied := copy(p, r.buffer)
		r.buffer = r.buffer[copied:]
		return copied, nil
	}

	// Read next chunk from stream
	chunk, err := r.stream.Recv()
	if err != nil {
		// io.EOF is the expected signal for stream end from Recv()
		return 0, err
	}

	// Copy data to caller's buffer 'p', store leftover in internal buffer
	data := chunk.GetData()
	copied := copy(p, data)
	if copied < len(data) {
		r.buffer = data[copied:] // Store leftover
	}
	return copied, nil
}
