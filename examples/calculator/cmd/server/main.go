package main

import (
	"flag"
	"fmt"
	"log"
	rpc "stream-rpc"
	proto "stream-rpc/examples/calculator/proto"
	calculator "stream-rpc/examples/calculator/service"
	stream "stream-rpc/stream/libp2p"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
)

const protocolID = "/calculator/1.0.0"

func makeServerCalculation(peer *rpc.RpcPeer, a, b int32) {
	// Test Add
	addReq := &proto.AddRequest{A: a, B: b}
	addResp := &proto.AddResponse{}
	if err := peer.Call("Calculator.Add", addReq, addResp); err != nil {
		log.Printf("Server Add error: %v\n", err)
		return
	}
	fmt.Printf("Server: %d + %d = %d\n", a, b, addResp.Result)

	// Test Multiply
	mulReq := &proto.MultiplyRequest{A: a, B: b}
	mulResp := &proto.MultiplyResponse{}
	if err := peer.Call("Calculator.Multiply", mulReq, mulResp); err != nil {
		log.Printf("Server Multiply error: %v\n", err)
		return
	}
	fmt.Printf("Server: %d * %d = %d\n", a, b, mulResp.Result)
}

func handleStream(s network.Stream) {
	log.Printf("New connection from: %s\n", s.Conn().RemotePeer())

	// Create the libp2p stream wrapper
	libp2pStream := stream.NewLibP2PStream(s)

	// Create a new RPC peer
	peer := rpc.NewRpcPeer(libp2pStream)
	defer peer.Close()

	// Register the calculator service
	peer.RegisterService("Calculator", &calculator.CalculatorService{})
	done := make(chan struct{})
	// Handle stream closure
	peer.OnStreamClose(func(err error) {
		if err != nil {
			log.Printf("Stream error: %v\n", err)
		} else {
			log.Println("Stream closed normally")
		}
		close(done)
	})

	// Start periodic calculations
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	counter := int32(100)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			makeServerCalculation(peer, counter, counter+5)
			counter += 5
		}
	}
}

func main() {
	port := flag.Int("port", 9000, "port to listen on")
	wsPort := flag.Int("wsport", 8080, "WebSocket port to listen on")
	flag.Parse()

	// Create transport options
	var transports []libp2p.Option

	// Add WebSocket transport
	wsAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ws", *wsPort)
	transports = append(transports,
		libp2p.Transport(websocket.New),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port), // Regular TCP
			wsAddr, // WebSocket
		),
	)

	// Create a new libp2p host with both TCP and WebSocket transports
	h, err := libp2p.New(transports...)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	// Print the host's addresses
	fmt.Printf("Node ID: %s\n", h.ID())
	fmt.Println("Listening addresses:")
	for _, addr := range h.Addrs() {
		fmt.Printf("  - %s/p2p/%s\n", addr, h.ID())
	}

	// Set up the RPC handler
	h.SetStreamHandler(protocolID, handleStream)

	// Keep the server running
	select {}
}
