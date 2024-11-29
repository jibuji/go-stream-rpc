# Go RPC Framework

A flexible RPC framework that works over any duplex stream. It provides a simple way to build RPC services that can work over any bidirectional stream like WebSocket, TCP, or libp2p.

## Features

- Protocol Buffer support
- Works with any duplex stream (WebSocket, TCP, libp2p, etc.)
- Built-in stream implementations:
  - libp2p
  - WebSocket
- Simple API with minimal dependencies
- Context support
- Concurrent request handling
- Automatic message framing
- Request timeout handling

## Installation 

```bash
go get github.com/yourusername/rpc
```

## Quick Start

1. Define your service using Protocol Buffers:

```protobuf
syntax = "proto3";
package calculator;

service Calculator {
  rpc Add(AddRequest) returns (AddResponse);
}

message AddRequest {
  int32 a = 1;
  int32 b = 2;
}

message AddResponse {
  int32 result = 1;
}
```

2. Implement your service:

```go
type CalculatorService struct{}

func (s *CalculatorService) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
    return &pb.AddResponse{Result: req.A + req.B}, nil
}
```

3. Create an RPC server:

```go
peer := rpc.NewRpcPeer(stream)
peer.RegisterService("Calculator", &CalculatorService{})
```

4. Call the service from client:

```go
peer := rpc.NewRpcPeer(clientStream)
request := &pb.AddRequest{A: 5, B: 3}
response := &pb.AddResponse{}
err := peer.Call("Calculator.Add", request, response)
```

## Available Stream Implementations

- **WebSocket**: `stream/websocket`
- **libp2p**: `stream/libp2p`

## Examples

See the [examples](examples/) directory for complete examples:
- [Calculator Service](examples/calculator/): Basic calculator service demonstrating RPC usage
- More examples coming soon...

## Project Structure

```
.
├── stream/           # Stream interface and implementations
│   ├── libp2p/      # libp2p stream implementation
│   └── websocket/   # WebSocket stream implementation
├── examples/         # Example applications
├── rpc.go           # Core RPC implementation
└── doc.go           # Package documentation
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License (or your chosen license)