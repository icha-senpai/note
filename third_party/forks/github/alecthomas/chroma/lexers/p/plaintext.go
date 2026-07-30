package p

import (
	. "github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma" // nolint
	"github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma/lexers/internal"
)

var Plaintext = internal.Register(MustNewLazyLexer(
	&Config{
		Name:      "plaintext",
		Aliases:   []string{"text", "plain", "no-highlight"},
		Filenames: []string{"*.txt"},
		MimeTypes: []string{"text/plain"},
		Priority:  0.1,
	},
	internal.PlaintextRules,
))
