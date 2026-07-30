// Copyright (c) 2019-present, Scribli
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package lex

import "unicode/utf8"

type Lexer struct {
	input  []byte
	length int
	offset int
	width  int
}

func NewLexer(input []byte) (ret *Lexer) {
	ret = &Lexer{input: input, length: len(input)}
	if 0 < ret.length && ItemNewline != ret.input[ret.length-1] {
		ret.input = append(ret.input, ItemNewline)
		ret.length++
	}
	return
}

func (l *Lexer) NextLine() (ret []byte) {
	if l.offset >= l.length {
		return
	}

	var b, nb byte
	i := l.offset
	for ; i < l.length; i += l.width {
		b = l.input[i]
		if ItemNewline == b {
			i++
			break
		} else if ItemCarriageReturn == b {
			if i < l.length-1 {
				nb = l.input[i+1]
				if ItemNewline == nb { // \r\n
					l.input = append(l.input[:i], l.input[i+1:]...)
					l.length--
				} else { // \rX
					l.input[i] = ItemNewline
				}
			} else { // \rEOF
				l.input[i] = ItemNewline
			}
			i++
			break
		} else if '\u0000' == b {
			l.input = append(l.input, 0, 0)
			copy(l.input[i+2:], l.input[i:])
			l.input[i], l.input[i+1], l.input[i+2] = '\xEF', '\xBF', '\xBD'
			l.length += 2
			l.width = 3
			continue
		}

		if utf8.RuneSelf <= b {
			_, l.width = utf8.DecodeRune(l.input[i:])
		} else {
			l.width = 1
		}
	}
	ret = l.input[l.offset:i]
	l.offset = i
	return
}
