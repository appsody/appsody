package isbinary

import (
	"io"
)

// BlockSize is the amount of data that will be considered when testing for
// binary data.
const BlockSize = 512

// Test will return whether or not the data contained in the first BlockSize
// bytes of the input buffer is binary data.  It will attempt to properly
// handle UTF-8 encoded data - see the source for more information.
func Test(buf []byte) bool {
	if len(buf) == 0 {
		return false
	}
	if len(buf) > BlockSize {
		buf = buf[:BlockSize]
	}

	// Check for UTF-8 byte-order mark
	if len(buf) >= 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
		return false
	}

	var suspiciousBytes int
	for i := 0; i < len(buf); i++ {
		// A null char means that it's binary
		if buf[i] == 0x00 {
			return true
		}

		// Check for non-printable characters
		if (buf[i] < 7 || buf[i] > 14) && (buf[i] < 32 || buf[i] > 127) {
			// Check if this is UTF-8.  The UTF-8 encoding is:
			//
			//  |    Bits of Code Point     |    First Code Point    |    Last Code Point    |
			//  +---------------------------+------------------------+-----------------------+
			//  |            7              |        U+0000          |        U+007F         |
			//  |           11              |        U+0080          |        U+07FF         |
			//  |           16              |        U+0800          |        U+FFFF         |
			//  |           21              |        U+10000         |        U+1FFFFF       |
			//
			// And the corresponding byte patterns are:
			//
			//  |    Bits of Code Point     |  Byte 1  |  Byte 2  |  Byte 3  |  Byte 4  |
			//  +---------------------------+----------+----------+----------+----------+
			//  |            7              | 0xxxxxxx |          |          |          |
			//  |           11              | 110xxxxx | 10xxxxxx |          |          |
			//  |           16              | 1110xxxx | 10xxxxxx | 10xxxxxx |          |
			//  |           21              | 11110xxx | 10xxxxxx | 10xxxxxx | 10xxxxxx |
			//
			// What we do is check if the current byte matches the first code point of a
			// UTF-8 string (currently, only the two or three-byte version), and then
			// verify the following byte(s).
			//
			// We can verify how these characters encode by encoding the upper and lower
			// bounds of each case, and observing the decimal values of the corresponding
			// bytes.  A quick Python function can be used to verify (run on Python 3):
			//
			//     def to_binary(s):
			//       chars = s.encode('utf-8')
			//       for ch in chars:
			//         print(bin(ch)[2:] + ' ', end='')
			//       print('')
			//
			// For the two-byte case, we get:
			//
			//     >>> to_binary('\u0080')
			//     11000010 10000000
			//     >>> to_binary('\u07FF')
			//     11011111 10111111
			//
			// This results in a range of [194, 223] for the first byte, and [128, 191]
			// for the second.
			//
			// Similarly, for the three-byte case, we get:
			//     >>> to_binary('\u0800')
			//     11100000 10100000 10000000
			//     >>> to_binary('\uFFFF')
			//     11101111 10111111 10111111
			//
			// Or, a range of [224, 239] for the first byte, and the same range as above
			// for the second and third bytes.
			//
			// The above dictates the logic that we use here.
			if buf[i] >= 194 && buf[i] <= 223 && (i+1) < len(buf) {
				i++
				if buf[i] >= 128 && buf[i] <= 191 {
					continue
				}
			} else if buf[i] >= 224 && buf[i] <= 239 && (i+2) < len(buf) {
				i++
				if buf[i] >= 128 && buf[i] <= 191 && buf[i+1] >= 128 && buf[i+1] <= 191 {
					i++
					continue
				}
			}

			// TODO(andrew-d): We should probably add in the third case, now that emoji
			// is a thing...
			suspiciousBytes++
		}
	}

	// If the percentage of suspicious bytes is larger than 10%, we treat this as binary.
	if ((suspiciousBytes * 100) / len(buf)) > 10 {
		return true
	}

	return false
}

// TestReader performs the same checks as Test, but will read the data to test
// from the provided io.Reader.
func TestReader(r io.Reader) (bool, error) {
	buf := make([]byte, BlockSize)
	n, err := r.Read(buf)
	if err != nil {
		return false, err
	}

	return Test(buf[:n]), nil
}
