# Lightweight RPC Framework

A simple, efficient RPC framework for Go applications with Protocol Buffers support and libp2p integration.

## Features
- Simple wire protocol for efficient communication
- Protocol Buffers integration
- Support for unary and streaming RPCs
- libp2p transport layer support
- Automatic code generation
- Easy-to-use API

## Installation
```bash
go get github.com/jibuji/stream-rpc
go install github.com/jibuji/stream-rpc/cmd/protoc-gen-stream-rpc
```

## Quick Start

### 1. Define your service in Protocol Buffers
```protobuf
syntax = "proto3";

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
```

### 2. Generate code
```bash
protoc --go_out=. --stream-rpc_out=. calculator.proto
```

### 3. Implement server
```go
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

### 4. Create client
```go
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
    addResp := calculatorClient.Add(&proto.AddRequest{A: 5, B: 3})
    log.Printf("5 + 3 = %d\n", addResp.Result)
}
```

## Documentation
- [Architecture Overview](docs/architecture.md)
- [Getting Started Guide](docs/getting_started.md)
- [Wire Protocol Specification](docs/wire_protocol.md)

## Examples
Check out the [examples](examples/) directory for complete working examples.

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
This project is licensed under the MIT License - see the LICENSE file for details.