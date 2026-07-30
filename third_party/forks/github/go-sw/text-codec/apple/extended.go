// SPDX-License-Identifier: BSD-1-Clause
//
// Mapping data derived from Unicode, Inc. files at:
// https://www.unicode.org/Public/MAPPINGS/VENDORS/APPLE/
// See the NOTICE file in the root of this repository for license information.
//
// Overrides data and logics derived from fonttools (MIT) at:
// https://github.com/fonttools/fonttools
// See the NOTICE file in the root for the MIT permission notice.

package apple

import (
	"unicode/utf8"

	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding/japanese"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding/korean"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding/simplifiedchinese"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding/traditionalchinese"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform"
)

// Extended encodings that wrap base encodings with custom byte overrides.
// These follow the pattern used by fonttools for Mac OS CJK encodings.

// ByteOverride describes a single-byte override.
//
// If Immediate is true, the override is applied whenever the byte appears at a
// character boundary.
// If Immediate is false, the override is only applied as a fallback when the
// base decoder cannot decode the byte/sequence (e.g., undefined bytes or a
// lead byte left incomplete at EOF).
type ByteOverride struct {
	Rune      rune
	Immediate bool
	// PreferBase controls encoding behavior.
	//
	// When true, the encoder will try to encode Rune using the base encoding
	// first, and only fall back to emitting the single-byte override if the base
	// encoding cannot represent it. This can improve interoperability with
	// non-extended decoders.
	PreferBase bool
}

// NewExtendedEncoding creates an encoding that wraps a base encoding with
// single-byte overrides.
//
// This is useful for Apple-style legacy encodings and other edge cases where
// specific bytes should decode to a different rune than the base encoding.
func NewExtendedEncoding(name string, base encoding.Encoding, overrides map[byte]ByteOverride) encoding.Encoding {
	// Copy the map so callers can't mutate behavior after construction.
	copyOverrides := make(map[byte]ByteOverride, len(overrides))
	for b, o := range overrides {
		copyOverrides[b] = o
	}
	return &extendedEncoding{
		name:      name,
		base:      base,
		overrides: copyOverrides,
	}
}

// Japanese is the Mac OS Japanese encoding.
// It extends Shift-JIS with Apple-specific single-byte overrides.
var Japanese encoding.Encoding = &extendedEncoding{
	name: "MacJapanese",
	base: japanese.ShiftJIS,
	// Overrides from fonttools x_mac_japanese_ttx
	// Shift-JIS lead bytes: 0x81-0x9F, 0xE0-0xFC
	overrides: map[byte]ByteOverride{
		0x5C: {Rune: 0x00A5, Immediate: true}, // YEN SIGN (instead of REVERSE SOLIDUS)
		0x7E: {Rune: 0x007E, Immediate: true}, // TILDE (instead of OVERLINE)
		0x80: {Rune: 0x005C, Immediate: true}, // REVERSE SOLIDUS
		0xA0: {Rune: 0x00A0, Immediate: true}, // NO-BREAK SPACE
		// Note: for the Mac TTX encoding we also treat 0xFC as immediate even though
		// it is a Shift-JIS lead byte. This ensures 0xFC 0x7E decodes as "|~".
		0xFC: {Rune: 0x007C, Immediate: true},                   // VERTICAL LINE
		0xFD: {Rune: 0x00A9, Immediate: true},                   // COPYRIGHT SIGN
		0xFE: {Rune: 0x2122, Immediate: true},                   // TRADE MARK SIGN
		0xFF: {Rune: 0x2026, Immediate: true, PreferBase: true}, // HORIZONTAL ELLIPSIS
	},
}

// Korean is the Mac OS Korean encoding.
// It extends EUC-KR with Apple-specific single-byte overrides.
var Korean encoding.Encoding = &extendedEncoding{
	name: "MacKorean",
	base: korean.EUCKR,
	// Overrides from fonttools x_mac_korean_ttx
	// EUC-KR lead bytes: 0x81-0xFE
	overrides: map[byte]ByteOverride{
		0x80: {Rune: 0x00A0, Immediate: true},                   // NO-BREAK SPACE
		0x81: {Rune: 0x20A9, Immediate: true},                   // WON SIGN
		0x82: {Rune: 0x2014, Immediate: true},                   // EM DASH
		0x83: {Rune: 0x00A9, Immediate: true},                   // COPYRIGHT SIGN
		0xFE: {Rune: 0x2122, Immediate: true},                   // TRADE MARK SIGN
		0xFF: {Rune: 0x2026, Immediate: true, PreferBase: true}, // HORIZONTAL ELLIPSIS
	},
}

// ChineseSimplified is the Mac OS Simplified Chinese encoding.
// It extends GBK with Apple-specific single-byte overrides.
var ChineseSimplified encoding.Encoding = &extendedEncoding{
	name: "MacChineseSimplified",
	base: simplifiedchinese.GBK,
	// Overrides from fonttools x_mac_simp_chinese_ttx
	// GBK lead bytes: 0x81-0xFE
	// Note: 0x80 is Euro sign in GBK/CP936, but Apple maps to ü
	overrides: map[byte]ByteOverride{
		0x80: {Rune: 0x00FC, Immediate: true},                   // LATIN SMALL LETTER U WITH DIAERESIS (GBK has Euro here)
		0xA0: {Rune: 0x00A0, Immediate: true},                   // NO-BREAK SPACE
		0xFD: {Rune: 0x00A9, Immediate: true},                   // COPYRIGHT SIGN
		0xFE: {Rune: 0x2122, Immediate: true},                   // TRADE MARK SIGN
		0xFF: {Rune: 0x2026, Immediate: true, PreferBase: true}, // HORIZONTAL ELLIPSIS
	},
}

// ChineseTraditional is the Mac OS Traditional Chinese encoding.
// It extends Big5 with Apple-specific single-byte overrides.
var ChineseTraditional encoding.Encoding = &extendedEncoding{
	name: "MacChineseTraditional",
	base: traditionalchinese.Big5,
	// Overrides from fonttools x_mac_trad_chinese_ttx
	// Big5 lead bytes: 0x81-0xFE
	overrides: map[byte]ByteOverride{
		0x80: {Rune: 0x005C, Immediate: true},                   // REVERSE SOLIDUS
		0xA0: {Rune: 0x00A0, Immediate: true},                   // NO-BREAK SPACE
		0xFD: {Rune: 0x00A9, Immediate: true},                   // COPYRIGHT SIGN
		0xFE: {Rune: 0x2122, Immediate: true},                   // TRADE MARK SIGN
		0xFF: {Rune: 0x2026, Immediate: true, PreferBase: true}, // HORIZONTAL ELLIPSIS
	},
}

// extendedEncoding wraps a base encoding with single-byte overrides.
type extendedEncoding struct {
	name string
	base encoding.Encoding
	// overrides maps a byte to its override rune and the policy for when to apply it.
	// If immediate is true, the override is applied whenever the byte appears at a
	// character boundary. If immediate is false, the override is only applied as a
	// fallback when the base decoder cannot decode the byte/sequence.
	overrides map[byte]ByteOverride
}

func (e *extendedEncoding) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{
		Transformer: &extendedDecoder{
			base:      e.base.NewDecoder().Transformer,
			overrides: e.overrides,
		},
	}
}

func (e *extendedEncoding) NewEncoder() *encoding.Encoder {
	// Build reverse map for encoding
	encodeOverrides := make(map[rune]encodeOverride)
	for b, o := range e.overrides {
		encodeOverrides[o.Rune] = encodeOverride{b: b, preferBase: o.PreferBase}
	}
	return &encoding.Encoder{
		Transformer: &extendedEncoder{
			base:      e.base.NewEncoder().Transformer,
			overrides: encodeOverrides,
		},
	}
}

func (e *extendedEncoding) String() string {
	return e.name
}

// extendedDecoder handles decoding with single-byte overrides.
// It follows the fonttools approach: overrides are only applied when the
// base decoder cannot handle a byte. This ensures multi-byte sequences
// (like Shift-JIS 0x8C 0xA0 for 権) are decoded correctly even when a
// trail byte happens to match an override.
type extendedDecoder struct {
	base      transform.Transformer
	overrides map[byte]ByteOverride
}

func (d *extendedDecoder) Reset() {
	d.base.Reset()
}

func (d *extendedDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		b := src[nSrc]

		// Check if this byte should be immediately overridden.
		// Immediate overrides are bytes that this encoding treats as standalone
		// single-byte codes when they appear at a character boundary.
		//
		// Even if a byte is a lead byte in the *base* encoding, we may still mark
		// it immediate for TTX-style semantics. Trail bytes are protected because
		// we never split a potential 2-byte sequence; see the nSrc+2 scan below.
		if o, ok := d.overrides[b]; ok {
			if o.Immediate {
				r := o.Rune
				size := utf8.RuneLen(r)
				if nDst+size > len(dst) {
					err = transform.ErrShortDst
					return
				}
				nDst += utf8.EncodeRune(dst[nDst:], r)
				nSrc++
				continue
			}
		}

		// For ASCII bytes without override, decode just this one byte.
		// This prevents the base decoder from greedily consuming following
		// bytes that might have overrides.
		if b < 0x80 {
			if nDst >= len(dst) {
				err = transform.ErrShortDst
				return
			}
			dst[nDst] = b
			nDst++
			nSrc++
			continue
		}

		// High byte that's either a lead byte or has no override.
		// Let base decoder handle it. For lead bytes with overrides (like 0xFC
		// in Shift-JIS), the override only applies when the base decoder
		// produces a replacement character (indicating invalid/incomplete sequence).
		//
		// Limit input to base decoder: find where the next immediate override byte is
		// so that the base decoder doesn't consume override bytes that should be
		// processed separately. For CJK encodings, multi-byte sequences are at most
		// 2 bytes, so start scanning at nSrc+2 (after potential trail byte).
		end := len(src)
		for i := nSrc + 2; i < len(src); i++ {
			if o, ok := d.overrides[src[i]]; ok && o.Immediate {
				// Found an immediate override byte - stop here
				end = i
				break
			}
		}

		segmentAtEOF := atEOF && end == len(src)
		startDst := nDst
		nd, ns, e := d.base.Transform(dst[nDst:], src[nSrc:end], segmentAtEOF)

		// Check if base decoder produced a replacement character for a single
		// byte that has an override. This happens when a lead byte is alone
		// at EOF - the decoder produces U+FFFD instead of ErrShortSrc.
		if ns == 1 && nd == 3 && e == nil {
			if o, ok := d.overrides[src[nSrc]]; ok {
				// Check if output is replacement character (EF BF BD)
				if dst[startDst] == 0xEF && dst[startDst+1] == 0xBF && dst[startDst+2] == 0xBD {
					r := o.Rune
					// Replace with our override
					size := utf8.RuneLen(r)
					if startDst+size > len(dst) {
						err = transform.ErrShortDst
						return
					}
					utf8.EncodeRune(dst[startDst:], r)
					nDst = startDst + size
					nSrc++
					continue
				}
			}
		}

		nDst += nd
		nSrc += ns

		if ns > 0 {
			if e == transform.ErrShortDst {
				err = e
				return
			}
			continue
		}

		// Base decoder made no progress
		if e == transform.ErrShortDst {
			err = e
			return
		}

		if e == transform.ErrShortSrc {
			if !atEOF {
				// Need more input - caller should buffer and retry
				err = e
				return
			}
			// At EOF with incomplete sequence - check override for lead byte
			if o, ok := d.overrides[b]; ok {
				r := o.Rune
				size := utf8.RuneLen(r)
				if nDst+size > len(dst) {
					err = transform.ErrShortDst
					return
				}
				nDst += utf8.EncodeRune(dst[nDst:], r)
				nSrc++
				continue
			}
			err = e
			return
		}

		// Other error or nil with no progress - shouldn't happen normally
		if e != nil {
			err = e
			return
		}

		if atEOF {
			break
		}

		err = transform.ErrShortSrc
		return
	}
	return
}

// extendedEncoder handles encoding with single-byte overrides.
type extendedEncoder struct {
	base      transform.Transformer
	overrides map[rune]encodeOverride
}

type encodeOverride struct {
	b          byte
	preferBase bool
}

func (e *extendedEncoder) Reset() {
	e.base.Reset()
}

func (e *extendedEncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		// Decode the next rune from UTF-8
		r, size := utf8.DecodeRune(src[nSrc:])

		// Handle incomplete rune at end of input
		if r == utf8.RuneError && size == 1 {
			if !atEOF && !utf8.FullRune(src[nSrc:]) {
				err = transform.ErrShortSrc
				return
			}
		}

		// Check if this rune has an override
		if o, ok := e.overrides[r]; ok {
			if o.preferBase {
				nd, ns, err2 := e.base.Transform(dst[nDst:], src[nSrc:nSrc+size], true)
				if err2 == nil && ns > 0 {
					nDst += nd
					nSrc += ns
					continue
				}
				if err2 == transform.ErrShortDst {
					err = err2
					return
				}
				// Fall back to single-byte override.
			}

			if nDst >= len(dst) {
				err = transform.ErrShortDst
				return
			}
			dst[nDst] = o.b
			nDst++
			nSrc += size
			continue
		}

		// For non-override runes, encode only THIS rune with base transformer.
		// We must not pass the entire remaining input because subsequent runes
		// may have overrides that the base encoder doesn't support (e.g., ©, ™).
		// Use atEOF=true since we have a complete rune.
		nd, ns, err2 := e.base.Transform(dst[nDst:], src[nSrc:nSrc+size], true)
		nDst += nd
		nSrc += ns

		if err2 != nil {
			err = err2
			return
		}

		// If no progress was made, the rune couldn't be encoded
		if ns == 0 {
			// This shouldn't normally happen if override map is complete
			err = transform.ErrShortDst
			return
		}
	}
	return
}
