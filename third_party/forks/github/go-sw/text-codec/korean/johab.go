// SPDX-License-Identifier: BSD-1-Clause
//
// Johab encoding implementation based on KS C 5601-1992 Annex 3.
// Algorithm derived from CPython cjkcodecs and Unicode mapping files.

package korean

import (
	"unicode/utf8"

	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform"
)

// Johab is the Korean Johab encoding (KS C 5601-1992 Annex 3).
var Johab encoding.Encoding = &johabEncoding{}

type johabEncoding struct{}

func (e *johabEncoding) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: &johabDecoder{}}
}

func (e *johabEncoding) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: &johabEncoder{}}
}

func (e *johabEncoding) String() string {
	return "Johab"
}

// Hangul syllable constants
const (
	hangulSBase  = 0xAC00 // First Hangul syllable in Unicode
	hangulLCount = 19     // Number of leading consonants
	hangulVCount = 21     // Number of vowels
	hangulTCount = 28     // Number of trailing consonants (including none)
	hangulNCount = hangulVCount * hangulTCount
	hangulSCount = hangulLCount * hangulNCount
)

// Decomposed Jamo ranges (U+1100-U+11FF)
const (
	jamoLBase = 0x1100 // Choseong (leading consonants) U+1100-U+1112
	jamoLEnd  = 0x1112
	jamoVBase = 0x1161 // Jungseong (vowels) U+1161-U+1175
	jamoVEnd  = 0x1175
	jamoTBase = 0x11A8 // Jongseong (trailing consonants) U+11A8-U+11C2
	jamoTEnd  = 0x11C2
)

// Sentinel values for lookup tables
const (
	_fill = 0xFD // Filler for optional components
	_none = 0xFF // Invalid index
)

// u2johabIdx maps Unicode Jamo L/V/T indices to Johab 5-bit indices.
// Index 0 in jongseong means "no final consonant" and maps to _FILL.
var (
	// Choseong (initial consonant) L index -> Johab index
	// Unicode L: ㄱㄲㄴㄷㄸㄹㅁㅂㅃㅅㅆㅇㅈㅉㅊㅋㅌㅍㅎ (19 values)
	u2johabChoseong = [19]byte{
		0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11,
		0x12, 0x13, 0x14,
	}

	// Jungseong (vowel) V index -> Johab index
	// Unicode V: ㅏㅐㅑㅒㅓㅔㅕㅖㅗㅘㅙㅚㅛㅜㅝㅞㅟㅠㅡㅢㅣ (21 values)
	u2johabJungseong = [21]byte{
		// Matches CPython's u2johabidx_jungseong.
		0x03, 0x04, 0x05, 0x06, 0x07,
		0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x1A, 0x1B, 0x1C, 0x1D,
	}

	// Jongseong (final consonant) T index -> Johab index
	// T=0 means no final consonant, uses Johab index 1 (filler position)
	// Unicode T: (none)ㄱㄲㄳㄴㄵㄶㄷㄹㄺㄻㄼㄽㄾㄿㅀㅁㅂㅄㅅㅆㅇㅈㅊㅋㅌㅍㅎ (28 values)
	u2johabJongseong = [28]byte{
		// Matches CPython's u2johabidx_jongseong (note the gap at 0x12).
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x10, 0x11, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D,
	}
)

// johab2uIdx maps Johab 5-bit indices back to Unicode Jamo L/V/T indices.
var (
	// Johab choseong index -> Unicode L index
	johab2uChoseong = [32]byte{
		_none, _fill, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
		14, 15, 16, 17, 18, _none, _none, _none, _none, _none, _none, _none, _none, _none, _none, _none,
	}

	// Johab jungseong index -> Unicode V index
	johab2uJungseong = [32]byte{
		// Matches CPython's johabidx_jungseong.
		_none, _none, _fill, 0, 1, 2, 3, 4,
		_none, _none, 5, 6, 7, 8, 9, 10,
		_none, _none, 11, 12, 13, 14, 15, 16,
		_none, _none, 17, 18, 19, 20, _none, _none,
	}

	// Johab jongseong index -> Unicode T index
	johab2uJongseong = [32]byte{
		// Matches CPython's johabidx_jongseong (note the unused 0x12 index).
		_none, _fill, 1, 2, 3, 4, 5, 6,
		7, 8, 9, 10, 11, 12, 13, 14,
		15, 16, _none, 17, 18, 19, 20, 21,
		22, 23, 24, 25, 26, 27, _none, _none,
	}
)

// decomposedToCompat maps decomposed Jamo (U+1100-U+11FF) to compatibility Jamo (U+3131-U+318E)
var (
	// Choseong (leading consonants) U+1100-U+1112 -> Compatibility Jamo
	choseongToCompat = [19]rune{
		0x3131, // U+1100 ㄱ -> U+3131
		0x3132, // U+1101 ㄲ -> U+3132
		0x3134, // U+1102 ㄴ -> U+3134
		0x3137, // U+1103 ㄷ -> U+3137
		0x3138, // U+1104 ㄸ -> U+3138
		0x3139, // U+1105 ㄹ -> U+3139
		0x3141, // U+1106 ㅁ -> U+3141
		0x3142, // U+1107 ㅂ -> U+3142
		0x3143, // U+1108 ㅃ -> U+3143
		0x3145, // U+1109 ㅅ -> U+3145
		0x3146, // U+110A ㅆ -> U+3146
		0x3147, // U+110B ㅇ -> U+3147
		0x3148, // U+110C ㅈ -> U+3148
		0x3149, // U+110D ㅉ -> U+3149
		0x314A, // U+110E ㅊ -> U+314A
		0x314B, // U+110F ㅋ -> U+314B
		0x314C, // U+1110 ㅌ -> U+314C
		0x314D, // U+1111 ㅍ -> U+314D
		0x314E, // U+1112 ㅎ -> U+314E
	}

	// Jungseong (vowels) U+1161-U+1175 -> Compatibility Jamo
	jungseongToCompat = [21]rune{
		0x314F, // U+1161 ㅏ -> U+314F
		0x3150, // U+1162 ㅐ -> U+3150
		0x3151, // U+1163 ㅑ -> U+3151
		0x3152, // U+1164 ㅒ -> U+3152
		0x3153, // U+1165 ㅓ -> U+3153
		0x3154, // U+1166 ㅔ -> U+3154
		0x3155, // U+1167 ㅕ -> U+3155
		0x3156, // U+1168 ㅖ -> U+3156
		0x3157, // U+1169 ㅗ -> U+3157
		0x3158, // U+116A ㅘ -> U+3158
		0x3159, // U+116B ㅙ -> U+3159
		0x315A, // U+116C ㅚ -> U+315A
		0x315B, // U+116D ㅛ -> U+315B
		0x315C, // U+116E ㅜ -> U+315C
		0x315D, // U+116F ㅝ -> U+315D
		0x315E, // U+1170 ㅞ -> U+315E
		0x315F, // U+1171 ㅟ -> U+315F
		0x3160, // U+1172 ㅠ -> U+3160
		0x3161, // U+1173 ㅡ -> U+3161
		0x3162, // U+1174 ㅢ -> U+3162
		0x3163, // U+1175 ㅣ -> U+3163
	}

	// Jongseong (trailing consonants) U+11A8-U+11C2 -> Compatibility Jamo
	jongseongToCompat = [27]rune{
		0x3131, // U+11A8 ㄱ -> U+3131
		0x3132, // U+11A9 ㄲ -> U+3132
		0x3133, // U+11AA ㄳ -> U+3133
		0x3134, // U+11AB ㄴ -> U+3134
		0x3135, // U+11AC ㄵ -> U+3135
		0x3136, // U+11AD ㄶ -> U+3136
		0x3137, // U+11AE ㄷ -> U+3137
		0x3139, // U+11AF ㄹ -> U+3139
		0x313A, // U+11B0 ㄺ -> U+313A
		0x313B, // U+11B1 ㄻ -> U+313B
		0x313C, // U+11B2 ㄼ -> U+313C
		0x313D, // U+11B3 ㄽ -> U+313D
		0x313E, // U+11B4 ㄾ -> U+313E
		0x313F, // U+11B5 ㄿ -> U+313F
		0x3140, // U+11B6 ㅀ -> U+3140
		0x3141, // U+11B7 ㅁ -> U+3141
		0x3142, // U+11B8 ㅂ -> U+3142
		0x3144, // U+11B9 ㅄ -> U+3144
		0x3145, // U+11BA ㅅ -> U+3145
		0x3146, // U+11BB ㅆ -> U+3146
		0x3147, // U+11BC ㅇ -> U+3147
		0x3148, // U+11BD ㅈ -> U+3148
		0x314A, // U+11BE ㅊ -> U+314A
		0x314B, // U+11BF ㅋ -> U+314B
		0x314C, // U+11C0 ㅌ -> U+314C
		0x314D, // U+11C1 ㅍ -> U+314D
		0x314E, // U+11C2 ㅎ -> U+314E
	}
)

// johabDecoder decodes Johab to UTF-8.
type johabDecoder struct {
	transform.NopResetter
}

func (d *johabDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		b1 := src[nSrc]

		// Single-byte ASCII range (0x00-0x7F)
		if b1 < 0x80 {
			if nDst >= len(dst) {
				err = transform.ErrShortDst
				return
			}
			dst[nDst] = b1
			nDst++
			nSrc++
			continue
		}

		// Need second byte for double-byte sequence
		if nSrc+1 >= len(src) {
			if atEOF {
				// Invalid: incomplete sequence at EOF
				err = transform.ErrShortSrc
				return
			}
			err = transform.ErrShortSrc
			return
		}

		b2 := src[nSrc+1]

		// Try to decode as Hangul syllable using bit manipulation
		r, ok := decodeJohabSyllable(b1, b2)
		if ok {
			size := utf8.RuneLen(r)
			if nDst+size > len(dst) {
				err = transform.ErrShortDst
				return
			}
			nDst += utf8.EncodeRune(dst[nDst:], r)
			nSrc += 2
			continue
		}

		// Try Hanja/symbol lookup tables
		if r, ok := lookupJohabToUnicode(b1, b2); ok {
			size := utf8.RuneLen(r)
			if nDst+size > len(dst) {
				err = transform.ErrShortDst
				return
			}
			nDst += utf8.EncodeRune(dst[nDst:], r)
			nSrc += 2
			continue
		}

		// Invalid sequence - output replacement character
		size := utf8.RuneLen(utf8.RuneError)
		if nDst+size > len(dst) {
			err = transform.ErrShortDst
			return
		}
		nDst += utf8.EncodeRune(dst[nDst:], utf8.RuneError)
		nSrc += 2
	}
	return
}

// decodeJohabSyllable decodes a Johab two-byte sequence to a Hangul syllable.
// Returns the Unicode code point and true if valid, or 0 and false if not a valid syllable.
func decodeJohabSyllable(b1, b2 byte) (rune, bool) {
	// Johab syllable encoding:
	// byte1: 1cccccjj (high bit set, 5-bit choseong, 2 high bits of jungseong)
	// byte2: jjjtttt (3 low bits of jungseong, 5-bit jongseong)

	if b1 < 0x84 || b1 > 0xD3 {
		return 0, false
	}

	// Extract 5-bit indices
	cho := (b1 >> 2) & 0x1F
	jung := ((b1 & 0x03) << 3) | ((b2 >> 5) & 0x07)
	jong := b2 & 0x1F

	// Validate and convert to Unicode indices
	if cho >= 32 || jung >= 32 || jong >= 32 {
		return 0, false
	}

	choU := johab2uChoseong[cho]
	jungU := johab2uJungseong[jung]
	jongU := johab2uJongseong[jong]

	if choU == _none || jungU == _none || jongU == _none {
		return 0, false
	}

	// Handle filler cases (individual Jamo, not full syllables)
	if choU == _fill || jungU == _fill {
		return 0, false // Not a composed syllable
	}

	// Convert jongseong filler to 0 (no final consonant)
	if jongU == _fill {
		jongU = 0
	}

	// Compose Unicode Hangul syllable
	// S = SBase + (L * VCount * TCount) + (V * TCount) + T
	r := rune(hangulSBase) + rune(choU)*hangulNCount + rune(jungU)*hangulTCount + rune(jongU)

	if r < hangulSBase || r > hangulSBase+hangulSCount-1 {
		return 0, false
	}

	return r, true
}

// johabEncoder encodes UTF-8 to Johab.
type johabEncoder struct {
	transform.NopResetter
}

func (e *johabEncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		r, size := utf8.DecodeRune(src[nSrc:])

		if r == utf8.RuneError && size == 1 {
			if !atEOF && !utf8.FullRune(src[nSrc:]) {
				err = transform.ErrShortSrc
				return
			}
		}

		// ASCII range
		if r < 0x80 {
			if nDst >= len(dst) {
				err = transform.ErrShortDst
				return
			}
			dst[nDst] = byte(r)
			nDst++
			nSrc += size
			continue
		}

		// Hangul syllable (U+AC00-U+D7A3)
		if r >= hangulSBase && r < hangulSBase+hangulSCount {
			if nDst+2 > len(dst) {
				err = transform.ErrShortDst
				return
			}

			b1, b2, ok := encodeJohabSyllable(r)
			if !ok {
				err = &UnencodableRuneError{r}
				return
			}

			dst[nDst] = b1
			dst[nDst+1] = b2
			nDst += 2
			nSrc += size
			continue
		}

		// Decomposed Jamo - Choseong (U+1100-U+1112)
		if r >= jamoLBase && r <= jamoLEnd {
			compat := choseongToCompat[r-jamoLBase]
			r = compat
		}

		// Decomposed Jamo - Jungseong (U+1161-U+1175)
		if r >= jamoVBase && r <= jamoVEnd {
			compat := jungseongToCompat[r-jamoVBase]
			r = compat
		}

		// Decomposed Jamo - Jongseong (U+11A8-U+11C2)
		if r >= jamoTBase && r <= jamoTEnd {
			compat := jongseongToCompat[r-jamoTBase]
			r = compat
		}

		// Non-Hangul characters via lookup table (also covers compatibility Jamo)
		if b1, b2, ok := lookupUnicodeToJohab(r); ok {
			if nDst+2 > len(dst) {
				err = transform.ErrShortDst
				return
			}
			dst[nDst] = b1
			dst[nDst+1] = b2
			nDst += 2
			nSrc += size
			continue
		}

		// Cannot encode this character
		err = &UnencodableRuneError{r}
		return
	}
	return
}

// encodeJohabSyllable encodes a Unicode Hangul syllable to Johab bytes.
func encodeJohabSyllable(r rune) (byte, byte, bool) {
	if r < hangulSBase || r >= hangulSBase+hangulSCount {
		return 0, 0, false
	}

	// Decompose Unicode syllable
	// S = SBase + (L * VCount * TCount) + (V * TCount) + T
	sIndex := r - hangulSBase
	l := sIndex / hangulNCount
	v := (sIndex % hangulNCount) / hangulTCount
	t := sIndex % hangulTCount

	if l >= hangulLCount || v >= hangulVCount || t >= hangulTCount {
		return 0, 0, false
	}

	// Convert to Johab indices
	cho := u2johabChoseong[l]
	jung := u2johabJungseong[v]
	jong := u2johabJongseong[t]

	// Encode to bytes
	// byte1: 1cccccjj (high bit set, 5-bit choseong, 2 high bits of jungseong)
	// byte2: jjjtttt (3 low bits of jungseong, 5-bit jongseong)
	b1 := byte(0x80) | (cho << 2) | (jung >> 3)
	b2 := ((jung & 0x07) << 5) | jong

	return b1, b2, true
}

// UnencodableRuneError is returned when a rune cannot be encoded.
type UnencodableRuneError struct {
	Rune rune
}

func (e *UnencodableRuneError) Error() string {
	return "korean: cannot encode rune " + string(e.Rune) + " in Johab encoding"
}
