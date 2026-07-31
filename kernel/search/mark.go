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
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/gulu"
	"github.com/icha-senpai/note/third_party/forks/lute/lex"
)

func MarkText(text string, keyword string, beforeLen int, caseSensitive bool) (pos int, marked string) {
	if "" == keyword {
		return -1, text
	}
	keywords := SplitKeyword(keyword)
	marked = EncloseHighlighting(text, keywords, "<mark>", "</mark>", caseSensitive, false)

	pos = strings.Index(marked, "<mark>")
	if 0 > pos {
		return
	}

	var before []rune
	var count int
	for i := pos; 0 < i; {
		r, size := utf8.DecodeLastRuneInString(marked[:i])
		i -= size
		before = append([]rune{r}, before...)
		count++
		if beforeLen < count {
			before = append([]rune("..."), before...)
			break
		}
	}
	marked = string(before) + marked[pos:]
	return
}

const (
	TermSep         = "__term@sep__"
	SearchMarkLeft  = "__@mark__"
	SearchMarkRight = "__mark@__"
)

func SplitKeyword(keyword string) (keywords []string) {
	keyword = strings.TrimSpace(keyword)
	if "" == keyword {
		return
	}
	words := strings.Split(keyword, TermSep)
	if 1 < len(words) {
		for _, word := range words {
			if "" == word {
				continue
			}
			keywords = append(keywords, word)
		}
	} else {
		keywords = append(keywords, keyword)
	}
	return
}

func EncloseHighlighting(text string, keywords []string, openMark, closeMark string, caseSensitive, splitWords bool) (ret string) {
	ic := "(?i)"
	if caseSensitive {
		ic = "(?)"
	}

	var re strings.Builder
	re.WriteString(ic + "(")
	for i, k := range keywords {
		if "" == k {
			continue
		}

		wordBoundary := false
		if splitWords {
			wordBoundary = lex.IsASCIILetterNums(gulu.Str.ToBytes(k)) // Improve virtual reference split words
		}
		if !util.SearchHanSensitive {

			k = hanInsensitiveRegexp(util.EscapeHTML(k))
		} else {
			k = regexp.QuoteMeta(util.EscapeHTML(k))
		}
		re.WriteString("(")
		if wordBoundary {
			re.WriteString("\\b")
		}
		re.WriteString(k)
		if wordBoundary {
			re.WriteString("\\b")
		}
		re.WriteString(")")
		if i < len(keywords)-1 {
			re.WriteString("|")
		}
	}
	re.WriteString(")")

	ret = util.EscapeHTML(text)

	ret = strings.ReplaceAll(ret, "&#34;", "\ue000")
	ret = strings.ReplaceAll(ret, "&lt;", "\ue001")
	ret = strings.ReplaceAll(ret, "&gt;", "\ue002")
	ret = strings.ReplaceAll(ret, "&#39;", "\ue003")
	if reg, err := regexp.Compile(re.String()); err == nil {
		ret = reg.ReplaceAllStringFunc(ret, func(s string) string { return openMark + s + closeMark })
	}
	ret = strings.ReplaceAll(ret, "\ue000", "&#34;")
	ret = strings.ReplaceAll(ret, "\ue001", "&lt;")
	ret = strings.ReplaceAll(ret, "\ue002", "&gt;")
	ret = strings.ReplaceAll(ret, "\ue003", "&#39;")

	ret = strings.ReplaceAll(ret, "\\<span", "\\\\<span")
	return
}

const (
	MarkDataType            = "search-mark"
	VirtualBlockRefDataType = "virtual-block-ref"
)

func GetMarkSpanStart(dataType string) string {
	return fmt.Sprintf("<span data-type=\"%s\">", dataType)
}

func GetMarkSpanEnd() string {
	return "</span>"
}
