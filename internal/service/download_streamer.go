package service

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"

	"barry-server-go/internal/config"
	"barry-server-go/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type defaultDownloadStreamer struct {
	cfg        *config.Config
	dataReader io.Reader // Source of data (e.g., ConstantReader)
}

func NewDefaultDownloadStreamer(cfg *config.Config) DownloadStreamer {
	// Initialize the data source here
	dataReader := utils.NewConstantReader(cfg.ChunkSizeBytes)
	return &defaultDownloadStreamer{
		cfg:        cfg,
		dataReader: dataReader,
	}
}

func (s *defaultDownloadStreamer) StreamDownloadData(ctx context.Context, writer io.Writer, chunkSize int) error {
	log.Printf("Service: Streaming download data with chunk size: %d", chunkSize)
	buffer := make([]byte, chunkSize)

	for {
		// Check context cancellation *before* potentially blocking read/write
		select {
		case <-ctx.Done():
			log.Printf("Service: Download context cancelled.")
			// Map context error to gRPC status
			return status.FromContextError(ctx.Err()).Err()
		default:
			// Continue if context is not cancelled
		}

		// Read from the source
		n, readErr := s.dataReader.Read(buffer)
		if readErr != nil && readErr != io.EOF { // EOF is unexpected for ConstantReader
			log.Printf("Service: Error reading from data source: %v", readErr)
			return status.Errorf(codes.Internal, "internal data source error: %v", readErr)
		}
		if n == 0 { // Should not happen with ConstantReader unless chunksize is 0
			continue
		}

		// Write to the output stream (gRPC stream writer)
		_, writeErr := writer.Write(buffer[:n])
		if writeErr != nil {
			log.Printf("Service: Error writing to download stream: %v", writeErr)
			// Handle common client disconnection errors
			if errors.Is(writeErr, io.EOF) || strings.Contains(writeErr.Error(), "broken pipe") || strings.Contains(writeErr.Error(), "connection reset by peer") {
				return status.Error(codes.Canceled, "client disconnected during download")
			}
			return status.Errorf(codes.Internal, "failed to write download data: %v", writeErr)
		}

		// Flush if the writer supports it (gRPC stream writer usually handles buffering)
		if flusher, ok := writer.(interface{ Flush() }); ok {
			flusher.Flush()
		}
	}
	// This loop continues until context is cancelled or an error occurs
}
