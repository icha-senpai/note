// Package quick provides simple, no-configuration source code highlighting.
package quick

import (
	"io"

	"github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma"
	"github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma/formatters"
	"github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma/lexers"
	"github.com/icha-senpai/note/third_party/forks/github/alecthomas/chroma/styles"
)

// Highlight some text.
//
// Lexer, formatter and style may be empty, in which case a best-effort is made.
func Highlight(w io.Writer, source, lexer, formatter, style string) error {
	// Determine lexer.
	l := lexers.Get(lexer)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	// Determine formatter.
	f := formatters.Get(formatter)
	if f == nil {
		f = formatters.Fallback
	}

	// Determine style.
	s := styles.Get(style)
	if s == nil {
		s = styles.Fallback
	}

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return f.Format(w, s, it)
}
