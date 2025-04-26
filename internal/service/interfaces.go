package service

import (
	"context"
	"io"

	pb "barry-server-go/proto/speedtest" // Alias proto package
)

// ServerProvider defines the interface for getting server information.
type ServerProvider interface {
	GetServers(ctx context.Context, limit int32) ([]*pb.ServerInfo, error)
}

// DownloadStreamer defines the interface for handling download data streaming.
type DownloadStreamer interface {
	// StreamDownloadData generates and writes download data to the writer.
	// It should respect the context cancellation.
	StreamDownloadData(ctx context.Context, writer io.Writer, chunkSize int) error
}

// UploadHandler defines the interface for processing uploaded data streams.
type UploadHandler interface {
	// HandleUploadStream reads from the reader (client stream) and returns total bytes read.
	// It should respect the context cancellation.
	HandleUploadStream(ctx context.Context, reader io.Reader) (bytesReceived int64, err error)
}

// IPDetector defines the interface for determining the client's IP address.
type IPDetector interface {
	DetectClientIP(ctx context.Context) (ip string, city string, country string, err error)
}
