package libp2p

import (
	"github.com/libp2p/go-libp2p/core/network"
)

type LibP2PStream struct {
	stream network.Stream
}

func NewLibP2PStream(stream network.Stream) *LibP2PStream {
	return &LibP2PStream{
		stream: stream,
	}
}

func (s *LibP2PStream) Read(p []byte) (n int, err error) {
	return s.stream.Read(p)
}

func (s *LibP2PStream) Write(p []byte) (n int, err error) {
	n, err = s.stream.Write(p)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (s *LibP2PStream) Close() error {
	return s.stream.Close()
}
