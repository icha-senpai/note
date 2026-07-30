// SPDX-License-Identifier: BSD-1-Clause
//
// Mapping data derived from Unicode, Inc. files at:
// https://www.unicode.org/Public/MAPPINGS/VENDORS/APPLE/
// See the NOTICE file in the root of this repository for license information.

// Package apple provides legacy Mac OS text encodings.
//
// This package implements the [github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding] interface
// for various legacy Mac OS character encodings. These encodings were historically
// used on Macintosh computers before the transition to Unicode.
//
// # Standard Library Encodings
//
// Note: Mac OS Roman and Mac OS Cyrillic are already provided by the Go standard
// library in [github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding/charmap]:
//   - [charmap.Macintosh] - Mac OS Roman (Western European)
//   - [charmap.MacintoshCyrillic] - Mac OS Cyrillic
//
// This package provides the remaining Mac OS encodings not covered by the standard library.
//
// # Supported Encodings
//
// European:
//   - [CentralEuropean] - Mac OS Central European
//   - [Croatian] - Mac OS Croatian
//   - [Greek] - Mac OS Greek
//   - [Iceland] - Mac OS Icelandic
//   - [Romanian] - Mac OS Romanian
//   - [Turkish] - Mac OS Turkish
//   - [Celtic] - Mac OS Celtic
//   - [Gaelic] - Mac OS Gaelic
//   - [Ukraine] - Mac OS Ukrainian
//
// Middle Eastern:
//   - [Arabic] - Mac OS Arabic
//   - [Farsi] - Mac OS Farsi/Persian
//   - [Hebrew] - Mac OS Hebrew
//
// Indic:
//   - [Devanagari] - Mac OS Devanagari
//   - [Gujarati] - Mac OS Gujarati
//   - [Gurmukhi] - Mac OS Gurmukhi
//
// East Asian (CJK):
//   - [Japanese] - Mac OS Japanese (extends Shift-JIS)
//   - [Korean] - Mac OS Korean (extends EUC-KR)
//   - [ChineseSimplified] - Mac OS Simplified Chinese (extends GBK)
//   - [ChineseTraditional] - Mac OS Traditional Chinese (extends Big5)
//
// Other:
//   - [Thai] - Mac OS Thai
//   - [Inuit] - Mac OS Inuit
//   - [Dingbats] - Mac OS Dingbats
//   - [Symbol] - Mac OS Symbol
//   - [Keyboard] - Mac OS Keyboard glyphs
//
// # Usage
//
// Each encoding provides NewDecoder() and NewEncoder() methods:
//
//	import "github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec/apple"
//
//	// Decode Mac Central European to UTF-8
//	decoder := apple.CentralEuropean.NewDecoder()
//	utf8Bytes, err := decoder.Bytes(macCentEuroBytes)
//
//	// Encode UTF-8 to Mac Central European
//	encoder := apple.CentralEuropean.NewEncoder()
//	macCentEuroBytes, err := encoder.Bytes(utf8Bytes)
//
// For streaming, use with [github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform]:
//
//	import "github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform"
//
//	reader := transform.NewReader(file, apple.CentralEuropean.NewDecoder())
//	writer := transform.NewWriter(file, apple.CentralEuropean.NewEncoder())
//
// # Data Source
//
// The encoding tables are derived from Unicode Consortium mapping files
// available at https://www.unicode.org/Public/MAPPINGS/VENDORS/APPLE/.
// See the NOTICE file for licensing information.
package apple
