//refer to:
// 1. https://github.com/nickpisacane/framedconn
// 2. https://github.com/netty/netty/blob/eb7f751ba519cbcab47d640cd18757f09d077b55/codec/src/main/java/io/netty/handler/codec/LengthFieldBasedFrameDecoder.java
// 3. https://github.com/netty/netty/blob/eb7f751ba519cbcab47d640cd18757f09d077b55/codec/src/main/java/io/netty/handler/codec/LengthFieldPrepender.java
package goframe

import (
	"net"
)

// FrameConn is a conn that can send and receive framed data.
type FrameConn interface {
	// Reads a "frame" from the connection.
	ReadFrame() ([]byte, error)

	// Writes a "frame" to the connection.
	WriteFrame(p []byte) error

	// Closes the connections, truncates any buffers.
	Close() error

	// Returns the underlying connection.
	Conn() net.Conn
}
