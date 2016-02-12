// Package lzf implements compression using LibLZF.
package lzf

// #cgo CFLAGS: -O3 -DHLOG=14
// #include "src/lzf_c.c"
// #include "src/lzf_d.c"
import "C"

import (
	"errors"
	"unsafe"
)

func p(in []byte) unsafe.Pointer {
	if len(in) == 0 {
		return unsafe.Pointer(nil)
	}
	return unsafe.Pointer(&in[0])
}

func clen(s []byte) C.uint {
	return C.uint(len(s))
}

// CompressBound calculates the size of the output buffer needed by Compress.
func CompressBound(input []byte) int {
	// should less than 104% of the original size
	return int(float64(len(input))*1.04) + 1
}

// Compress compresses `input` and puts the content in `output`.
// len(output) should have enough space for the compressed data.
// Returns the number of bytes in the `output` slice.
func Compress(input, output []byte) (outSize uint, err error) {
	outSize = uint(C.lzf_compress(p(input), clen(input), p(output), clen(output)))
	if outSize == 0 {
		err = errors.New("insufficient space for compression")
	}
	return
}

// Decompress decompresses `input` and puts the content in `output`.
// len(output) should have enough space for the uncompressed data.
// Returns the number of bytes in the `output` slice.
func Decompress(input, output []byte) (outSize uint, err error) {
	var errCode C.int = 0
	outSize = uint(C.lzf_decompress(p(input), clen(input), p(output), clen(output), &errCode))
	if outSize > 0 {
		return
	}
	switch errCode {
	case C.E2BIG:
		err = errors.New("insufficient space for decompression")
	case C.EINVAL:
		err = errors.New("invalid compressed data")
	default:
		err = errors.New("unknown error")
	}
	return
}
