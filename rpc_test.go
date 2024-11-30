package streamrpc

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

// MockStream implements Stream interface for testing
type MockStream struct {
	readData  []byte
	writeData []byte
	readErr   error
	writeErr  error
	closed    bool
	mu        sync.Mutex
}

func NewMockStream() *MockStream {
	return &MockStream{
		readData:  make([]byte, 0),
		writeData: make([]byte, 0),
	}
}

func (m *MockStream) Read(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.readErr != nil {
		return 0, m.readErr
	}

	if len(m.readData) == 0 {
		return 0, io.EOF
	}

	n = copy(p, m.readData)
	m.readData = m.readData[n:]
	return n, nil
}

func (m *MockStream) Write(p []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.writeErr != nil {
		return 0, m.writeErr
	}

	m.writeData = append(m.writeData, p...)
	return len(p), nil
}

func (m *MockStream) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// Mock service for testing
type MockService struct {
	t *testing.T
}

type MockRequest struct {
	Value int32
}

type MockResponse struct {
	Result int32
}

func (m *MockService) Add(ctx context.Context, req *MockRequest) (*MockResponse, error) {
	return &MockResponse{Result: req.Value + 1}, nil
}

// Tests
func TestRpcPeer_NewRpcPeer(t *testing.T) {
	stream := NewMockStream()
	peer := NewRpcPeer(stream)

	if peer == nil {
		t.Fatal("NewRpcPeer returned nil")
	}

	if peer.stream != stream {
		t.Error("Stream not properly set")
	}

	if peer.services == nil {
		t.Error("Services map not initialized")
	}

	if peer.pendingCalls == nil {
		t.Error("PendingCalls map not initialized")
	}
}

func TestRpcPeer_RegisterService(t *testing.T) {
	peer := NewRpcPeer(NewMockStream())
	service := &MockService{t: t}

	peer.RegisterService("calculator", service)

	if _, exists := peer.services["calculator"]; !exists {
		t.Error("Service not properly registered")
	}
}

func TestRpcPeer_GetNextRequestID(t *testing.T) {
	peer := NewRpcPeer(NewMockStream())

	// Test sequential IDs
	id1 := peer.getNextRequestID()
	id2 := peer.getNextRequestID()

	if id1 == 0 {
		t.Error("First ID should not be 0")
	}

	if id2 <= id1 {
		t.Error("Second ID should be greater than first ID")
	}

	if (id1 & RequestIDMSB) != 0 {
		t.Error("MSB should not be set in request ID")
	}
}

func TestRpcPeer_Close(t *testing.T) {
	stream := NewMockStream()
	peer := NewRpcPeer(stream)

	// Create a pending call
	peer.pendingCalls[1] = make(chan []byte, 1)

	err := peer.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	if !stream.closed {
		t.Error("Stream not closed")
	}

	if len(peer.pendingCalls) != 0 {
		t.Error("Pending calls not cleared")
	}
}

func TestRpcPeer_OnStreamClose(t *testing.T) {
	stream := NewMockStream()
	peer := NewRpcPeer(stream)

	called := false
	handler := func(err error) {
		called = true
	}

	peer.OnStreamClose(handler)

	// Simulate stream close by setting read error
	stream.readErr = errors.New("connection closed")

	// Wait for handler to be called
	time.Sleep(100 * time.Millisecond)

	if !called {
		t.Error("Stream close handler not called")
	}
}

// Add more tests for error handling, message formatting, etc.
