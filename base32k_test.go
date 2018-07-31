/* CC0 - free software.
To the extent possible under law, all copyright and related or neighboring
rights to this work are waived. See the LICENSE file for more information. */

package base32k

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

// 00000000 11111111 00000000 11111111 10101010 01010101 10101010 01010101 11111111 10100101 01011010 11110000 00001111 10101010 01010101 00000000
var srcData = []byte{0x00, 0xff, 0x00, 0xff, 0xaa, 0x55, 0xaa, 0x55, 0xff, 0xa5, 0x5a, 0xf0, 0x0f, 0xaa, 0x55, 0x00}

var encodeExpectedBytes = map[int][]byte{
	//    0     1     2     3     4     5     6     7     8     9     10    11    12    13    14    15    16    17    18    19    20    21    22    23    24    25    26    27
	// ... .. ..00 00 0000  010101011010101 000001111111100 000101101010100 101111111110101 010110101010010 101011010101011 111111000000001 111111100000000
	16: {0xe7, 0xbc, 0x80, 0xe7, 0xb8, 0x81, 0xe5, 0x9a, 0xab, 0xe4, 0xb5, 0x92, 0xe5, 0xbf, 0xb5, 0xe8, 0xad, 0x94, 0xe8, 0x8f, 0xbc, 0xe4, 0xab, 0x95, 0xe8, 0x80, 0x80, 0x69},
	// 010101011010101 000001111111100 000101101010100 101111111110101 010110101010010 101011010101011 111111000000001 111111100000000
	15: {0xe7, 0xbc, 0x80, 0xe7, 0xb8, 0x81, 0xe5, 0x9a, 0xab, 0xe4, 0xb5, 0x92, 0xe5, 0xbf, 0xb5, 0xe8, 0xad, 0x94, 0xe8, 0x8f, 0xbc, 0xe4, 0xab, 0x95},
	// ...111111110101 010110101010010 101011010101011 111111000000001 111111100000000
	9: {0xe7, 0xbc, 0x80, 0xe7, 0xb8, 0x81, 0xe5, 0x9a, 0xab, 0xe4, 0xb5, 0x92, 0xe8, 0xbf, 0xb5, 0x6d},
	// ...........0101 010110101010010 101011010101011 111111000000001 111111100000000
	8: {0xe7, 0xbc, 0x80, 0xe7, 0xb8, 0x81, 0xe5, 0x9a, 0xab, 0xe4, 0xb5, 0x92, 0xe8, 0x80, 0x85, 0x65},
	// ....10101010010 101011010101011 111111000000001 111111100000000
	7: {0xe7, 0xbc, 0x80, 0xe7, 0xb8, 0x81, 0xe5, 0x9a, 0xab, 0xe8, 0x95, 0x92, 0x6c},
	// ............010 101011010101011 111111000000001 111111100000000
	6: {0xe7, 0xbc, 0x80, 0xe7, 0xb8, 0x81, 0xe5, 0x9a, 0xab, 0xe8, 0x80, 0x82, 0x64},
	// .....1010101011 111111000000001 111111100000000
	5: {0xe7, 0xbc, 0x80, 0xe7, 0xb8, 0x81, 0xe8, 0x8a, 0xab, 0x6b},
	// .............11 111111000000001 111111100000000
	4: {0xe7, 0xbc, 0x80, 0xe7, 0xb8, 0x81, 0xe8, 0x80, 0x83, 0x63},
	// ......000000001 111111100000000
	3: {0xe7, 0xbc, 0x80, 0xe8, 0x80, 0x81, 0x6a},
	// ..............1 111111100000000
	2: {0xe7, 0xbc, 0x80, 0xe8, 0x80, 0x81, 0x62},
	// .......00000000
	1: {0xe8, 0x80, 0x80, 0x69},
	0: {},
}

var encodeExpectedStrings = map[int]string{
	16: "缀縁嚫䵒念譔菼䫕耀i",
	15: "缀縁嚫䵒念譔菼䫕",
	9:  "缀縁嚫䵒迵m",
	8:  "缀縁嚫䵒者e",
	7:  "缀縁嚫蕒l",
	6:  "缀縁嚫耂d",
	5:  "缀縁芫k",
	4:  "缀縁考c",
	3:  "缀老j",
	2:  "缀老b",
	1:  "耀i",
	0:  "",
}

func TestEncode(t *testing.T) {
	for n, expected := range encodeExpectedBytes {
		t.Run(fmt.Sprintf("data_size_%d", n), func(t *testing.T) {
			encoded := Encode(srcData[:n])
			if len(encoded) < len(expected) {
				t.Error(fmt.Sprintf("[%d] Encoded too short: %d (need: %d)", n, len(encoded), len(expected)))
				return
			}
			if len(encoded) > len(expected) {
				t.Error(fmt.Sprintf("[%d] Encoded too long: %d (need: %d), extra: %x", n, len(encoded), len(expected), encoded[len(expected):]))
				return
			}
			for i, b := range expected {
				if encoded[i] != b {
					t.Error(fmt.Sprintf("[%d](%d) Expected: %0.8b (0x%0.2x) got: %0.8b (0x%0.2x)", n, i, b, b, encoded[i], encoded[i]))
				}
			}
		})
	}
}

func TestEncodeToString(t *testing.T) {
	for n, expectedString := range encodeExpectedStrings {
		t.Run(fmt.Sprintf("string_length_%d", n), func(t *testing.T) {
			encodedString := EncodeToString(srcData[:n])
			if encodedString != expectedString {
				t.Error(fmt.Sprintf("[%d] String '%s' doesn't match expected string '%s'\n", n, encodedString, expectedString))
			}
		})
	}
}

func TestDecode(t *testing.T) {
	for n, srcBytes := range encodeExpectedBytes {
		t.Run(fmt.Sprintf("data_size_%d", n), func(t *testing.T) {
			decoded, err := Decode(srcBytes)
			if err != nil {
				t.Error("Error in Decode:", err)
			}
			expected := srcData[:n]
			if len(decoded) < len(expected) {
				t.Error(fmt.Sprintf("[%d] Decoded too short: %d (need: %d)", n, len(decoded), len(expected)))
				return
			}
			if len(decoded) > len(expected) {
				t.Error(fmt.Sprintf("[%d] Decoded too long: %d (need: %d)", n, len(decoded), len(expected)))
				return
			}
			for i, b := range expected {
				if decoded[i] != b {
					t.Error(fmt.Sprintf("[%d](%d) Expected: %0.8b (%0.2x) got: %0.8b (%0.2x)", n, i, b, b, decoded[i], decoded[i]))
				}
			}
		})
	}
}

func TestDecodeFromString(t *testing.T) {
	for n, decodeSrcString := range encodeExpectedStrings {
		t.Run(fmt.Sprintf("string_length_%d", n), func(t *testing.T) {
			decoded, err := DecodeFromString(decodeSrcString)
			if err != nil {
				t.Error(fmt.Sprintf("[%d]Error while decoding: %s", n, err))
			}
			for i, b := range decoded {
				if b != srcData[i] {
					t.Error(fmt.Sprintf("[%d](%d) Expected 0x%0.2x, got 0x%0.2x\n", n, i, decoded, srcData[i]))
				}
			}
		})
	}
}

func TestGetRuneFromBytes(t *testing.T) {
	data := []byte{0xf0, 0xa5, 0x5a, 0xa5}
	// 11110000 10100101 01011010 10100101

	// 0010 0101 1111 0000 => 25f0
	// 0101 0010 1111 1000 => 52f8
	// 0010 1001 0111 1100 => 297c
	// 0101 0100 1011 1110 => 54be
	// 0010 1010 0101 1111 => 2a5f
	// 0101 0101 0010 1111 => 552f
	// 0110 1010 1001 0111 => 6a97
	// 0011 0101 0100 1011 => 354b

	i := uint(0)
	expectedIndices := []uint{1, 2, 2, 2, 2, 2, 2, 2}
	expectedBits := []uint{7, 0, 1, 2, 3, 4, 5, 6}
	expectedValues := []uint16{0x25f0, 0x52f8, 0x297c, 0x54be, 0x2a5f, 0x552f, 0x6a97, 0x354b}
	for _, b := range []uint{0, 1, 2, 3, 4, 5, 6, 7} {
		t.Run(fmt.Sprintf("bit_offset_%d", b), func(t *testing.T) {
			value, newIndex, newBit, err := getRuneFromBytes(data, i, uint(b))
			if err != nil {
				t.Error(fmt.Sprintf("[b=%d] err raised: %s", b, err))
			}
			if newIndex != expectedIndices[b] {
				t.Error(fmt.Sprintf("[b=%d] index incorrect, expected: %d, got: %d", b, expectedIndices[b], newIndex))
			}
			if newBit != expectedBits[b] {
				t.Error(fmt.Sprintf("[b=%d] bit index incorrect, expected: %d, got: %d", b, expectedBits[b], newBit))
			}
			if value != expectedValues[b] {
				t.Error(fmt.Sprintf("[b=%d] value incorrect, expected: 0x%0.2x, got: 0x%0.2x", b, expectedValues[b], value))
			}
		})
	}
}

func TestGetBytesFromRune(t *testing.T) {
	runes := []rune{0x25f0, 0x52f8, 0x297c, 0x54be, 0x2a5f, 0x552f, 0x6a97, 0x354b}
	expectedBytes := map[uint]struct {
		data      []byte
		remainder byte
		bit       uint
	}{
		// 010 0101 1111 0000
		0: {[]byte{0xf0}, 0x25, 7},
		// 0100 1011 1110 000.
		1: {[]byte{0xe0, 0x4b}, 0x00, 0},
		// 0 1001 0111 1100 00..
		2: {[]byte{0xc0, 0x97}, 0x00, 1},
		// 01 0010 1111 1000 0...
		3: {[]byte{0x80, 0x2f}, 0x01, 2},
		// 010 0101 1111 0000 ....
		4: {[]byte{0x00, 0x5f}, 0x02, 3},
		// 0100 1011 1110 000. ....
		5: {[]byte{0x00, 0xbe}, 0x04, 4},
		// 0 1001 0111 1100 00.. ....
		6: {[]byte{0x00, 0x7c}, 0x09, 5},
		// 01 0010 1111 1000 0... ....
		7: {[]byte{0x00, 0xf8}, 0x012, 6},
	}
	for b, expected := range expectedBytes {
		t.Run(fmt.Sprintf("bit_offset_%d", b), func(t *testing.T) {
			data, remainder, bit := getBytesFromRune(uint16(runes[0]), 0, b)
			if bit != expected.bit {
				t.Error(fmt.Sprintf("[%d] bit index incorrect, expected: %d, got: %d", b, expected.bit, bit))
			}
			if remainder != expected.remainder {
				t.Error(fmt.Sprintf("[%d] remainder incorrect, expected: 0x%0.2x, got: 0x%0.2x", b, expected.remainder, remainder))
			}
			if len(data) > len(expected.data) {
				t.Error(fmt.Sprintf("[%d] data too long: %d (need: %d)", b, len(data), len(expected.data)))
			}
			if len(data) < len(expected.data) {
				t.Error(fmt.Sprintf("[%d] data too short: %d (need: %d)", b, len(data), len(expected.data)))
			}
			for j, b := range data {
				if b != expected.data[j] {
					t.Error(fmt.Sprintf("[%d](%d) data incorrect, expected: 0x%0.2x, got: 0x%0.2x", b, j, expected.data[j], b))
				}
			}
		})
	}
}

func TestEncodeDecode(t *testing.T) {
	lengths := []int{100, 10000, 1000000, 100000000}
	for _, length := range lengths {
		if length > 10000 && testing.Short() {
			t.Skip("Skipping very long decode(encode()) tests")
		}
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			var dataBuf bytes.Buffer
			dataBuf.Grow(length)
			for i := int(0); i < length; i += 1 {
				dataBuf.WriteByte(byte(rand.Intn(256)))
			}
			expectedData := dataBuf.Bytes()
			data, err := Decode(Encode(expectedData))
			if err != nil {
				t.Error("Error while decoding:", err)
			}
			if len(data) > len(expectedData) {
				t.Error(fmt.Sprintf("[%d] data too long: %d (need: %d)", length, len(data), len(expectedData)))
			}
			if len(data) < len(expectedData) {
				t.Error(fmt.Sprintf("[%d] data too short: %d (need: %d)", length, len(data), len(expectedData)))
			}
			for i, b := range data {
				if b != expectedData[i] {
					t.Error(fmt.Sprintf("[%d](%d) data incorrect, expected: 0x%0.2x, got: 0x%0.2x", length, i, expectedData[i], b))
				}
			}
		})
	}
}
