// SPDX-License-Identifier: BSD-1-Clause
//
// Mapping data derived from Unicode, Inc. files at:
// https://www.unicode.org/Public/MAPPINGS/OBSOLETE/EASTASIA/KSC/
// See the NOTICE file in the root of this repository for license information.

// Package korean provides legacy Korean text encodings.
//
// This package implements the [github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding] interface
// for legacy Korean character encodings.
//
// # Supported Encodings
//
//   - [Johab] - KS C 5601-1992 Annex 3 (Johab encoding)
//
// # Johab Encoding
//
// Johab is a variable-width encoding for Korean that uses bit manipulation
// to encode Hangul syllables.
//
// The encoding structure:
//   - Single-byte: 0x00-0x7F (ASCII compatible, except 0x5C = Won sign)
//   - Double-byte: Hangul syllables, Jamo, Hanja, and symbols
//
// Jamo support includes:
//   - Precomposed syllables (U+AC00-U+D7A3) via bit computation
//   - Compatibility Jamo (U+3131-U+318E) via lookup table
//   - Decomposed Jamo (U+1100-U+11FF) mapped to compatibility equivalents
//
// For Hangul syllables (U+AC00-U+D7A3), the encoding uses:
//
//	byte1 = 0x80 | (cho << 2) | (jung >> 3)
//	byte2 = (jung << 5) | jong
//
// Where cho/jung/jong are 5-bit indices mapped from Unicode L/V/T components.
//
// # Usage
//
//	import "github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec/korean"
//
//	// Decode Johab to UTF-8
//	decoder := korean.Johab.NewDecoder()
//	utf8Bytes, err := decoder.Bytes(johabBytes)
//
//	// Encode UTF-8 to Johab
//	encoder := korean.Johab.NewEncoder()
//	johabBytes, err := encoder.Bytes(utf8Bytes)
//
// # Data Source
//
// The encoding tables are derived from:
//   - Unicode JOHAB.TXT: https://www.unicode.org/Public/MAPPINGS/OBSOLETE/EASTASIA/KSC/JOHAB.TXT
//   - CPython cjkcodecs: https://github.com/python/cpython/blob/main/Modules/cjkcodecs/_codecs_kr.c
//   - ICU charset data: https://github.com/unicode-org/icu-data
package korean
