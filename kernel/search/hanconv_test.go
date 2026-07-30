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

package search

import (
	"regexp"
	"testing"

	"github.com/icha-senpai/note/kernel/util"
)

func TestHanInsensitiveRegexp(t *testing.T) {
	const (
		poetryClassicSimplified  = "\u8bd7\u7ecf"
		poetryClassicTraditional = "\u8a69\u7d93"
		poetryMixedOne           = "\u8bd7\u7d93"
		poetryMixedTwo           = "\u8a69\u7ecf"
		poetryBookDifferent      = "\u8bd7\u4e66"
		hairTraditionalVariant   = "\u9aee"
		hairSimplified           = "\u53d1"
		hairTraditional          = "\u767c"
		middleA1                 = "\u4e2da1"
	)

	re, err := regexp.Compile("^" + hanInsensitiveRegexp(poetryClassicSimplified) + "$")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{poetryClassicSimplified, poetryClassicTraditional, poetryMixedOne, poetryMixedTwo} {
		if !re.MatchString(s) {
			t.Errorf("hanInsensitiveRegexp(%q) should match %q", poetryClassicSimplified, s)
		}
	}
	if re.MatchString(poetryBookDifferent) {
		t.Errorf("hanInsensitiveRegexp(%q) should not match %q", poetryClassicSimplified, poetryBookDifferent)
	}

	re2 := regexp.MustCompile("^" + hanInsensitiveRegexp(hairTraditionalVariant) + "$")
	for _, s := range []string{hairTraditionalVariant, hairSimplified, hairTraditional} {
		if !re2.MatchString(s) {
			t.Errorf("hanInsensitiveRegexp(%q) should match %q as an equivalent form of %q", hairTraditionalVariant, s, hairSimplified)
		}
	}

	if got := hanInsensitiveRegexp(middleA1); middleA1 != got {
		t.Errorf("hanInsensitiveRegexp(%q) = %q, should be unchanged", middleA1, got)
	}
}

func TestEncloseHighlightingHanInsensitive(t *testing.T) {
	const (
		poetryClassicSimplified  = "\u8bd7\u7ecf"
		poetryClassicTraditional = "\u8a69\u7d93"
		poetryResearch           = "\u8a69\u7d93\u7814\u7a76"
		research                 = "\u7814\u7a76"
	)

	old := util.SearchHanSensitive
	defer func() { util.SearchHanSensitive = old }()

	util.SearchHanSensitive = false
	got := EncloseHighlighting(poetryResearch, []string{poetryClassicSimplified}, "<mark>", "</mark>", false, false)
	if want := "<mark>" + poetryClassicTraditional + "</mark>" + research; want != got {
		t.Errorf("Han-insensitive highlighting = %q, want %q", got, want)
	}

	got = EncloseHighlighting("ABC "+poetryClassicTraditional, []string{"abc"}, "<mark>", "</mark>", false, false)
	if want := "<mark>ABC</mark> " + poetryClassicTraditional; want != got {
		t.Errorf("case-insensitive and Han-insensitive highlighting = %q, want %q", got, want)
	}

	util.SearchHanSensitive = true
	got = EncloseHighlighting(poetryResearch, []string{poetryClassicSimplified}, "<mark>", "</mark>", false, false)
	if want := poetryResearch; want != got {
		t.Errorf("default Han-sensitive highlighting should not mark the text, got %q", got)
	}
}
