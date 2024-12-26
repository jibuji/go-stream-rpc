# Lightweight RPC Framework

[![Go Reference](https://pkg.go.dev/badge/github.com/jibuji/go-stream-rpc.svg)](https://pkg.go.dev/github.com/jibuji/go-stream-rpc)

A simple, efficient RPC framework for Go applications with Protocol Buffers support and libp2p integration.

## Features
- Simple wire protocol for efficient communication
- Protocol Buffers integration
- Support for unary and streaming RPCs
- libp2p transport layer support
- Automatic code generation
- Easy-to-use API

## Limitations
- Each proto file must contain only one service definition
- Services must be defined in separate proto files

## Installation

### Install the library and protoc plugin
```bash
# Install the library
go get github.com/jibuji/go-stream-rpc@latest

# Install the protoc plugin
go install github.com/jibuji/go-stream-rpc/cmd/protoc-gen-stream-rpc@latest
```

### Publishing New Versions
To publish a new version:

1. Tag your release:
```bash
git tag v1.0.0  # Use appropriate version number
git push origin v1.0.0
```

2. Users can then install specific versions:
```bash
go get github.com/jibuji/go-stream-rpc@v1.0.0
go install github.com/jibuji/go-stream-rpc/cmd/protoc-gen-stream-rpc@v1.0.0
```

### Requirements
- Go 1.20 or later
- Protocol Buffers compiler (protoc)
- Go Protocol Buffers plugin: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`

## Quick Start

### 1. Define your service in Protocol Buffers
Create a file `calculator.proto`:
```protobuf
syntax = "proto3";

package calculator;
option go_package = "example/calculator/proto";

service Calculator {
  rpc Add(AddRequest) returns (AddResponse);
  rpc Multiply(MultiplyRequest) returns (MultiplyResponse);
}

message AddRequest {
  int32 a = 1;
  int32 b = 2;
}

message AddResponse {
  int32 result = 1;
}

message MultiplyRequest {
  int32 a = 1;
  int32 b = 2;
}

message MultiplyResponse {
  int32 result = 1;
}
```

### 2. Generate code
```bash
# Generate both protobuf and stream-rpc code
protoc --go_out=. --go_opt=paths=source_relative \
       --stream-rpc_out=. --stream-rpc_opt=paths=source_relative \
       calculator.proto
```

This will generate several files for each service in your proto file:
- `calculator.pb.go`: Protocol Buffers generated code
- `calculator_servicename_client.pb.go`: RPC client code for each service
- `calculator_servicename_server.pb.go`: RPC server interfaces for each service
- `service/servicename.go`: Service implementation skeleton for each service

For example, if your proto file has two services named `Calculator` and `Stats`, you'll get:
- `calculator_calculator_client.pb.go` and `calculator_stats_client.pb.go`
- `calculator_calculator_server.pb.go` and `calculator_stats_server.pb.go`
- `service/calculator.go` and `service/stats.go`

### 3. Implement the service
The generator creates a service skeleton in `service/calculator.go`. Implement your service logic there:

```go
package service

import (
    "context"
    "example/calculator/proto"
)

type CalculatorService struct {
    // Add any service-wide fields here
}

func (s *CalculatorService) Add(ctx context.Context, req *proto.AddRequest) (*proto.AddResponse, error) {
    return &proto.AddResponse{
        Result: req.A + req.B,
    }, nil
}

func (s *CalculatorService) Multiply(ctx context.Context, req *proto.MultiplyRequest) (*proto.MultiplyResponse, error) {
    return &proto.MultiplyResponse{
        Result: req.A * req.B,
    }, nil
}
```

### 4. Run the server
```go
package main

import (
    "log"
    rpc "github.com/jibuji/go-stream-rpc/rpc"
    "github.com/jibuji/go-stream-rpc/stream/libp2p"
    "github.com/libp2p/go-libp2p"
)

func handleStream(s network.Stream) {
    libp2pStream := stream.NewLibP2PStream(s)
    peer := rpc.NewRpcPeer(libp2pStream)
    defer peer.Close()

    peer.RegisterService("Calculator", &calculator.CalculatorService{})
    
    peer.OnStreamClose(func(err error) {
        if err != nil {
            log.Printf("Stream error: %v\n", err)
        }
    })
}

func main() {
    h, err := libp2p.New(
        libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/9000"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    h.SetStreamHandler("/calculator/1.0.0", handleStream)
    select {}
}
```

### 5. Create a client
```go
package main

import (
    "context"
    "log"
    rpc "github.com/jibuji/go-stream-rpc/rpc"
    "github.com/jibuji/go-stream-rpc/stream/libp2p"
    "github.com/libp2p/go-libp2p"
)

func main() {
    h, _ := libp2p.New()
    defer h.Close()

    // Connect to peer
    maddr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/p2p/QmPeerID")
    info, _ := peer.AddrInfoFromP2pAddr(maddr)
    ctx := context.Background()
    h.Connect(ctx, *info)
    s, _ := h.NewStream(ctx, info.ID, "/calculator/1.0.0")

    // Create RPC client
    libp2pStream := stream.NewLibP2PStream(s)
    peer := rpc.NewRpcPeer(libp2pStream)
    calculatorClient := proto.NewCalculatorClient(peer)

    // Make RPC calls
    addResp, err := calculatorClient.Add(ctx, &proto.AddRequest{A: 5, B: 3})
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("5 + 3 = %d\n", addResp.Result)
}
```

## Documentation
- [Architecture Overview](docs/architecture.md)
- [Getting Started Guide](docs/getting_started.md)
- [Wire Protocol Specification](docs/wire_protocol.md)
- [Development Log](devlog.md)

## Examples
Check out the [examples](examples/) directory for complete working examples.

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
This project is licensed under the MIT License - see the LICENSE file for details.