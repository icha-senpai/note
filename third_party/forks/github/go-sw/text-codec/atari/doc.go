// SPDX-License-Identifier: BSD-1-Clause
//
// Mapping data derived from Unicode, Inc. files at:
// https://www.unicode.org/Public/MAPPINGS/VENDORS/MISC/
// See the NOTICE file in the root of this repository for license information.

// Package atari provides the Atari ST text encoding.
//
// This package implements the [github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding] interface
// for the Atari ST/TT (TOS) character encoding, which was used on
// Atari ST, STe, TT, and Falcon computers.
//
// # Supported Encodings
//
//   - [ST] - Atari ST encoding
//
// # Usage
//
// The encoding provides NewDecoder() and NewEncoder() methods:
//
//	import "github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec/atari"
//
//	// Decode Atari ST to UTF-8
//	decoder := atari.ST.NewDecoder()
//	utf8Bytes, err := decoder.Bytes(atariBytes)
//
//	// Encode UTF-8 to Atari ST
//	encoder := atari.ST.NewEncoder()
//	atariBytes, err := encoder.Bytes(utf8Bytes)
//
// For streaming, use with [github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform]:
//
//	import "github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform"
//
//	reader := transform.NewReader(file, atari.ST.NewDecoder())
//	writer := transform.NewWriter(file, atari.ST.NewEncoder())
//
// # Data Source
//
// The encoding table is derived from Unicode Consortium mapping files
// available at https://www.unicode.org/Public/MAPPINGS/VENDORS/MISC/.
// See the NOTICE file for licensing information.
package atari
