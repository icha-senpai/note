package d

import (
	. "github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma" // nolint
	"github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma/lexers/internal"
)

// Diff lexer.
var Diff = internal.Register(MustNewLazyLexer(
	&Config{
		Name:      "Diff",
		Aliases:   []string{"diff", "udiff"},
		EnsureNL:  true,
		Filenames: []string{"*.diff", "*.patch"},
		MimeTypes: []string{"text/x-diff", "text/x-patch"},
	},
	diffRules,
))

func diffRules() Rules {
	return Rules{
		"root": {
			{` .*\n`, Text, nil},
			{`\+.*\n`, GenericInserted, nil},
			{`-.*\n`, GenericDeleted, nil},
			{`!.*\n`, GenericStrong, nil},
			{`@.*\n`, GenericSubheading, nil},
			{`([Ii]ndex|diff).*\n`, GenericHeading, nil},
			{`=.*\n`, GenericHeading, nil},
			{`.*\n`, Text, nil},
		},
	}
}
