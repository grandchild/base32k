/* CC0 - free software.
To the extent possible under law, all copyright and related or neighboring
rights to this work are waived. See the LICENSE file for more information. */

// base32k is a binary-to-text encoding with a better encoding ratio in
// character-limited situations such as twitter.
//
// base32k is a slightly whimsical binary-to-text encoding, which transforms
// raw binary data into (possibly obscene combinations of) UTF-8 characters
// from the CJK and Hangul unicode blocks. Its alphabet consists of 2^15 =
// 32768 characters, hence the name. In comparison to other encodings like
// base64 or base122, this does not save space in terms of bytes, but it is
// smaller than those two in terms of characters. This is only useful when for
// some reason the medium (*cough*twitter*cough*) is character-limited rather
// than byte-limited.
//
// base32k has an encoding ratio of 15 bits per unicode glyph, which amounts
// to a ratio of 15/24 (0.625) plus one byte padding in 14 out of 15 cases.
// This makes it a worse encoding than base64, which has a 3/4 (0.75) and
// base122, which has 7/8 (0.875).
//
// A twitter message may be 280 characters long, but only 140 CJK glyphs.
// Still, this encoding slightly outperforms base64 and even base122 in the
// space available in a single tweet:
//
//          space ratio   char ratio   bytes per tweet
//  base64     0.75           6           210
//  base122    0.875          7           245
//  base32k    0.625         15           256
//           ( more is better for all columns )
//
// base32k outperforming base122 on twitter results from the fact that twitter
// counts a CJK or Hangul glyph as two characters, whereas in utf8 it's
// actually 3 characters. This gives us, in effect, an encoding ratio of 15/16
// over base122's 7/8, a slight advantage.
//
// So, given good-enough font coverage of the basic multilingual unicode
// plane, this can be used to transmit data in situations where characters are
// limited, rather than disk space.
//
// This implementation will run out of memory when en-/decoding very large
// chunks of data (several gigabytes). But since this is aimed at character-
// limited settings this is not likely to be an issue.
package base32k

import (
	"bytes"
	"errors"
	"fmt"
)

// Code-Point-ranges ("lanes")
//
//    4000 - 9FFF  (CJK Scripts)
//    0100000000000000 -
//    1001111111111111
// AND
//    B000 - CFFF  (Hangul Scripts)
//    1011000000000000 -
//    1100111111111111
//
// -> 01xxxxxxxxxxxxxx [0]
// -> 100xxxxxxxxxxxxx [1]
// -> 1011xxxxxxxxxxxx [2]
// -> 1100xxxxxxxxxxxx [3]
// => 15 bits per glyph
//
// Note: The only way to improve on this without resorting to eldritch hacks
// such as half bits, would be to use all 16 bits, which would span the whole
// BMP, which contains unprintable characters in the very beginning and
// "private use" blocks at the end. All other planes beyond the BMP are
// encoded using 4 bytes, not to mention that they aren't nearly completely
// assigned.

// UTF-8 bit layout:
// UTF-8 encoding bits are represented differently: ","=1 and "."=0. Lanes are
// already shuffled from the order above to make encoding more straightforward
// (the final 3 bits of the first byte almost completely match the first three
// data bits, see `toLane` for the exception)
//   [1]   ,,,.10 0x  ,.xxxxxx  ,.xxxxxx
//   [2]   ,,,.10 11  ,.xxxxxx  ,.xxxxxx
//   [3]   ,,,.11 00  ,.xxxxxx  ,.xxxxxx
//   [0]   ,,,.01 xx  ,.xxxxxx  ,.xxxxxx

const BITS_PER_RUNE = 15
const BYTES_PER_RUNE = 15
const BYTE_LEN = 8
const PAD_START_SYMBOL = rune('a') // 0x61

var toLane = [...]uint16{ // {3 MSBs -> prefix}
	/*0b000:*/ 0x8000, // 1.000 [0]
	/*0b001:*/ 0x9000, // 1.001 [0]
	/*0b010:*/ 0x4000, // 0.100 [0] -> act as if .010
	/*0b011:*/ 0xb000, // 1.011 [3]
	/*0b100:*/ 0xc000, // 1.100 [4]
	/*0b101:*/ 0x5000, // 0.101 [0]
	/*0b110:*/ 0x6000, // 0.110 [0]
	/*0b111:*/ 0x7000, // 0.111 [0]
	/* pad: */ 0xf000, // 1.111 [0]
}
var fromLane = [...]byte{
	/*0x0:*/ 0xfe, // padding
	/*0x1:*/ 0xff, // invalid
	/*0x2:*/ 0xff, // invalid
	/*0x3:*/ 0xff, // invalid
	/*0x4:*/ 2, //    0.100 -> .010
	/*0x5:*/ 5, //    0.101
	/*0x6:*/ 6, //    0.110
	/*0x7:*/ 7, //    0.111
	/*0x8:*/ 0, //    1.000
	/*0x9:*/ 1, //    1.001
	/*0xa:*/ 0xff, // invalid
	/*0xb:*/ 3, //    1.011
	/*0xc:*/ 4, //    1.100
	/*0xd:*/ 0xff, // invalid
	/*0xe:*/ 0xff, // invalid
	/*0xf:*/ 0xff, // invalid
}

// Encode encodes a given byte array of data into a base32k byte array.
func Encode(src []byte) (dest []byte) { return encode(src) }

// Decode decodes a given base32k byte array back into a binary data byte
// array.
func Decode(src []byte) (dest []byte, err error) { return decode(src) }

// EncodeToString encodes a given byte array of data into a base32k string.
func EncodeToString(src []byte) (dest string) { return string(encode(src)) }

// DecodeFromString decodes a given base32k string back into a binary data
// byte array.
func DecodeFromString(s string) (dest []byte, err error) { return decode([]byte(s)) }

func encode(src []byte) (dest []byte) {
	if len(src) == 0 {
		return
	}
	var destBuf bytes.Buffer
	destBuf.Grow(EncodedLength(len(src)))
	r, i, b, d := uint16(0), uint(0), uint(0), uint(0)
	var err error
	for {
		r, i, b, err = getRuneFromBytes(src, i, b)
		if err != nil {
			break
		}
		prefix := toLane[r>>12]
		r = r&0x0fff | prefix
		destBuf.WriteRune(rune(r))
	}
	r, d, err = getLastRune(src, i, b)
	if err == nil {
		prefix := toLane[r>>12]
		r = r&0x0fff | prefix
		destBuf.WriteRune(rune(r))
		if d > 0 {
			destBuf.WriteRune(PAD_START_SYMBOL + rune(d))
		}
	}
	return destBuf.Bytes()
}

func getRuneFromBytes(src []byte, index uint, bit uint) (value uint16, newIndex uint, newBit uint, err error) {
	if index+2 == uint(len(src)) && bit > 1 ||
		index+2 > uint(len(src)) {
		return 0, index, bit, errors.New("End of input")
	}
	value = uint16(src[index] >> bit)
	value += uint16(src[index+1]) << (BYTE_LEN - bit)
	if bit > 1 { // we skipped more than 1 bit of the first byte & thus need some of the third byte as well
		value += uint16(src[index+2]) << (BYTE_LEN*2 - bit)
	}
	value &= 0x7fff
	if bit != 0 {
		newIndex = index + 2
	} else {
		newIndex = index + 1
	}
	newBit = (bit + BITS_PER_RUNE) % BYTE_LEN
	return
}

func getLastRune(src []byte, index uint, bit uint) (value uint16, digits uint, err error) {
	switch uint(len(src)) - index {
	case 2:
		value = uint16(src[index] >> bit)
		value += uint16(src[index+1]) << (BYTE_LEN - bit)
		digits = BITS_PER_RUNE + 1 - bit
	case 1:
		value = uint16(src[index] >> bit)
		digits = BITS_PER_RUNE + 1 - BYTE_LEN - bit
	default:
		err = errors.New("End of input reached")
	}
	return
}

func decode(src []byte) (data []byte, err error) {
	if len(src) == 0 {
		return
	}
	runes := bytes.Runes(src)
	var destBuf bytes.Buffer
	destBuf.Grow(DecodedLength(len(src), src[len(src)-1]))
	data, remainder, b := []byte{}, byte(0), uint(0)
	for i, r := range runes {
		prefix := fromLane[r>>12]
		if prefix == 0xff {
			return []byte{}, errors.New(fmt.Sprintf(
				"Invalid character at position %d: %s", i, string(r),
			))
		} else if prefix == 0xfe {
			if r <= PAD_START_SYMBOL && r >= (PAD_START_SYMBOL+BITS_PER_RUNE) || i != len(runes)-1 {
				return []byte{}, errors.New(fmt.Sprintf(
					"Invalid character or misplaced padding character at position %d: %s", i, string(r),
				))
			}
			padding := BITS_PER_RUNE - (r - PAD_START_SYMBOL)
			if padding >= 8 {
				destBuf.Truncate(destBuf.Len() - 1)
			}
			break
		}
		value := uint16(r)&0x0fff + uint16(prefix)<<12
		data, remainder, b = getBytesFromRune(value, remainder, b)
		destBuf.Write(data)
	}
	return destBuf.Bytes(), nil
}

func getBytesFromRune(value uint16, remainder byte, bit uint) (data []byte, newRemainder byte, newBit uint) {
	data = []byte{}
	data = append(data, byte(value<<bit)+remainder)
	if bit != 0 { // a complete second byte is available
		data = append(data, byte(value>>(BYTE_LEN-bit)))
		newRemainder = byte(value >> (BYTE_LEN*2 - bit))
	} else {
		newRemainder = byte(value >> BYTE_LEN)
	}
	newBit = (bit + BITS_PER_RUNE) % BYTE_LEN
	return
}

// EncodedLength returns the length of the encoded string in characters. It
// does an integer ceiling(!) division of the bit-length of src.
// See: Warren Jr., Henry S. "Hacker's Delight" Pearson 2003 (14th printing
// 2011) p. 139
func EncodedLength(srcLength int) (length int) {
	rawLength := (srcLength*BYTE_LEN + BITS_PER_RUNE - 1) / BITS_PER_RUNE
	padded := srcLength%BITS_PER_RUNE != 0
	if padded {
		return rawLength + 1
	} else {
		return rawLength
	}
}

// DecodedLength returns the length of the data in bytes resulting from
// decoding the source string.
func DecodedLength(srcLength int, paddingRune byte) (length int) {
	if srcLength == 0 {
		return 0
	}
	padded := srcLength%BYTES_PER_RUNE != 0
	var rawLength, padding int
	if padded {
		padding = BITS_PER_RUNE - int(rune(paddingRune)-PAD_START_SYMBOL)
		rawLength = srcLength - 1
	} else {
		padding = 0
		rawLength = srcLength
	}
	return (rawLength*BITS_PER_RUNE + 1 - BITS_PER_RUNE - padding) / BYTE_LEN
}
