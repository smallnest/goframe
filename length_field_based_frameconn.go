package goframe

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
)

type lengthFieldBasedFrameConn struct {
	encoderConfig EncoderConfig
	decoderConfig DecoderConfig
	c             net.Conn
	r             *bufio.Reader
	w             *bufio.Writer
}

// EncoderConfig config for encoder.
type EncoderConfig struct {
	// ByteOrder is the ByteOrder of the length field.
	ByteOrder binary.ByteOrder
	// LengthFieldLength is the length of the length field.
	LengthFieldLength int
	// LengthAdjustment is the compensation value to add to the value of the length field
	LengthAdjustment int
	// LengthIncludesLengthFieldLength is true, the length of the prepended length field is added to the value of the prepended length field
	LengthIncludesLengthFieldLength bool
}

// DecoderConfig config for decoder.
type DecoderConfig struct {
	// ByteOrder is the ByteOrder of the length field.
	ByteOrder binary.ByteOrder
	// LengthFieldOffset is the offset of the length field
	LengthFieldOffset int
	// LengthFieldLength is the length of the length field
	LengthFieldLength int
	// LengthAdjustment is the compensation value to add to the value of the length field
	LengthAdjustment int
	// InitialBytesToStrip is the number of first bytes to strip out from the decoded frame
	InitialBytesToStrip int
}

// NewLengthFieldBasedFrameConn returns a wrapped Frame conn based on the length field.
// It is the go implementation of netty LengthFieldBasedFrameecoder and LengthFieldPrepender.
// you can see javadoc of them to learn more details.
func NewLengthFieldBasedFrameConn(encoderConfig EncoderConfig, decoderConfig DecoderConfig, conn net.Conn) FrameConn {
	return &lengthFieldBasedFrameConn{
		encoderConfig: encoderConfig,
		decoderConfig: decoderConfig,
		c:             conn,
		r:             bufio.NewReader(conn),
		w:             bufio.NewWriter(conn),
	}
}

func (fc *lengthFieldBasedFrameConn) ReadFrame() ([]byte, error) {
	var header []byte
	var err error
	if fc.decoderConfig.LengthFieldOffset > 0 { //discard header(offset)
		header, err = ReadN(fc.c, fc.decoderConfig.LengthFieldOffset)
		if err != nil {
			return nil, err
		}
	}

	lenBuf, frameLength, err := fc.getUnadjustedFrameLength()
	if err != nil {
		return nil, err
	}

	if fc.decoderConfig.LengthAdjustment > 0 { //discard adjust header
		_, err := ReadN(fc.c, fc.decoderConfig.LengthAdjustment)
		if err != nil {
			return nil, err
		}
	}

	// real message length
	msgLength := int(frameLength) + fc.decoderConfig.LengthAdjustment
	msg, err := ReadN(fc.r, msgLength)
	if err != nil {
		return nil, err
	}

	fullMessage := make([]byte, len(header)+len(lenBuf)+msgLength)
	copy(fullMessage, header)
	copy(fullMessage[len(header):], lenBuf)
	copy(fullMessage[len(header)+len(lenBuf):], msg)

	return fullMessage[fc.decoderConfig.InitialBytesToStrip:], nil
}

func (fc *lengthFieldBasedFrameConn) getUnadjustedFrameLength() (lenBuf []byte, n uint64, err error) {
	switch fc.decoderConfig.LengthFieldLength {
	case 1:
		b, err := fc.r.ReadByte()
		return []byte{b}, uint64(b), err
	case 2:
		lenBuf := make([]byte, 2)
		_, err = fc.r.Read(lenBuf)
		if err != nil {
			return nil, 0, err
		}
		return lenBuf, uint64(fc.decoderConfig.ByteOrder.Uint16(lenBuf)), nil
	case 3:
		lenBuf := make([]byte, 3)
		_, err = fc.r.Read(lenBuf)
		if err != nil {
			return nil, 0, err
		}
		return lenBuf, readUint24(fc.decoderConfig.ByteOrder, lenBuf), nil
	case 4:
		lenBuf := make([]byte, 4)
		_, err = fc.r.Read(lenBuf)
		if err != nil {
			return nil, 0, err
		}
		return lenBuf, uint64(fc.decoderConfig.ByteOrder.Uint32(lenBuf)), nil
	case 8:
		lenBuf := make([]byte, 8)
		_, err = fc.r.Read(lenBuf)
		if err != nil {
			return nil, 0, err
		}
		return lenBuf, uint64(fc.decoderConfig.ByteOrder.Uint64(lenBuf)), nil
	default:
		return nil, 0, ErrUnsupportedlength
	}
}

func readUint24(byteOrder binary.ByteOrder, b []byte) uint64 {
	_ = b[2]
	if byteOrder == binary.LittleEndian {
		return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16
	}

	return uint64(b[2]) | uint64(b[1])<<8 | uint64(b[0])<<16
}

func (fc *lengthFieldBasedFrameConn) WriteFrame(p []byte) error {
	length := len(p) + fc.encoderConfig.LengthAdjustment
	if fc.encoderConfig.LengthIncludesLengthFieldLength {
		length += fc.encoderConfig.LengthFieldLength
	}

	if length < 0 {
		return ErrTooLessLength
	}

	var err error
	switch fc.encoderConfig.LengthFieldLength {
	case 1:
		if length >= 256 {
			return fmt.Errorf("length does not fit into a byte: %d", length)
		}
		err = fc.w.WriteByte(byte(length))
		if err != nil {
			return err
		}
	case 2:
		if length >= 65536 {
			return fmt.Errorf("length does not fit into a short integer: %d", length)
		}
		buf := make([]byte, 2)
		fc.encoderConfig.ByteOrder.PutUint16(buf, uint16(length))
		_, err = fc.w.Write(buf)
		if err != nil {
			return err
		}
	case 3:
		if length >= 16777216 {
			return fmt.Errorf("length does not fit into a medium integer: %d", length)
		}
		buf := writeUint24(fc.encoderConfig.ByteOrder, length)
		_, err = fc.w.Write(buf)
		if err != nil {
			return err
		}
	case 4:
		buf := make([]byte, 4)
		fc.encoderConfig.ByteOrder.PutUint32(buf, uint32(length))
		_, err = fc.w.Write(buf)
		if err != nil {
			return err
		}
	case 8:
		buf := make([]byte, 8)
		fc.encoderConfig.ByteOrder.PutUint64(buf, uint64(length))
		_, err = fc.w.Write(buf)
		if err != nil {
			return err
		}
	default:
		return ErrUnsupportedlength
	}

	_, err = fc.w.Write(p)
	fc.w.Flush()
	return err
}

func writeUint24(byteOrder binary.ByteOrder, v int) []byte {
	b := make([]byte, 3)
	if byteOrder == binary.LittleEndian {
		b[0] = byte(v)
		b[1] = byte(v >> 8)
		b[2] = byte(v >> 16)
	} else {
		b[2] = byte(v)
		b[1] = byte(v >> 8)
		b[0] = byte(v >> 16)
	}
	return b
}

func (fc *lengthFieldBasedFrameConn) Close() error {
	return fc.c.Close()
}

func (fc *lengthFieldBasedFrameConn) Conn() net.Conn {
	return fc.c
}
