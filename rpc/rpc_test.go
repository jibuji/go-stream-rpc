package rpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jibuji/go-stream-rpc/session"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
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

// Mock protobuf messages for testing
type MockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Value         int32 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

type MockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
	Result        int32 `protobuf:"varint,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (m *MockRequest) Reset()         { *m = MockRequest{} }
func (m *MockRequest) String() string { return fmt.Sprintf("MockRequest{Value: %d}", m.Value) }
func (*MockRequest) ProtoMessage()    {}
func (m *MockRequest) ProtoReflect() protoreflect.Message {
	return nil
}

func (m *MockResponse) Reset()         { *m = MockResponse{} }
func (m *MockResponse) String() string { return fmt.Sprintf("MockResponse{Result: %d}", m.Result) }
func (*MockResponse) ProtoMessage()    {}
func (m *MockResponse) ProtoReflect() protoreflect.Message {
	return nil
}

// Mock service for testing
type MockService struct{}

func (s *MockService) Add(ctx context.Context, req *MockRequest) *MockResponse {
	return &MockResponse{Result: req.Value + 1}
}

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

		sess := session.From(peer.ctx)
		if sess == nil {
			t.Error("Default session not created")
		}
	})

	t.Run("with custom session", func(t *testing.T) {
		stream := NewMockStream()
		mockSession := session.NewMemSession()
		mockSession.Set("preexisting", "data")

		peer := NewRpcPeer(stream, WithSession(mockSession))

		sess := session.From(peer.ctx)
		if sess != mockSession {
			t.Error("Custom session not properly set")
		}

		if val := sess.Get("preexisting"); val != "data" {
			t.Errorf("Custom session data not preserved, got %v, want %v", val, "data")
		}
	})
}

func TestRpcPeer_RegisterService(t *testing.T) {
	peer := NewRpcPeer(NewMockStream())
	service := &MockService{}

	peer.RegisterService("Calculator", service)

	if _, exists := peer.services["Calculator"]; !exists {
		t.Error("Service not properly registered")
	}
}

func TestRpcPeer_GetNextRequestID(t *testing.T) {
	peer := NewRpcPeer(NewMockStream())

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

func TestRpcPeer_Wait(t *testing.T) {
	t.Run("normal closure", func(t *testing.T) {
		stream := NewMockStream()
		peer := NewRpcPeer(stream)

		go func() {
			time.Sleep(100 * time.Millisecond)
			stream.readErr = &websocket.CloseError{
				Code: websocket.CloseNormalClosure,
			}
		}()

		err := peer.Wait()
		if err != nil {
			t.Errorf("Expected nil error for normal closure, got %v", err)
		}
	})

	t.Run("error closure", func(t *testing.T) {
		stream := NewMockStream()
		peer := NewRpcPeer(stream)

		testErr := errors.New("test error")
		go func() {
			time.Sleep(100 * time.Millisecond)
			stream.readErr = testErr
		}()

		err := peer.Wait()
		if err == nil || !strings.Contains(err.Error(), testErr.Error()) {
			t.Errorf("Expected error containing %v, got %v", testErr, err)
		}
	})
}

func TestRpcPeer_ErrorHandling(t *testing.T) {
	t.Run("method not found", func(t *testing.T) {
		stream := NewMockStream()
		peer := NewRpcPeer(stream)

		req := &MockRequest{Value: 42}
		resp := &MockResponse{}

		err := peer.Call("NonExistentService.Method", req, resp)
		if err == nil {
			t.Error("Expected error for non-existent service")
		}
	})

	t.Run("invalid method name", func(t *testing.T) {
		stream := NewMockStream()
		peer := NewRpcPeer(stream)

		req := &MockRequest{Value: 42}
		resp := &MockResponse{}

		err := peer.Call("InvalidMethodName", req, resp)
		if err == nil {
			t.Error("Expected error for invalid method name")
		}
	})
}
