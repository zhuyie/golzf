package lzf

import (
	"strings"
	"testing"
)

func TestCompression(t *testing.T) {
	input := []byte(strings.Repeat("Hello world, this is quite something", 10))
	output := make([]byte, CompressBound(input))
	outSize, err := Compress(input, output)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Compress failed: Output buffer is empty")
	}
	output = output[:outSize]

	decompressed := make([]byte, len(input))
	outSize, err = Decompress(output, decompressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}
	if int(outSize) != len(input) {
		t.Fatalf("Decompress failed: expected outSize %v, got %v", len(input), outSize)
	}
	if string(decompressed) != string(input) {
		t.Fatalf("Decompress failed: output != input: %q != %q", decompressed, input)
	}
}

func TestNoCompression(t *testing.T) {
	input := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	output := make([]byte, CompressBound(input))
	outSize, err := Compress(input, output)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Compress failed: Output buffer is empty")
	}
	output = output[:outSize]

	decompressed := make([]byte, len(input))
	outSize, err = Decompress(output, decompressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}
	if int(outSize) != len(input) {
		t.Fatalf("Decompress failed: expected outSize %v, got %v", len(input), outSize)
	}
	if string(decompressed) != string(input) {
		t.Fatalf("Decompress failed: output != input: %q != %q", decompressed, input)
	}
}

func TestCompressionError(t *testing.T) {
	input := []byte(strings.Repeat("Hello world, this is quite something", 10))
	output := make([]byte, 1)
	_, err := Compress(input, output)
	if err == nil {
		t.Fatalf("Compress should have failed but didn't")
	}
}

func TestDecompressionError(t *testing.T) {
	input := []byte(strings.Repeat("Hello world, this is quite something", 10))
	output := make([]byte, CompressBound(input))
	outSize, err := Compress(input, output)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	if outSize == 0 {
		t.Fatal("Compress failed: Output buffer is empty")
	}
	output = output[:outSize]

	decompressed := make([]byte, len(input)-1)
	outSize, err = Decompress(output, decompressed)
	if err == nil {
		t.Fatalf("Decompress should have failed")
	}

	decompressed = make([]byte, len(input))
	output[0] = output[0] + 10
	outSize, err = Decompress(output, decompressed)
	if err == nil {
		t.Fatalf("Decompress should have failed")
	}
	output[0] = output[0] - 10
	outSize, err = Decompress(output, decompressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}
}
