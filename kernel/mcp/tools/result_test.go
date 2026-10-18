// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package tools

import "testing"

func TestSliceRunes(t *testing.T) {
	sliced, total, next := sliceRunes("hello world", 0, 5)
	if sliced != "hello" || total != 11 || next != 5 {
		t.Fatalf("got %q total=%d next=%d", sliced, total, next)
	}
	sliced, total, next = sliceRunes("hello world", 5, 0)
	if sliced != " world" || total != 11 || next != -1 {
		t.Fatalf("got %q total=%d next=%d", sliced, total, next)
	}
	sliced, _, next = sliceRunes("hello", 99, 3)
	if sliced != "" || next != -1 {
		t.Fatalf("got %q next=%d", sliced, next)
	}
	// rune-safe on multibyte text
	sliced, total, next = sliceRunes("héllo", 0, 2)
	if sliced != "hé" || total != 5 || next != 2 {
		t.Fatalf("got %q total=%d next=%d", sliced, total, next)
	}
}
