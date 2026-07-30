package b

import (
	. "github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma" // nolint
	"github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma/lexers/internal"
)

// Bnf lexer.
var Bnf = internal.Register(MustNewLazyLexer(
	&Config{
		Name:      "BNF",
		Aliases:   []string{"bnf"},
		Filenames: []string{"*.bnf"},
		MimeTypes: []string{"text/x-bnf"},
	},
	bnfRules,
))

func bnfRules() Rules {
	return Rules{
		"root": {
			{`(<)([ -;=?-~]+)(>)`, ByGroups(Punctuation, NameClass, Punctuation), nil},
			{`::=`, Operator, nil},
			{`[^<>:]+`, Text, nil},
			{`.`, Text, nil},
		},
	}
}
