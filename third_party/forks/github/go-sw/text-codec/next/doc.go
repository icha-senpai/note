// SPDX-License-Identifier: BSD-1-Clause
//
// Mapping data derived from Unicode, Inc. files at:
// https://www.unicode.org/Public/MAPPINGS/VENDORS/NEXT/
// See the NOTICE file in the root of this repository for license information.

// Package next provides the NeXTSTEP text encoding.
//
// This package implements the [github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding] interface
// for the NeXTSTEP character encoding, which was used on NeXT computers
// and early versions of macOS (before the transition to Unicode).
//
// # Supported Encodings
//
//   - [NeXTSTEP] - NeXTSTEP encoding
//
// # Usage
//
// The encoding provides NewDecoder() and NewEncoder() methods:
//
//	import "github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec/next"
//
//	// Decode NeXTSTEP to UTF-8
//	decoder := next.NeXTSTEP.NewDecoder()
//	utf8Bytes, err := decoder.Bytes(nextStepBytes)
//
//	// Encode UTF-8 to NeXTSTEP
//	encoder := next.NeXTSTEP.NewEncoder()
//	nextStepBytes, err := encoder.Bytes(utf8Bytes)
//
// For streaming, use with [github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform]:
//
//	import "github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform"
//
//	reader := transform.NewReader(file, next.NeXTSTEP.NewDecoder())
//	writer := transform.NewWriter(file, next.NeXTSTEP.NewEncoder())
//
// # Data Source
//
// The encoding table is derived from Unicode Consortium mapping files
// available at https://www.unicode.org/Public/MAPPINGS/VENDORS/NEXT/.
// See the NOTICE file for licensing information.
package next
