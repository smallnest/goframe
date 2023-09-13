package goframe

import (
	"bufio"
	"io"
	"net"
	"sync"
)

type fixedLengthFrameConn struct {
	frameLength int
	c           net.Conn
	r           *bufio.Reader
	w           *bufio.Writer
	m           sync.RWMutex
}

// NewFixedLengthFrameConn returns a fixed length Frame conn.
func NewFixedLengthFrameConn(frameLength int, conn net.Conn) FrameConn {
	return &fixedLengthFrameConn{
		frameLength: frameLength,
		c:           conn,
		r:           bufio.NewReader(conn),
		w:           bufio.NewWriter(conn),
		m:           sync.RWMutex{},
	}
}
func (fc *fixedLengthFrameConn) ReadFrame() ([]byte, error) {
	fc.m.RLock()
	defer fc.m.RUnlock()
	buf := make([]byte, fc.frameLength)
	_, err := io.ReadFull(fc.r, buf)
	return buf, err
}

func (fc *fixedLengthFrameConn) WriteFrame(p []byte) error {
	fc.m.Lock()
	defer fc.m.Unlock()
	l := len(p)
	if l%fc.frameLength != 0 {
		return ErrUnexpectedFixedLength
	}

	for i := 0; i < l; i += fc.frameLength {
		_, err := fc.w.Write(p[i : i+fc.frameLength])
		if err != nil {
			return err
		}
	}

	fc.w.Flush()
	return nil
}

func (fc *fixedLengthFrameConn) Close() error {
	return fc.c.Close()
}

func (fc *fixedLengthFrameConn) Conn() net.Conn {
	return fc.c
}
