package tcp

import "net"

type TCPStream struct {
	conn net.Conn
}

func NewTCPStream(conn net.Conn) *TCPStream {
	return &TCPStream{conn: conn}
}

func (s *TCPStream) Read(p []byte) (n int, err error) {
	return s.conn.Read(p)
}

func (s *TCPStream) Write(p []byte) (n int, err error) {
	return s.conn.Write(p)
}

func (s *TCPStream) Close() error {
	return s.conn.Close()
}
