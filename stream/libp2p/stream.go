package libp2p

import (
	"github.com/libp2p/go-libp2p/core/network"
)

type LibP2PStream struct {
	Stream network.Stream
}

func NewLibP2PStream(stream network.Stream) *LibP2PStream {
	return &LibP2PStream{
		Stream: stream,
	}
}

func (s *LibP2PStream) Read(p []byte) (n int, err error) {
	return s.Stream.Read(p)
}

func (s *LibP2PStream) Write(p []byte) (n int, err error) {
	n, err = s.Stream.Write(p)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (s *LibP2PStream) Close() error {
	return s.Stream.Close()
}
