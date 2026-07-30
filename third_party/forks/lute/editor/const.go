// Copyright (c) 2019-present, b3log.org
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package editor

const Caret = "‸"

const CaretNewline = Caret + "\n"

var CaretTokens = []byte(Caret)

var CaretRune = []rune(Caret)[0]

var CaretNewlineTokens = []byte(CaretNewline)

const CaretReplacement = "caretreplacement"

const FrontEndCaret = "<wbr>"

const FrontEndCaretSelfClose = "<wbr/>"

const IALValEscNewLine = "_esc_newline_"

const (
	Zwsp = "\u200b"

	Zwj = "\u200d"
)
