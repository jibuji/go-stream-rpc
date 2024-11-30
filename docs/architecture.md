# RPC Framework Architecture

## Overview
This is a lightweight RPC (Remote Procedure Call) framework that enables communication between services using a simple wire protocol. The framework supports both unary and streaming RPC calls.

## Core Components

### 1. Transport Layer
- Based on TCP for reliable communication
- Handles raw byte streams between client and server
- Implements connection management and message framing

### 2. Protocol Format
The framework uses a simple binary protocol for message exchange:

#### Request Format
```
[total length (4 bytes)][request ID (4 bytes)][method name length (4 bytes)][method name][payload]
```

#### Response Format
```
[total length (4 bytes)][response ID (4 bytes)][payload]
```

#### Error Response Format
```
[total length (4 bytes)][response ID (4 bytes)][error code (4 bytes)][error message]
```

### 3. Code Generator
- Generates client and server stubs from Protocol Buffer definitions
- Handles serialization/deserialization of messages
- Creates type-safe RPC method handlers

## Message Flow
1. Client initiates connection to server
2. Client sends RPC request with unique request ID
3. Server processes request and sends response
4. Client matches response to request using request ID

## Error Handling
- Framework-level errors (connection, protocol, etc.)
- Application-level errors (business logic)
- Error codes and messages follow standardized format