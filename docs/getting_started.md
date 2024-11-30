# Getting Started

## Prerequisites
- Go 1.16 or higher
- Protocol Buffers compiler (protoc)
- This framework's protoc plugin

## Installation
```bash
go get github.com/jibuji/go-stream-rpc
go install github.com/jibuji/go-stream-rpc/cmd/protoc-gen-stream-rpc
```

## Basic Usage

### 1. Define Your Service
Create a .proto file:
```protobuf
syntax = "proto3";

package calculator;

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

### 2. Generate Code
```bash
protoc --go_out=. --stream-rpc_out=. calculator.proto
```

### 3. Implement Server
```go
package main

import (
    "flag"
    "log"
    rpc "stream-rpc"
    proto "stream-rpc/examples/calculator/proto"
    calculator "stream-rpc/examples/calculator/proto/service"
    stream "stream-rpc/stream/libp2p"
)

func handleStream(s network.Stream) {
    // Create the libp2p stream wrapper
    libp2pStream := stream.NewLibP2PStream(s)

    // Create a new RPC peer
    peer := rpc.NewRpcPeer(libp2pStream)
    defer peer.Close()

    // Register the calculator service
    peer.RegisterService("Calculator", &calculator.CalculatorService{})

    // Handle stream closure
    peer.OnStreamClose(func(err error) {
        if err != nil {
            log.Printf("Stream error: %v\n", err)
        }
    })
}

func main() {
    port := flag.Int("port", 9000, "port to listen on")
    flag.Parse()

    // Create a new libp2p host
    h, err := libp2p.New(
        libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port)),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer h.Close()

    // Set up the RPC handler
    h.SetStreamHandler("/calculator/1.0.0", handleStream)

    // Keep the server running
    select {}
}
```

### 4. Create Client
```go
package main

import (
    "context"
    "flag"
    "log"
    rpc "stream-rpc"
    proto "stream-rpc/examples/calculator/proto"
    stream "stream-rpc/stream/libp2p"
)

func main() {
    targetPeer := flag.String("peer", "", "peer address in multiaddr format")
    flag.Parse()

    // Create a new libp2p host
    h, err := libp2p.New()
    if err != nil {
        log.Fatal(err)
    }
    defer h.Close()

    // Connect to peer and open stream
    maddr, _ := multiaddr.NewMultiaddr(*targetPeer)
    info, _ := peer.AddrInfoFromP2pAddr(maddr)
    ctx := context.Background()
    h.Connect(ctx, *info)
    s, _ := h.NewStream(ctx, info.ID, "/calculator/1.0.0")

    // Create RPC peer
    libp2pStream := stream.NewLibP2PStream(s)
    peer := rpc.NewRpcPeer(libp2pStream)
    defer peer.Close()

    // Create calculator client
    calculatorClient := proto.NewCalculatorClient(peer)

    // Make RPC calls
    addResp := calculatorClient.Add(&proto.AddRequest{A: 5, B: 3})
    log.Printf("5 + 3 = %d\n", addResp.Result)

    mulResp := calculatorClient.Multiply(&proto.MultiplyRequest{A: 5, B: 3})
    log.Printf("5 * 3 = %d\n", mulResp.Result)
}