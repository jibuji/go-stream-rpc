package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	rpc "github.com/jibuji/go-stream-rpc"
	proto "github.com/jibuji/go-stream-rpc/examples/calculator/proto"
	calculator "github.com/jibuji/go-stream-rpc/examples/calculator/proto/service"
	stream "github.com/jibuji/go-stream-rpc/stream/libp2p"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const protocolID = "/calculator/1.0.0"

func main() {
	targetPeer := flag.String("peer", "", "peer address in multiaddr format")
	flag.Parse()

	if *targetPeer == "" {
		log.Fatal("Please provide a peer address with -peer")
	}

	// Create a new libp2p host
	h, err := libp2p.New()
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	// Parse the target peer address
	maddr, err := multiaddr.NewMultiaddr(*targetPeer)
	if err != nil {
		log.Fatal(err)
	}

	// Extract the peer ID from the multiaddr
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the peer
	ctx := context.Background()
	if err := h.Connect(ctx, *info); err != nil {
		log.Fatal(err)
	}

	// Open a stream
	s, err := h.NewStream(ctx, info.ID, protocolID)
	if err != nil {
		log.Fatal(err)
	}
	// Create the libp2p stream wrapper
	libp2pStream := stream.NewLibP2PStream(s)

	// Create RPC peer
	peer := rpc.NewRpcPeer(libp2pStream)
	defer peer.Close()

	// Register the calculator service using the generated Register function
	// peer.RegisterService("Calculator", &calculator.CalculatorService{})
	proto.RegisterCalculatorServer(peer, &calculator.CalculatorService{})
	// Create calculator client
	calculatorClient := proto.NewCalculatorClient(peer)

	// Handle stream closure
	done := make(chan struct{})
	peer.OnStreamClose(func(err error) {
		if err != nil {
			log.Printf("Stream error: %v\n", err)
		} else {
			log.Println("Stream closed normally")
		}
		close(done)
	})

	// Make periodic RPC calls
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for i := 0; ; i++ {
		select {
		case <-ticker.C:
			a, b := int32(i), int32(i+1)

			// Much simpler RPC calls
			addResp := calculatorClient.Add(&proto.AddRequest{A: a, B: b})
			if addResp == nil {
				log.Printf("Add error: nil response\n")
				continue
			}
			fmt.Printf("Client: %d + %d = %d\n", a, b, addResp.Result)

			mulResp := calculatorClient.Multiply(&proto.MultiplyRequest{A: a, B: b})
			if mulResp == nil {
				log.Printf("Multiply error: nil response\n")
				continue
			}
			fmt.Printf("Client: %d * %d = %d\n", a, b, mulResp.Result)
		case <-done:
			return
		}
	}
}
