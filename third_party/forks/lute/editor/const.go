// Copyright (c) 2019-present, Scribli


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
