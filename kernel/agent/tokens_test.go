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

package agent

import "testing"

func TestTokenCounterUsesLocalEstimator(t *testing.T) {
	counter, err := getTokenCounter("any-model")
	if err != nil {
		t.Fatal(err)
	}
	if counter == nil {
		t.Fatal("expected token counter")
	}
	if got := counter.count("abc"); got != 1 {
		t.Fatalf("short text should round up to one token, got %d", got)
	}
	if got := counter.count(""); got != 0 {
		t.Fatalf("empty text should stay zero tokens, got %d", got)
	}
	if got := counter.count("你好"); got < 1 {
		t.Fatalf("CJK text should estimate to at least one token, got %d", got)
	}
}
