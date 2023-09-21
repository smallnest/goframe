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
	rm          sync.Mutex
	wm          sync.Mutex
}

// NewFixedLengthFrameConn returns a fixed length Frame conn.
func NewFixedLengthFrameConn(frameLength int, conn net.Conn) FrameConn {
	return &fixedLengthFrameConn{
		frameLength: frameLength,
		c:           conn,
		r:           bufio.NewReader(conn),
		w:           bufio.NewWriter(conn),
		rm:          sync.Mutex{},
		wm:          sync.Mutex{},
	}
}
func (fc *fixedLengthFrameConn) ReadFrame() ([]byte, error) {
	fc.rm.Lock()
	defer fc.rm.Unlock()
	buf := make([]byte, fc.frameLength)
	_, err := io.ReadFull(fc.r, buf)
	return buf, err
}

func (fc *fixedLengthFrameConn) WriteFrame(p []byte) error {
	fc.wm.Lock()
	defer fc.wm.Unlock()
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
