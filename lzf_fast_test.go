// +build !lzf_cgo

package lzf

import (
	"strings"
	"testing"
)

func BenchmarkCompressFast(b *testing.B) {
	input := []byte(strings.Repeat("Hello world, this is quite something", 5))
	output := make([]byte, len(input)-1)

	var outSize int
	var err error
	var htab = make([]uint32, htabSize)
	for n := 0; n < b.N; n++ {
		outSize, err = CompressFast(input, output, htab)
		if err != nil {
			b.Fatalf("CompressFast failed: %v", err)
		}
	}
	output = output[:outSize]

	decompressed := make([]byte, len(input))
	outSize, err = Decompress(output, decompressed)
	if err != nil {
		b.Fatalf("Decompress failed: %v", err)
	}
	if string(decompressed) != string(input) {
		b.Fatalf("Decompress failed: decompressed != input: %q != %q", decompressed, input)
	}
}
