package service

import (
	"context"
	"io"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type defaultUploadHandler struct {
	// Add dependencies if needed (e.g., config for buffer size)
}

func NewDefaultUploadHandler() UploadHandler {
	return &defaultUploadHandler{}
}

func (h *defaultUploadHandler) HandleUploadStream(ctx context.Context, reader io.Reader) (int64, error) {
	log.Println("Service: Handling upload stream.")
	// Use io.Copy to discard data efficiently. It handles context cancellation implicitly
	// if the underlying reader/writer respects the context (gRPC streams do).
	bytesReceived, err := io.Copy(io.Discard, reader)
	if err != nil {
		// Check if the error is due to context cancellation
		if ctx.Err() != nil {
			log.Printf("Service: Upload context cancelled during copy: %v", ctx.Err())
			return bytesReceived, status.FromContextError(ctx.Err()).Err()
		}
		log.Printf("Service: Error reading upload stream: %v", err)
		return bytesReceived, status.Errorf(codes.Internal, "error processing upload stream: %v", err)
	}
	log.Printf("Service: Upload finished. Bytes received: %d", bytesReceived)
	return bytesReceived, nil
}
