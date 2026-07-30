// SPDX-License-Identifier: BSD-1-Clause
//
// CPython-style lookup tables for Johab encoding.
//
// The actual tables are generated into `tables_generated.go` by cmd/johab-mapgen.
// They use 256-entry index tables pointing into contiguous mapping arrays (similar
// to CPython's dbcs_index/unim_index approach).

package korean

func lookupJohabToUnicode(b1, b2 byte) (rune, bool) {
	idx := johabDecIndex[b1]
	if idx.High == 0 {
		return 0, false
	}
	if b2 < idx.Low || b2 > idx.High {
		return 0, false
	}
	// 0 means unmapped.
	r := johabDecmap[idx.Offset+uint32(b2-idx.Low)]
	if r == 0 {
		return 0, false
	}
	return rune(r), true
}

func lookupUnicodeToJohab(r rune) (b1, b2 byte, ok bool) {
	if r < 0 || r > 0xFFFF {
		return 0, 0, false
	}

	h := byte(uint32(r) >> 8)
	l := byte(r)
	idx := johabEncIndex[h]
	if idx.High == 0 {
		return 0, 0, false
	}
	if l < idx.Low || l > idx.High {
		return 0, 0, false
	}
	code := johabEncmap[idx.Offset+uint32(l-idx.Low)]
	if code == 0 {
		return 0, 0, false
	}
	return byte(code >> 8), byte(code), true
}
