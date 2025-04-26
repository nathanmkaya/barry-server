package service

import (
	"context"
	"log"

	"barry-server-go/internal/config"
	pb "barry-server-go/proto/speedtest"
)

// simpleServerProvider implements ServerProvider by returning the server's own info.
type simpleServerProvider struct {
	cfg *config.Config
	// In a real app, this might have DB/Cache dependencies
}

// NewSimpleServerProvider creates a new simpleServerProvider.
func NewSimpleServerProvider(cfg *config.Config) ServerProvider {
	return &simpleServerProvider{cfg: cfg}
}

func (p *simpleServerProvider) GetServers(ctx context.Context, limit int32) ([]*pb.ServerInfo, error) {
	log.Printf("Service: GetServers called. Limit: %d", limit)
	// Bare minimum: return self. Real implementation would query DB/Cache.
	serverInfo := &pb.ServerInfo{
		Id:      p.cfg.ServerID,
		Url:     p.cfg.PublicURL,
		Region:  "dev-region", // Placeholder - Load from config/DB
		City:    "dev-city",   // Placeholder - Load from config/DB
		Country: "XX",         // Placeholder - Load from config/DB
	}
	servers := []*pb.ServerInfo{serverInfo}

	if limit > 0 && len(servers) > int(limit) {
		servers = servers[:limit]
	}
	return servers, nil
}
