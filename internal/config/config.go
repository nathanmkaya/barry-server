package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	ListenAddress  string // e.g., ":8080" or "0.0.0.0:8080"
	ServerID       string // Unique ID for this server instance
	PublicURL      string // Publicly accessible gRPC URL (e.g., "grpc.example.com:443")
	ChunkSizeBytes int    // Default chunk size for downloads
	// Add other configs: Region, City, Country, etc.
}

func Load() *Config {
	chunkSizeStr := getEnv("CHUNK_SIZE_BYTES", "65536") // 64KB default
	chunkSize, err := strconv.Atoi(chunkSizeStr)
	if err != nil {
		log.Fatalf("Invalid CHUNK_SIZE_BYTES: %v", err)
	}

	return &Config{
		ListenAddress:  getEnv("LISTEN_ADDRESS", ":8080"),
		ServerID:       getEnv("SERVER_ID", "default-go-server"),
		PublicURL:      getEnv("PUBLIC_URL", "localhost:8080"), // Needs to be accurate!
		ChunkSizeBytes: chunkSize,
		// Load Region, City, Country etc. from env or defaults
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
