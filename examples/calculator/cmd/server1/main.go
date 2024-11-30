package main

import (
	"flag"
	"fmt"
	"log"
	rpc "stream-rpc"
	proto "stream-rpc/examples/calculator/proto"
	calculator "stream-rpc/examples/calculator/proto/service"
	stream "stream-rpc/stream/libp2p"
	"time"

	"crypto/rand"
	"encoding/hex"
	"io/ioutil"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
)

const protocolID = "/calculator/1.0.0"

func makeServerCalculation(peer *rpc.RpcPeer, a, b int32) {
	calculatorClient := proto.NewCalculatorClient(peer)
	// Test Add
	addReq := &proto.AddRequest{A: a, B: b}
	addResp := calculatorClient.Add(addReq)
	if addResp == nil {
		log.Printf("Server Add error: nil response\n")
		return
	}
	fmt.Printf("Server: %d + %d = %d\n", a, b, addResp.Result)

	// Test Multiply
	mulReq := &proto.MultiplyRequest{A: a, B: b}
	mulResp := calculatorClient.Multiply(mulReq)
	if mulResp == nil {
		log.Printf("Server Multiply error: nil response\n")
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

	// Register the calculator service using the generated Register function
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

func loadOrCreatePrivateKey(path string) (crypto.PrivKey, error) {
	// Try to load existing key
	if keyData, err := ioutil.ReadFile(path); err == nil {
		decoded, err := hex.DecodeString(string(keyData))
		if err != nil {
			return nil, err
		}
		return crypto.UnmarshalPrivateKey(decoded)
	}

	// Generate new key if none exists
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rand.Reader)
	if err != nil {
		return nil, err
	}

	// Save the new key
	keyBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(path, []byte(hex.EncodeToString(keyBytes)), 0600)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func main() {
	port := flag.Int("port", 9000, "port to listen on")
	wsPort := flag.Int("wsport", 8080, "WebSocket port to listen on")
	flag.Parse()

	// Create transport options
	var transports []libp2p.Option

	// Load or create private key before creating the host
	priv, err := loadOrCreatePrivateKey("node.key")
	if err != nil {
		log.Fatal(err)
	}

	// Add identity option to transports
	transports = append(transports,
		libp2p.Identity(priv),
		libp2p.Transport(websocket.New),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port),      // Regular TCP
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ws", *wsPort), // WebSocket
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
