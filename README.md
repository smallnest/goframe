# goframe
[goframe](https://github.com/smallnest/goframe) provides wrapped net.Conn that can send and receive framed data.

[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/smallnest/goframe?status.png)](http://godoc.org/github.com/smallnest/goframe)  [![travis](https://travis-ci.org/smallnest/goframe.svg?branch=master)](https://travis-ci.org/smallnest/goframe) [![Go Report Card](https://goreportcard.com/badge/github.com/smallnest/goframe)](https://goreportcard.com/report/github.com/smallnest/goframe)


### wrapped conn interface
goframe contains one FrameConn interface and some concrete FrameConn implementations:

```go
type FrameConn interface {
	ReadFrame() ([]byte, error)
	WriteFrame(p []byte) error
	Close() error
	Conn() net.Conn
}
```

You can use `NewXXXXXXX` function to get a concrete FrameConn.

When you want to send a frame, you use `WriteFrame`.
When you want to receve a frame, you use `ReadFrame`.


### examples

Currently I have implemented some FrameConns that has same formats to Netty frame decoders and encoders.
So we can use Go and popular netty framework to communicate. I have created [some examples](https://github.com/smallnest/goframe/_examples) to demonstrate it.

Those examples contains client and server for both go and netty, so you can use any client and any server to start.


### concrete FrameConn implementations

1. **FixedLengthFrameConn**: each frame has fixed number of bytes. `ReadFrame` returns the fixed length bytes and `WriteFrame` can write `n * fixed length` bytes.
2. **LineBasedFrameConn**: split data with line separator (\n or \r\n). `WriteFrame` writes data p and appends "\r\n"
3. **DelimiterBasedFrameConn**: split data with a specific delimiter.
4. **lengthFieldBasedFrameConn**: it is a complex FrameConn. It contains several parameters and can produce many frame format. I has copy descriptions from [Netty javadoc](https://netty.io/4.1/api/io/netty/handler/codec/LengthFieldBasedFrameDecoder.html):

Here are some decoder example that will give you the basic idea on which option does what.

#### 2 bytes length field at offset 0, do not strip header

The value of the length field in this example is 12 (0x0C) which represents the length of "HELLO, WORLD". By default, the decoder assumes that the length field represents the number of the bytes that follows the length field. Therefore, it can be decoded with the simplistic parameter combination.

```
 lengthFieldOffset   = 0
 lengthFieldLength   = 2
 lengthAdjustment    = 0
 initialBytesToStrip = 0 (= do not strip header)


 BEFORE DECODE (14 bytes)         AFTER DECODE (14 bytes)
 +--------+----------------+      +--------+----------------+
 | Length | Actual Content |----->| Length | Actual Content |
 | 0x000C | "HELLO, WORLD" |      | 0x000C | "HELLO, WORLD" |
 +--------+----------------+      +--------+----------------+
```

#### 2 bytes length field at offset 0, strip header

Because we can get the length of the content by calling ByteBuf.readableBytes(), you might want to strip the length field by specifying initialBytesToStrip. In this example, we specified 2, that is same with the length of the length field, to strip the first two bytes.

```
 lengthFieldOffset   = 0
 lengthFieldLength   = 2
 lengthAdjustment    = 0
 initialBytesToStrip = 2 (= the length of the Length field)

 BEFORE DECODE (14 bytes)         AFTER DECODE (12 bytes)
 +--------+----------------+      +----------------+
 | Length | Actual Content |----->| Actual Content |
 | 0x000C | "HELLO, WORLD" |      | "HELLO, WORLD" |
 +--------+----------------+      +----------------+
```

#### 2 bytes length field at offset 0, do not strip header, the length field represents the length of the whole message

In most cases, the length field represents the length of the message body only, as shown in the previous examples. However, in some protocols, the length field represents the length of the whole message, including the message header. In such a case, we specify a non-zero lengthAdjustment. Because the length value in this example message is always greater than the body length by 2, we specify -2 as lengthAdjustment for compensation.

```
 lengthFieldOffset   =  0
 lengthFieldLength   =  2
 lengthAdjustment    = -2 (= the length of the Length field)
 initialBytesToStrip =  0

 BEFORE DECODE (14 bytes)         AFTER DECODE (14 bytes)
 +--------+----------------+      +--------+----------------+
 | Length | Actual Content |----->| Length | Actual Content |
 | 0x000E | "HELLO, WORLD" |      | 0x000E | "HELLO, WORLD" |
 +--------+----------------+      +--------+----------------+
```

#### 3 bytes length field at the end of 5 bytes header, do not strip header

The following message is a simple variation of the first example. An extra header value is prepended to the message. lengthAdjustment is zero again because the decoder always takes the length of the prepended data into account during frame length calculation.

```
 lengthFieldOffset   = 2 (= the length of Header 1)
 lengthFieldLength   = 3
 lengthAdjustment    = 0
 initialBytesToStrip = 0

 BEFORE DECODE (17 bytes)                      AFTER DECODE (17 bytes)
 +----------+----------+----------------+      +----------+----------+----------------+
 | Header 1 |  Length  | Actual Content |----->| Header 1 |  Length  | Actual Content |
 |  0xCAFE  | 0x00000C | "HELLO, WORLD" |      |  0xCAFE  | 0x00000C | "HELLO, WORLD" |
 +----------+----------+----------------+      +----------+----------+----------------+
```

#### 3 bytes length field at the beginning of 5 bytes header, do not strip header

This is an advanced example that shows the case where there is an extra header between the length field and the message body. You have to specify a positive lengthAdjustment so that the decoder counts the extra header into the frame length calculation.

```
 lengthFieldOffset   = 0
 lengthFieldLength   = 3
 lengthAdjustment    = 2 (= the length of Header 1)
 initialBytesToStrip = 0

 BEFORE DECODE (17 bytes)                      AFTER DECODE (17 bytes)
 +----------+----------+----------------+      +----------+----------+----------------+
 |  Length  | Header 1 | Actual Content |----->|  Length  | Header 1 | Actual Content |
 | 0x00000C |  0xCAFE  | "HELLO, WORLD" |      | 0x00000C |  0xCAFE  | "HELLO, WORLD" |
 +----------+----------+----------------+      +----------+----------+----------------+
``` 

#### 2 bytes length field at offset 1 in the middle of 4 bytes header, strip the first header field and the length field

This is a combination of all the examples above. There are the prepended header before the length field and the extra header after the length field. The prepended header affects the lengthFieldOffset and the extra header affects the lengthAdjustment. We also specified a non-zero initialBytesToStrip to strip the length field and the prepended header from the frame. If you don't want to strip the prepended header, you could specify 0 for initialBytesToSkip.

```
 lengthFieldOffset   = 1 (= the length of HDR1)
 lengthFieldLength   = 2
 lengthAdjustment    = 1 (= the length of HDR2)
 initialBytesToStrip = 3 (= the length of HDR1 + LEN)

 BEFORE DECODE (16 bytes)                       AFTER DECODE (13 bytes)
 +------+--------+------+----------------+      +------+----------------+
 | HDR1 | Length | HDR2 | Actual Content |----->| HDR2 | Actual Content |
 | 0xCA | 0x000C | 0xFE | "HELLO, WORLD" |      | 0xFE | "HELLO, WORLD" |
 +------+--------+------+----------------+      +------+----------------+
```

#### 2 bytes length field at offset 1 in the middle of 4 bytes header, strip the first header field and the length field, the length field represents the length of the whole message

Let's give another twist to the previous example. The only difference from the previous example is that the length field represents the length of the whole message instead of the message body, just like the third example. We have to count the length of HDR1 and Length into lengthAdjustment. Please note that we don't need to take the length of HDR2 into account because the length field already includes the whole header length.

```
 lengthFieldOffset   =  1
 lengthFieldLength   =  2
 lengthAdjustment    = -3 (= the length of HDR1 + LEN, negative)
 initialBytesToStrip =  3

 BEFORE DECODE (16 bytes)                       AFTER DECODE (13 bytes)
 +------+--------+------+----------------+      +------+----------------+
 | HDR1 | Length | HDR2 | Actual Content |----->| HDR2 | Actual Content |
 | 0xCA | 0x0010 | 0xFE | "HELLO, WORLD" |      | 0xFE | "HELLO, WORLD" |
 +------+--------+------+----------------+      +------+----------------+
``` 

For encoder, it prepends the length of the message. The length value is prepended as a binary form.
For example, LengthFieldLength(2) will encode the following 12-bytes string:

```
 +----------------+
 | "HELLO, WORLD" |
 +----------------+
```

into the following:

```
 +--------+----------------+
 + 0x000C | "HELLO, WORLD" |
 +--------+----------------+
```

If you turned on the lengthIncludesLengthFieldLength flag in the EncoderConfig, the encoded data would look like the following (12 (original data) + 2 (prepended data) = 14 (0xE)):

```
 +--------+----------------+
 + 0x000E | "HELLO, WORLD" |
 +--------+----------------+
```


