package tcp

import "net"

type TCPStream struct {
	Conn net.Conn
}

func NewTCPStream(conn net.Conn) *TCPStream {
	return &TCPStream{Conn: conn}
}

func (s *TCPStream) Read(p []byte) (n int, err error) {
	return s.Conn.Read(p)
}

func (s *TCPStream) Write(p []byte) (n int, err error) {
	return s.Conn.Write(p)
}

func (s *TCPStream) Close() error {
	return s.Conn.Close()
}
