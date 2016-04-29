package lzf

import "errors"

// error codes
var (
	ErrInsufficientBuffer = errors.New("insufficient output buffer")
	ErrDataCorruption     = errors.New("invalid compressed data")
	ErrUnknown            = errors.New("unknown error")
)
