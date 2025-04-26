package utils

import (
	"crypto/rand"
	"io"
	"log"
)

// ConstantReader always reads the same byte slice. More efficient than crypto/rand for high throughput.
type ConstantReader struct {
	data []byte
	pos  int
}

func NewConstantReader(size int) *ConstantReader {
	data := make([]byte, size)
	// Fill with a simple pattern or pre-generate random data once
	if _, err := rand.Read(data); err != nil {
		log.Printf("Warning: Failed to pre-generate random data, using zeros: %v", err)
	}
	return &ConstantReader{data: data}
}

func (r *ConstantReader) Read(p []byte) (n int, err error) {
	copied := 0
	for copied < len(p) {
		remainingInChunk := len(r.data) - r.pos
		if remainingInChunk == 0 {
			r.pos = 0 // Reset position to loop the data
			remainingInChunk = len(r.data)
		}
		toCopy := min(len(p)-copied, remainingInChunk)
		copy(p[copied:], r.data[r.pos:r.pos+toCopy])
		r.pos += toCopy
		copied += toCopy
	}
	return copied, nil
}

// CryptoRandReader reads directly from crypto/rand. Less efficient but truly random.
type CryptoRandReader struct{}

func (r *CryptoRandReader) Read(p []byte) (n int, err error) {
	return rand.Read(p)
}

// StreamData writes data from a reader to the writer in chunks.
// Returns the total bytes written or an error.
func StreamData(writer io.Writer, reader io.Reader, chunkSize int) (int64, error) {
	buffer := make([]byte, chunkSize)
	var totalBytesWritten int64 = 0

	for {
		// Use io.CopyBuffer for potentially better performance
		bytesWritten, err := io.CopyBuffer(writer, reader, buffer)
		totalBytesWritten += bytesWritten
		if err != nil {
			if err == io.EOF {
				return totalBytesWritten, nil // Expected EOF from reader maybe? Usually not for streams.
			}
			// Handle specific writer errors (like client disconnect) if possible
			return totalBytesWritten, err
		}
		// For continuous streams, io.Copy might block if reader blocks.
		// If using a custom reader that doesn't EOF, need another stop condition.
	}
}
