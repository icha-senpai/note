// SiYuan - Refactor your thinking
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
	"slices"
	"strings"
)

var hanSimpToTrads = map[rune][]rune{}

func init() {
	for t, s := range hanTradToSimp {
		hanSimpToTrads[s] = append(hanSimpToTrads[s], t)
	}
	for _, ts := range hanSimpToTrads {
		slices.Sort(ts)
	}
}

func hanCharClass(r rune) (ret []rune) {
	canon := r
	if s, ok := hanTradToSimp[r]; ok {
		canon = s
	}
	ret = append(ret, canon)
	ret = append(ret, hanSimpToTrads[canon]...)
	return
}

func hanInsensitiveRegexp(k string) string {
	var b strings.Builder
	for _, r := range k {
		class := hanCharClass(r)
		if 1 == len(class) {
			b.WriteString(regexp.QuoteMeta(string(r)))
			continue
		}
		b.WriteString("[")
		for _, c := range class {
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
		b.WriteString("]")
	}
	return b.String()
}
