package streamrpc

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

const (
	RequestIDMSB   = uint32(0x80000000) // Most significant bit mask
	RequestIDMask  = uint32(0x7fffffff) // Mask for actual request ID value
	MaxMessageSize = 10 * 1024 * 1024   // 10MB
)

type Stream interface {
	io.Reader
	io.Writer
}

type StreamCloseHandler func(error)

type RpcPeer struct {
	stream        Stream
	services      map[string]interface{}
	nextRequestID uint32
	mu            sync.Mutex
	writeMu       sync.Mutex
	pendingCalls  map[uint32]chan []byte
	onStreamClose StreamCloseHandler
}

func NewRpcPeer(stream Stream) *RpcPeer {
	peer := &RpcPeer{
		stream:        stream,
		services:      make(map[string]interface{}),
		nextRequestID: 1,
		pendingCalls:  make(map[uint32]chan []byte),
	}
	go peer.handleMessages()
	return peer
}

func (p *RpcPeer) RegisterService(name string, service interface{}) {
	p.services[name] = service
}

func (p *RpcPeer) getNextRequestID() uint32 {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Never use ID 0
	if p.nextRequestID == 0 {
		p.nextRequestID = 1
	}

	id := p.nextRequestID

	// Increment and mask, ensuring we don't exceed RequestIDMask
	p.nextRequestID = (p.nextRequestID + 1) & RequestIDMask

	// Double-check we haven't wrapped to 0
	if p.nextRequestID == 0 {
		p.nextRequestID = 1
	}

	// Ensure the ID we're returning doesn't have MSB set
	if (id & RequestIDMSB) != 0 {
		// This shouldn't happen with proper masking, but let's be safe
		id &= RequestIDMask
	}

	return id
}

func (p *RpcPeer) Call(methodName string, request proto.Message, response proto.Message) error {
	requestBytes, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	requestID := p.getNextRequestID()
	responseChan := make(chan []byte, 1)

	p.mu.Lock()
	p.pendingCalls[requestID] = responseChan
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.pendingCalls, requestID)
		p.mu.Unlock()
	}()

	if err := p.writeRequest(requestID, methodName, requestBytes); err != nil {
		return err
	}

	select {
	case responseBytes := <-responseChan:
		// Check if it's an error response by looking at the second MSB
		if (requestID & uint32(0x40000000)) != 0 {
			rpcErr, err := p.readErrorResponse(responseBytes)
			if err != nil {
				return fmt.Errorf("failed to read error response: %v", err)
			}
			return rpcErr
		}
		return proto.Unmarshal(responseBytes, response)
	case <-time.After(30 * time.Second):
		return fmt.Errorf("RPC call timeout")
	}
}

func (p *RpcPeer) handleMessages() {
	for {
		_, requestID, methodName, payload, err := p.readMessage()
		if err != nil {
			p.mu.Lock()
			handler := p.onStreamClose
			p.mu.Unlock()

			if handler != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
					websocket.CloseNoStatusReceived) {
					// normal closure
					handler(nil)
				} else {
					handler(err) // Error case
				}
			}
			return
		}

		isResponse := (requestID & RequestIDMSB) != 0
		if isResponse {
			originalRequestID := requestID & RequestIDMask
			p.mu.Lock()
			responseChan, ok := p.pendingCalls[originalRequestID]
			p.mu.Unlock()

			if ok {
				responseChan <- payload
			}
		} else {
			go p.handleRequest(requestID, methodName, payload)
		}
	}
}

func (p *RpcPeer) readMessage() (uint32, uint32, string, []byte, error) {
	var length uint32
	if err := binary.Read(p.stream, binary.BigEndian, &length); err != nil {
		return 0, 0, "", nil, err
	}

	if length < 4 || length > MaxMessageSize {
		return 0, 0, "", nil, fmt.Errorf("invalid message length: %d bytes", length)
	}

	var requestID uint32
	if err := binary.Read(p.stream, binary.BigEndian, &requestID); err != nil {
		return 0, 0, "", nil, err
	}

	var methodName string
	var payload []byte

	if (requestID & RequestIDMSB) == 0 {
		// Request message
		var methodNameLen uint8
		if err := binary.Read(p.stream, binary.BigEndian, &methodNameLen); err != nil {
			return 0, 0, "", nil, err
		}

		methodNameBytes := make([]byte, methodNameLen)
		if _, err := io.ReadFull(p.stream, methodNameBytes); err != nil {
			return 0, 0, "", nil, err
		}
		methodName = string(methodNameBytes)

		payloadLen := length - uint32(methodNameLen) - 5
		payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(p.stream, payload); err != nil {
			return 0, 0, "", nil, err
		}
	} else {
		// Response message
		payloadLen := length - 4
		payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(p.stream, payload); err != nil {
			return 0, 0, "", nil, err
		}
	}

	return length, requestID, methodName, payload, nil
}

func (p *RpcPeer) writeRequest(requestID uint32, methodName string, payload []byte) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	methodNameBytes := []byte(methodName)
	totalLength := uint32(len(payload) + len(methodNameBytes) + 5)

	if err := binary.Write(p.stream, binary.BigEndian, totalLength); err != nil {
		return err
	}

	if err := binary.Write(p.stream, binary.BigEndian, requestID); err != nil {
		return err
	}

	if err := binary.Write(p.stream, binary.BigEndian, uint8(len(methodNameBytes))); err != nil {
		return err
	}

	if _, err := p.stream.Write(methodNameBytes); err != nil {
		return err
	}

	_, err := p.stream.Write(payload)
	return err
}

func (p *RpcPeer) writeResponse(requestID uint32, payload []byte) error {
	// Ensure the incoming requestID doesn't have MSB set
	requestID &= RequestIDMask

	// Set MSB for response
	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	responseID := requestID | RequestIDMSB

	// Validate the resulting ID
	if (responseID & RequestIDMask) != requestID {
		return fmt.Errorf("invalid response ID generated: original=%d, response=%d", requestID, responseID)
	}

	totalLength := uint32(len(payload) + 4)

	if err := binary.Write(p.stream, binary.BigEndian, totalLength); err != nil {
		return err
	}

	if err := binary.Write(p.stream, binary.BigEndian, responseID); err != nil {
		return err
	}

	_, err := p.stream.Write(payload)
	return err
}

func (p *RpcPeer) handleRequest(requestID uint32, methodName string, payload []byte) {
	parts := strings.Split(methodName, ".")
	if len(parts) != 2 {
		p.writeErrorResponse(requestID, ErrorCodeInvalidRequest, "invalid method name format")
		return
	}

	serviceName, methodName := parts[0], parts[1]
	service, ok := p.services[serviceName]
	if !ok {
		p.writeErrorResponse(requestID, ErrorCodeMethodNotFound, fmt.Sprintf("service %s not found", serviceName))
		return
	}

	serviceValue := reflect.ValueOf(service)
	method := serviceValue.MethodByName(methodName)
	if !method.IsValid() {
		fmt.Printf("Method not found: %s.%s\n", serviceName, methodName)
		p.writeErrorResponse(requestID, ErrorCodeMethodNotFound, fmt.Sprintf("method %s not found", methodName))
		return
	}

	// Create the appropriate request message type
	methodType := method.Type()
	if methodType.NumIn() != 2 { // Context and request message
		p.writeErrorResponse(requestID, ErrorCodeInvalidRequest, "invalid method signature")
		return
	}

	// Create and unmarshal the request message
	requestMsgType := methodType.In(1).Elem()
	requestMsg := reflect.New(requestMsgType).Interface().(proto.Message)
	if err := proto.Unmarshal(payload, requestMsg); err != nil {
		fmt.Printf("Unmarshal error: %v\n", err)
		p.writeErrorResponse(requestID, ErrorCodeInternalError, fmt.Sprintf("failed to unmarshal request: %v", err))
		return
	}

	// Call the method with context and request message
	results := method.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(requestMsg),
	})

	if len(results) != 1 {
		p.writeErrorResponse(requestID, ErrorCodeInternalError, "invalid method return values, should be only one return value")
		return
	}

	// Marshal the response
	response := results[0].Interface().(proto.Message)
	responseBytes, err := proto.Marshal(response)
	if err != nil {
		p.writeErrorResponse(requestID, ErrorCodeInternalError, fmt.Sprintf("failed to marshal response: %v", err))
		return
	}

	p.writeResponse(requestID, responseBytes)
}

func (p *RpcPeer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Clear pending calls
	for _, ch := range p.pendingCalls {
		close(ch)
	}
	p.pendingCalls = make(map[uint32]chan []byte)

	// Close stream if it implements io.Closer
	if closer, ok := p.stream.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (p *RpcPeer) OnStreamClose(handler StreamCloseHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onStreamClose = handler
}

var ErrNotImplemented = errors.New("method not implemented")

// ErrorCode represents different types of framework-level errors
type ErrorCode uint32

const (
	ErrorCodeUnknown ErrorCode = iota
	ErrorCodeMethodNotFound
	ErrorCodeInvalidRequest
	ErrorCodeMalformedRequest
	ErrorCodeInvalidMessageFormat
	ErrorCodeInternalError
)

// RPCError represents a framework-level RPC error
type RPCError struct {
	Code    ErrorCode
	Message string
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

func (p *RpcPeer) writeErrorResponse(requestID uint32, code ErrorCode, message string) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	// Convert message to bytes
	messageBytes := []byte(message)

	// Total length = 4 (errorCode) + len(message)
	totalLength := uint32(4 + len(messageBytes))

	// Set MSB for response and second MSB for error
	responseID := (requestID & RequestIDMask) | RequestIDMSB | uint32(0x40000000)

	// Write total length
	if err := binary.Write(p.stream, binary.BigEndian, totalLength); err != nil {
		return err
	}

	// Write response ID
	if err := binary.Write(p.stream, binary.BigEndian, responseID); err != nil {
		return err
	}

	// Write error code
	if err := binary.Write(p.stream, binary.BigEndian, uint32(code)); err != nil {
		return err
	}

	// Write error message
	_, err := p.stream.Write(messageBytes)
	return err
}

func (p *RpcPeer) readErrorResponse(payload []byte) (*RPCError, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("error payload too short")
	}

	// Read error code
	errorCode := ErrorCode(binary.BigEndian.Uint32(payload[:4]))

	// Read error message
	message := string(payload[4:])

	return &RPCError{
		Code:    errorCode,
		Message: message,
	}, nil
}
