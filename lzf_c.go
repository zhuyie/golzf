// +build lzf_cgo

// Package lzf implements LZF compression algorithm.
package lzf

// #cgo CFLAGS: -O3 -DHLOG=14
// #include "src/lzf_c.c"
// #include "src/lzf_d.c"
import "C"

import "unsafe"

func p(in []byte) unsafe.Pointer {
	if len(in) == 0 {
		return unsafe.Pointer(nil)
	}
	return unsafe.Pointer(&in[0])
}

func clen(s []byte) C.uint {
	return C.uint(len(s))
}

// Compress compresses `input` and puts the content in `output`.
// len(output) should have enough space for the compressed data.
// Returns the number of bytes in the `output` slice.
func Compress(input, output []byte) (outSize int, err error) {
	outSize = int(C.lzf_compress(p(input), clen(input), p(output), clen(output)))
	if outSize == 0 {
		err = ErrInsufficientBuffer
	}
	return
}

// Decompress decompresses `input` and puts the content in `output`.
// len(output) should have enough space for the uncompressed data.
// Returns the number of bytes in the `output` slice.
func Decompress(input, output []byte) (outSize int, err error) {
	var errCode C.int
	outSize = int(C.lzf_decompress(p(input), clen(input), p(output), clen(output), &errCode))
	if outSize > 0 {
		return
	}
	switch errCode {
	case C.E2BIG:
		err = ErrInsufficientBuffer
	case C.EINVAL:
		err = ErrDataCorruption
	default:
		err = ErrUnknown
	}
	return
}
