package goframe

import (
	"bufio"
	"net"
)

var (
	// CRLFBytes is a const bytes for crlf.
	CRLFBytes = []byte("\n")
)

type lineBasedFrameConn struct {
	c net.Conn
	r *bufio.Reader
	w *bufio.Writer
}

// NewLineBasedFrameConn returns a line based Frame conn.
func NewLineBasedFrameConn(conn net.Conn) FrameConn {
	return &lineBasedFrameConn{
		c: conn,
		r: bufio.NewReader(conn),
		w: bufio.NewWriter(conn),
	}
}

func (fc *lineBasedFrameConn) ReadFrame() ([]byte, error) {
	// var (
	// 	isPrefix bool
	// 	err      error
	// 	line, ln []byte
	// )

	// for isPrefix && err == nil {
	// 	line, isPrefix, err = fc.r.ReadLine()
	// 	ln = append(ln, line...)
	// 	if err != nil {
	// 		return ln, err
	// 	}
	// }

	data, err := fc.r.ReadBytes('\n')
	if err == nil && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	return data, err
}

func (fc *lineBasedFrameConn) WriteFrame(p []byte) error {
	_, err := fc.w.Write(p)
	if err != nil {
		return err
	}
	_, err = fc.w.Write(CRLFBytes)
	if err != nil {
		return err
	}
	fc.w.Flush()
	return nil
}

func (fc *lineBasedFrameConn) Close() error {
	return fc.c.Close()
}

func (fc *lineBasedFrameConn) Conn() net.Conn {
	return fc.c
}
