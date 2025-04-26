package service

import (
	"context"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type defaultIPDetector struct {
	// Add dependencies if needed (e.g., GeoIP lookup client)
}

func NewDefaultIPDetector() IPDetector {
	return &defaultIPDetector{}
}

func (d *defaultIPDetector) DetectClientIP(ctx context.Context) (ip string, city string, country string, err error) {
	log.Println("Service: Detecting client IP.")
	ip = "Unknown"
	city = "Unknown" // Placeholder
	country = "XX"   // Placeholder

	// 1. Check gRPC Peer
	p, ok := peer.FromContext(ctx)
	if ok {
		if tcpAddr, ok := p.Addr.(*net.TCPAddr); ok {
			ip = tcpAddr.IP.String()
		} else {
			ip = p.Addr.String() // Fallback
		}
	}

	// 2. Check Headers (if behind proxy)
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
			// Use the first IP, potentially clean it
			ip = strings.TrimSpace(strings.Split(ips[0], ",")[0])
		} else if ips := md.Get("x-real-ip"); len(ips) > 0 {
			ip = strings.TrimSpace(ips[0])
		}
	}

	// 3. GeoIP Lookup (Placeholder)
	// if ip != "Unknown" {
	//    geoResult, geoErr := d.geoIPClient.Lookup(ctx, ip)
	//    if geoErr == nil {
	// 		 city = geoResult.City
	//       country = geoResult.CountryCode
	//    } else {
	//       log.Printf("GeoIP lookup failed for %s: %v", ip, geoErr)
	//    }
	// }

	log.Printf("Service: Detected client IP: %s", ip)
	return ip, city, country, nil
}
