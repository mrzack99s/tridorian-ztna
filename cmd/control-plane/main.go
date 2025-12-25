package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"tridorian-ztna/internal/api/controlplane"

	"github.com/quic-go/quic-go"
)

func main() {
	fmt.Println("Starting Tridorian ZTNA Control Plane (QUIC)...")

	handler := controlplane.NewHandler()

	// Generate a simple TLS config (For production, use proper mTLS with CA)
	tlsConf := generateTLSConfig()

	listener, err := quic.ListenAddr(":9999", tlsConf, nil)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	fmt.Println("Control Plane listening on UDP :9999 (QUIC)")

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn, handler)
	}
}

func handleConnection(conn *quic.Conn, h *controlplane.Handler) {
	fmt.Printf("New Agent connection from: %s\n", conn.RemoteAddr())

	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Printf("Connection closed by %s: %v", conn.RemoteAddr(), err)
			return
		}

		go h.HandleAgentStream(conn, stream)
	}
}

// generateTLSConfig is for development. In production, use certificates.
func generateTLSConfig() *tls.Config {
	// Dummy TLS config for now
	return &tls.Config{
		InsecureSkipVerify: true, // DO NOT USE IN PRODUCTION
		NextProtos:         []string{"tridorian-ztna-v1"},
	}
}
