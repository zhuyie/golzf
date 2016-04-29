// +build !lzf_cgo

// Package lzf implements LZF compression algorithm.
package lzf

const (
	htabLog  uint32 = 14
	htabSize uint32 = 1 << htabLog
	maxLit          = 1 << 5
	maxOff          = 1 << 13
	maxRef          = (1 << 8) + (1 << 3)

	// HashTableSize is the size of hashtable
	HashTableSize = htabSize
)

// Compress compresses `input` and puts the content in `output`.
// len(output) should have enough space for the compressed data.
// Returns the number of bytes in the `output` slice.
func Compress(input, output []byte) (int, error) {
	return CompressFast(input, output, nil)
}

// CompressFast compresses `input` and puts the content in `output`.
// len(output) should have enough space for the compressed data.
// Returns the number of bytes in the `output` slice.
func CompressFast(input, output []byte, htab []uint32) (int, error) {

	var hval, ref, hslot, off uint32
	var inputIndex, outputIndex, lit int

	inputLength := len(input)
	if inputLength == 0 {
		return 0, nil
	}
	outputLength := len(output)
	if outputLength == 0 {
		return 0, ErrInsufficientBuffer
	}

	if htab == nil {
		htab = make([]uint32, htabSize)
	} else {
		/* turn on the following memset will make compression deterministic and SLOWER */
		//for i := range htab {
		//	htab[i] = 0
		//}
	}

	lit = 0 /* start run */
	outputIndex++

	hval = uint32(input[inputIndex])<<8 | uint32(input[inputIndex+1])
	for inputIndex < inputLength-2 {
		hval = (hval << 8) | uint32(input[inputIndex+2])
		hslot = ((hval >> (3*8 - htabLog)) - hval*5) & (htabSize - 1)
		ref = htab[hslot]
		htab[hslot] = uint32(inputIndex)
		off = uint32(inputIndex) - ref - 1

		if off < maxOff &&
			(ref > 0) &&
			(input[ref] == input[inputIndex]) &&
			(input[ref+1] == input[inputIndex+1]) &&
			(input[ref+2] == input[inputIndex+2]) {

			/* match found at *ref++ */
			len := 2
			maxLen := inputLength - inputIndex - len
			if maxLen > maxRef {
				maxLen = maxRef
			}

			if outputIndex+3+1 >= outputLength { /* first a faster conservative test */
				nlit := 0
				if lit == 0 {
					nlit = 1
				}
				if outputIndex-nlit+3+1 >= outputLength { /* second the exact but rare test */
					return 0, ErrInsufficientBuffer
				}
			}

			output[outputIndex-lit-1] = byte(lit - 1) /* stop run */
			if lit == 0 {
				outputIndex-- /* undo run if length is zero */
			}

			for {
				len++
				if (len >= maxLen) || (input[int(ref)+len] != input[inputIndex+len]) {
					break
				}
			}

			len -= 2 /* len is now #octets - 1 */
			inputIndex++

			if len < 7 {
				output[outputIndex] = byte((off >> 8) + uint32(len<<5))
				outputIndex++
			} else {
				output[outputIndex] = byte((off >> 8) + (7 << 5))
				output[outputIndex+1] = byte(len - 7)
				outputIndex += 2
			}

			output[outputIndex] = byte(off)
			outputIndex += 2
			lit = 0 /* start run */

			inputIndex += len + 1

			if inputIndex >= inputLength-2 {
				break
			}

			inputIndex -= 2

			hval = uint32(input[inputIndex])<<8 | uint32(input[inputIndex+1])
			hval = (hval << 8) | uint32(input[inputIndex+2])
			hslot = ((hval >> (3*8 - htabLog)) - (hval * 5)) & (htabSize - 1)
			htab[hslot] = uint32(inputIndex)
			inputIndex++

			hval = (hval << 8) | uint32(input[inputIndex+2])
			hslot = ((hval >> (3*8 - htabLog)) - (hval * 5)) & (htabSize - 1)
			htab[hslot] = uint32(inputIndex)
			inputIndex++

		} else {
			/* one more literal byte we must copy */
			if outputIndex >= outputLength {
				return 0, ErrInsufficientBuffer
			}

			lit++
			output[outputIndex] = input[inputIndex]
			outputIndex++
			inputIndex++

			if lit == maxLit {
				output[outputIndex-lit-1] = byte(lit - 1) /* stop run */
				lit = 0                                   /* start run */
				outputIndex++
			}
		}
	}

	if outputIndex+3 >= outputLength { /* at most 3 bytes can be missing here */
		return 0, ErrInsufficientBuffer
	}

	for inputIndex < inputLength {
		lit++
		output[outputIndex] = input[inputIndex]
		outputIndex++
		inputIndex++

		if lit == maxLit {
			output[outputIndex-lit-1] = byte(lit - 1) /* stop run */
			lit = 0                                   /* start run */
			outputIndex++
		}
	}

	output[outputIndex-lit-1] = byte(lit - 1) /* end run */
	if lit == 0 {                             /* undo run if length is zero */
		outputIndex--
	}

	return outputIndex, nil
}

// Decompress decompresses `input` and puts the content in `output`.
// len(output) should have enough space for the uncompressed data.
// Returns the number of bytes in the `output` slice.
func Decompress(input, output []byte) (int, error) {

	var inputIndex, outputIndex int

	inputLength := len(input)
	outputLength := len(output)
	if inputLength == 0 {
		return 0, nil
	}

	for inputIndex < inputLength {
		ctrl := int(input[inputIndex])
		inputIndex++

		if ctrl < (1 << 5) { /* literal run */
			ctrl++

			if outputIndex+ctrl > outputLength {
				return 0, ErrInsufficientBuffer
			}

			if inputIndex+ctrl > inputLength {
				return 0, ErrDataCorruption
			}

			copy(output[outputIndex:outputIndex+ctrl], input[inputIndex:inputIndex+ctrl])
			inputIndex += ctrl
			outputIndex += ctrl

		} else { /* back reference */
			length := ctrl >> 5
			ref := outputIndex - ((ctrl & 0x1f) << 8) - 1

			if inputIndex >= inputLength {
				return 0, ErrDataCorruption
			}

			if length == 7 {
				length += int(input[inputIndex])
				inputIndex++

				if inputIndex >= inputLength {
					return 0, ErrDataCorruption
				}
			}

			ref -= int(input[inputIndex])
			inputIndex++

			if outputIndex+length+2 > outputLength {
				return 0, ErrInsufficientBuffer
			}

			if ref < 0 {
				return 0, ErrDataCorruption
			}

			// Can't use copy(...) here, because it has special handling when source and destination overlap.
			for i := 0; i < length+2; i++ {
				output[outputIndex+i] = output[ref+i]
			}
			outputIndex += length + 2
		}
	}

	return outputIndex, nil
}
