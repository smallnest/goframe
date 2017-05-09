package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"time"

	"github.com/smallnest/goframe"
)

func main() {
	l, err := net.Listen("tcp", ":9981")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	encoderConfig := goframe.EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               4,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}

	decoderConfig := goframe.DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    0,
		InitialBytesToStrip: 4,
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		c := goframe.NewLengthFieldBasedFrameConn(encoderConfig, decoderConfig, conn)
		go func(conn goframe.FrameConn) {
			for {
				b, err := c.ReadFrame()
				if err != nil {
					if err == io.EOF {
						return
					}
					panic(err)
				}
				fmt.Println(string(b))

				s := fmt.Sprintf("%d: %s", time.Now().UnixNano()/1e6, string(b))
				c.WriteFrame([]byte(s))
			}
		}(c)
	}
}
