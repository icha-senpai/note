// Scribli - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
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

package model

import "testing"

func TestBlockProtocolURL(t *testing.T) {
	const id = "20260729000000-abcdefg"

	if got := makeBlockProtocolURL(id); got != "scribli://blocks/"+id {
		t.Fatalf("unexpected primary block URL %q", got)
	}

	got, ok := cutBlockProtocolURL("scribli://blocks/" + id)
	if !ok {
		t.Fatal("expected scribli block URL to be recognized")
	}
	if got != id {
		t.Fatalf("expected %q, got %q", id, got)
	}

	for _, externalURL := range []string{
		"other://blocks/" + id,
		"https://example.com/" + id,
	} {
		got, ok = cutBlockProtocolURL(externalURL)
		if ok || got != externalURL {
			t.Fatalf("unexpected external URL match [got=%q, ok=%v]", got, ok)
		}
	}
}
