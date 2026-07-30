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

package search

import (
	"regexp"
	"testing"

	"github.com/icha-senpai/note/kernel/util"
)

func TestHanInsensitiveRegexp(t *testing.T) {
	re, err := regexp.Compile("^" + hanInsensitiveRegexp("诗经") + "$")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"诗经", "詩經", "诗經", "詩经"} {
		if !re.MatchString(s) {
			t.Errorf("hanInsensitiveRegexp(诗经) should match %q", s)
		}
	}
	if re.MatchString("诗书") {
		t.Errorf("hanInsensitiveRegexp(诗经) should not match 诗书")
	}

	re2 := regexp.MustCompile("^" + hanInsensitiveRegexp("髮") + "$")
	for _, s := range []string{"髮", "发", "發"} {
		if !re2.MatchString(s) {
			t.Errorf("hanInsensitiveRegexp(髮) should match %q as an equivalent form of 发", s)
		}
	}

	if got := hanInsensitiveRegexp("中a1"); "中a1" != got {
		t.Errorf("hanInsensitiveRegexp(中a1) = %q, should be unchanged", got)
	}
}

func TestEncloseHighlightingHanInsensitive(t *testing.T) {
	old := util.SearchHanSensitive
	defer func() { util.SearchHanSensitive = old }()

	util.SearchHanSensitive = false
	got := EncloseHighlighting("詩經研究", []string{"诗经"}, "<mark>", "</mark>", false, false)
	if want := "<mark>詩經</mark>研究"; want != got {
		t.Errorf("Han-insensitive highlighting = %q, want %q", got, want)
	}

	got = EncloseHighlighting("ABC 詩經", []string{"abc"}, "<mark>", "</mark>", false, false)
	if want := "<mark>ABC</mark> 詩經"; want != got {
		t.Errorf("case-insensitive and Han-insensitive highlighting = %q, want %q", got, want)
	}

	util.SearchHanSensitive = true
	got = EncloseHighlighting("詩經研究", []string{"诗经"}, "<mark>", "</mark>", false, false)
	if want := "詩經研究"; want != got {
		t.Errorf("default Han-sensitive highlighting should not mark the text, got %q", got)
	}
}
