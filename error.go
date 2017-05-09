package goframe

import "errors"

var (
	// ErrUnsupportedlength unsupported lengthFieldLength.
	ErrUnsupportedlength = errors.New("unsupported lengthFieldLength. (expected: 1, 2, 3, 4, or 8)")
	// ErrTooLessLength adjusted frame length is less than zero
	ErrTooLessLength = errors.New("Adjusted frame length is less than zero")
	// ErrUnexpectedFixedLength is not unexpected fixed length for writting fixed length frames.
	ErrUnexpectedFixedLength = errors.New("Unexpected fixed length")
)
