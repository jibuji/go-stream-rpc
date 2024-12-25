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

// MockSession implements Session interface for testing
type MockSession struct {
	data map[interface{}]interface{}
	mu   sync.RWMutex
}

func NewMockSession() *MockSession {
	return &MockSession{
		data: make(map[interface{}]interface{}),
	}
}

func (s *MockSession) Get(key interface{}) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *MockSession) Set(key, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Tests
func TestRpcPeer_NewRpcPeer(t *testing.T) {
	t.Run("with default session", func(t *testing.T) {
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

		// Test default session
		session := From(peer.ctx)
		if session == nil {
			t.Error("Default session not created")
		}

		// Test session functionality
		session.Set("test", "value")
		if val := session.Get("test"); val != "value" {
			t.Errorf("Session Get/Set not working, got %v, want %v", val, "value")
		}
	})

	t.Run("with custom session", func(t *testing.T) {
		stream := NewMockStream()
		mockSession := NewMockSession()
		mockSession.Set("preexisting", "data")

		peer := NewRpcPeer(stream, WithSession(mockSession))

		// Verify the custom session was used
		session := From(peer.ctx)
		if session != mockSession {
			t.Error("Custom session not properly set")
		}

		// Test preexisting data
		if val := session.Get("preexisting"); val != "data" {
			t.Errorf("Custom session data not preserved, got %v, want %v", val, "data")
		}
	})
}

func TestCreateDefaultSessionContext(t *testing.T) {
	ctx := CreateDefaultSessionContext()

	session := From(ctx)
	if session == nil {
		t.Fatal("CreateDefaultSessionContext did not create a session")
	}

	// Test session is a MemSession
	if _, ok := session.(*MemSession); !ok {
		t.Error("Created session is not a MemSession")
	}

	// Test session functionality
	session.Set("key", "value")
	if val := session.Get("key"); val != "value" {
		t.Errorf("Session Get/Set not working, got %v, want %v", val, "value")
	}
}

func TestMemSession(t *testing.T) {
	session := NewMemSession()

	// Test basic Get/Set
	session.Set("string", "value")
	if val := session.Get("string"); val != "value" {
		t.Errorf("String value mismatch, got %v, want %v", val, "value")
	}

	// Test with different types
	session.Set("int", 42)
	if val := session.Get("int"); val != 42 {
		t.Errorf("Int value mismatch, got %v, want %v", val, 42)
	}

	// Test non-existent key
	if val := session.Get("nonexistent"); val != nil {
		t.Errorf("Non-existent key should return nil, got %v", val)
	}

	// Test overwriting value
	session.Set("key", "value1")
	session.Set("key", "value2")
	if val := session.Get("key"); val != "value2" {
		t.Errorf("Overwritten value mismatch, got %v, want %v", val, "value2")
	}
}

func TestFrom(t *testing.T) {
	t.Run("with valid session", func(t *testing.T) {
		session := NewMemSession()
		ctx := context.WithValue(context.Background(), SessionContextKey, session)

		retrieved := From(ctx)
		if retrieved != session {
			t.Error("From did not return the correct session")
		}
	})

	t.Run("with no session", func(t *testing.T) {
		ctx := context.Background()
		if session := From(ctx); session != nil {
			t.Error("From should return nil when no session is present")
		}
	})

	t.Run("with invalid session type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), SessionContextKey, "not a session")
		if session := From(ctx); session != nil {
			t.Error("From should return nil when context value is not a Session")
		}
	})
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
