// SPDX-License-Identifier: BSD-1-Clause
//
// Mapping data derived from Unicode, Inc. files at:
// https://www.unicode.org/Public/MAPPINGS/VENDORS/APPLE/
// See the NOTICE file in the root of this repository for license information.

package apple

import (
	"unicode/utf8"

	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform"
)

// singleByteDecoder provides the Transform implementation for single-byte decoders.
// It converts bytes from a legacy encoding to UTF-8 using a lookup table.
type singleByteDecoder struct {
	transform.NopResetter
	toUnicode *[256]rune
}

// Transform implements transform.Transformer.
func (d *singleByteDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for _, b := range src {
		r := d.toUnicode[b]

		// Calculate the number of bytes needed to encode this rune
		size := utf8.RuneLen(r)
		if size < 0 {
			r = utf8.RuneError
		}

		// Check if we have enough space in the destination buffer
		if nDst+size > len(dst) {
			err = transform.ErrShortDst
			return
		}

		// Encode the rune to UTF-8
		nDst += utf8.EncodeRune(dst[nDst:], r)
		nSrc++
	}
	return
}

// singleByteEncoder provides the Transform implementation for single-byte encoders.
// It converts UTF-8 to bytes in a legacy encoding using a reverse lookup map.
type singleByteEncoder struct {
	transform.NopResetter
	fromUnicode map[rune]byte
	toUnicode   *[256]rune // Used to check for valid ASCII range
}

// Transform implements transform.Transformer.
func (e *singleByteEncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		// Decode the next rune from UTF-8
		r, size := utf8.DecodeRune(src[nSrc:])

		// Handle incomplete rune at end of input
		if r == utf8.RuneError && size == 1 {
			if !atEOF && !utf8.FullRune(src[nSrc:]) {
				err = transform.ErrShortSrc
				return
			}
			// Invalid UTF-8, try to find a mapping or use replacement
		}

		// Check if we have space in the destination buffer
		if nDst >= len(dst) {
			err = transform.ErrShortDst
			return
		}

		// Try to find the byte mapping for this rune
		if b, ok := e.fromUnicode[r]; ok {
			dst[nDst] = b
			nDst++
			nSrc += size
			continue
		}

		// For ASCII range (0x00-0x7F), use direct mapping if the encoding
		// supports it (most Mac encodings do for the printable ASCII range)
		if r < 0x80 {
			// Check if the encoding has the same mapping for this byte
			if e.toUnicode[r] == r {
				dst[nDst] = byte(r)
				nDst++
				nSrc += size
				continue
			}
		}

		// Character cannot be encoded - return error
		// The caller can use transform.Chain with a replacement error handler
		err = &UnencodableRuneError{Rune: r, Encoding: "Mac OS encoding"}
		return
	}
	return
}

// UnencodableRuneError is returned when a Unicode rune cannot be encoded
// in the target encoding.
type UnencodableRuneError struct {
	Rune     rune
	Encoding string
}

func (e *UnencodableRuneError) Error() string {
	return "apple: cannot encode rune " + string(e.Rune) + " in " + e.Encoding
}

// makeDecoder creates a decoder for a single-byte encoding.
func makeDecoder(toUnicode *[256]rune) *singleByteDecoder {
	return &singleByteDecoder{toUnicode: toUnicode}
}

// makeEncoder creates an encoder for a single-byte encoding.
func makeEncoder(toUnicode *[256]rune, fromUnicode map[rune]byte) *singleByteEncoder {
	return &singleByteEncoder{
		toUnicode:   toUnicode,
		fromUnicode: fromUnicode,
	}
}
