package b

import (
	. "github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma" // nolint
	"github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma/lexers/internal"
)

// Brainfuck lexer.
var Brainfuck = internal.Register(MustNewLazyLexer(
	&Config{
		Name:      "Brainfuck",
		Aliases:   []string{"brainfuck", "bf"},
		Filenames: []string{"*.bf", "*.b"},
		MimeTypes: []string{"application/x-brainfuck"},
	},
	brainfuckRules,
))

func brainfuckRules() Rules {
	return Rules{
		"common": {
			{`[.,]+`, NameTag, nil},
			{`[+-]+`, NameBuiltin, nil},
			{`[<>]+`, NameVariable, nil},
			{`[^.,+\-<>\[\]]+`, Comment, nil},
		},
		"root": {
			{`\[`, Keyword, Push("loop")},
			{`\]`, Error, nil},
			Include("common"),
		},
		"loop": {
			{`\[`, Keyword, Push()},
			{`\]`, Keyword, Pop(1)},
			Include("common"),
		},
	}
}
