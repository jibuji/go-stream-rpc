# Wire Protocol Specification

## Overview
The framework uses a binary protocol for efficient data transmission. All integers are encoded in big-endian format.

## Message Types

### 1. Request Message
```
[total length (4 bytes)][request ID (4 bytes)][method name length (4 bytes)][method name][payload]
```
- Total length: Length of entire message including header
- Request ID: Unique identifier for request-response matching
- Method name length: Length of the method name string
- Method name: UTF-8 encoded service method name
- Payload: Protobuf-encoded request message

### 2. Response Message
```
[total length (4 bytes)][response ID (4 bytes)][payload]
```
- Total length: Length of entire message
- Response ID: Matches the request ID
- Payload: Protobuf-encoded response message

### 3. Error Response
```
[total length (4 bytes)][response ID (4 bytes)][error code (4 bytes)][error message]
```
- Error code: Predefined error code
- Error message: UTF-8 encoded error description

## Error Codes
```go
const (
    ErrorCodeUnknown            uint32 = 0
    ErrorCodeMethodNotFound     uint32 = 1
    ErrorCodeInvalidRequest     uint32 = 2
    ErrorCodeMalformedRequest   uint32 = 3
    ErrorCodeInvalidMessageFormat uint32 = 4
    ErrorCodeInternalError      uint32 = 5
)
```

## Streaming
For streaming RPCs, multiple messages are sent using the same format with the same request ID.