package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketStream struct {
	Conn         *websocket.Conn
	readBuffer   []byte // Buffer for remaining data
	readDeadline time.Duration
}

func NewWebSocketStream(conn *websocket.Conn) *WebSocketStream {
	return &WebSocketStream{
		Conn:         conn,
		readDeadline: 30 * time.Second, // Default timeout
	}
}

func (s *WebSocketStream) Read(p []byte) (n int, err error) {
	// If we have remaining data in the buffer, use it first
	if len(s.readBuffer) > 0 {
		n = copy(p, s.readBuffer)
		s.readBuffer = s.readBuffer[n:]
		return n, nil
	}

	_, message, err := s.Conn.ReadMessage()
	if err != nil {
		return 0, err
	}

	// Copy what we can into p
	n = copy(p, message)

	// If we couldn't copy everything, save the rest
	if n < len(message) {
		s.readBuffer = make([]byte, len(message)-n)
		copy(s.readBuffer, message[n:])
	}

	return n, nil
}

func (s *WebSocketStream) Write(p []byte) (n int, err error) {
	err = s.Conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *WebSocketStream) Close() error {
	return s.Conn.Close()
}

// Optional: Add method to configure read deadline
func (s *WebSocketStream) SetReadDeadline(d time.Duration) {
	s.readDeadline = d
}
