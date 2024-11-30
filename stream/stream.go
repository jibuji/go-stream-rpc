package stream

import (
	"io"
)

// Stream represents a bidirectional communication channel
type Stream interface {
	io.Reader
	io.Writer
	io.Closer
}
