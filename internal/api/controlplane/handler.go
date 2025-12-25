package controlplane

import (
	"fmt"
	"io"

	"github.com/quic-go/quic-go"
)

type Handler struct {
	// Add dependencies like DB or Service layer
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleAgentStream(conn *quic.Conn, stream *quic.Stream) {
	defer stream.Close()

	buf := make([]byte, 1024)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from stream: %v\n", err)
			}
			return
		}

		data := string(buf[:n])
		fmt.Printf("Received from Agent [%s]: %s\n", conn.RemoteAddr(), data)

		// Simple Echo for now
		stream.Write([]byte("ACK: " + data))
	}
}
