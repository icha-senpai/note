package css

import "github.com/icha-senpai/note/third_party/forks/github/gorilla/css/scanner"

func skipRules(s *scanner.Scanner) {
	var (
		open    int
		close   int
		started bool
	)
	for {
		if started && close >= open {
			return
		}
		token := s.Next()
		if token.Type == scanner.TokenEOF || token.Type == scanner.TokenError {
			return
		}
		if token.Type == scanner.TokenChar {
			if token.Value == "{" {
				open++
				started = true
				continue
			}
			if token.Value == "}" {
				close++
				started = true
				continue
			}
		}
	}
}
