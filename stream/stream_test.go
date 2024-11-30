package stream

import (
	"net"
	"testing"
	"time"

	"github.com/jibuji/go-stream-rpc/stream/tcp"
)

func TestTCPStream_ReadWrite(t *testing.T) {
	// Create a TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Connect client and server
	done := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Failed to accept connection: %v", err)
			return
		}
		defer conn.Close()

		stream := tcp.NewTCPStream(conn)

		// Read data
		buf := make([]byte, 1024)
		n, err := stream.Read(buf)
		if err != nil {
			t.Errorf("Failed to read from stream: %v", err)
			return
		}

		// Echo back
		_, err = stream.Write(buf[:n])
		if err != nil {
			t.Errorf("Failed to write to stream: %v", err)
			return
		}

		close(done)
	}()

	// Client connection
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	stream := tcp.NewTCPStream(conn)

	// Test data
	testData := []byte("Hello, World!")

	// Write test data
	_, err = stream.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write to stream: %v", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read from stream: %v", err)
	}

	if string(buf[:n]) != string(testData) {
		t.Errorf("Data mismatch. Got %s, want %s", string(buf[:n]), string(testData))
	}

	// Wait for server to complete
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Test timed out")
	}
}

func TestWebSocketStream_ReadWrite(t *testing.T) {
	// Similar to TCP test but using WebSocket
	// You'll need to set up a WebSocket server and client
	// This is a basic structure - implement according to your WebSocket implementation
}
