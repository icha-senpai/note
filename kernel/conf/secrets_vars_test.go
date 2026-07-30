// Scribli - Refactor your thinking
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

package conf

import "testing"

func TestSecretsResolveExplicit(t *testing.T) {
	s := &Secrets{Items: []*Secret{{Name: "KEY", Value: "k1"}}}
	cases := []struct{ in, want string }{
		{"{{secrets.KEY}}", "k1"},
		{"pre {{secrets.KEY}} post", "pre k1 post"},
		{"{{secrets.MISSING}}", "{{secrets.MISSING}}"},
		{"$KEY", "k1"},
		{"${KEY}", "k1"},
		{"$MISSING", "$MISSING"},
		{"$100", "$100"},
		{"", ""},
	}
	for _, c := range cases {
		if got := s.Resolve(c.in); got != c.want {
			t.Errorf("Secrets.Resolve(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestVariablesResolveExplicit(t *testing.T) {
	v := &Variables{Items: []*Variable{{Name: "VAR", Value: "v1"}}}
	cases := []struct{ in, want string }{
		{"{{vars.VAR}}", "v1"},
		{"{{vars.MISSING}}", "{{vars.MISSING}}"},
		{"$VAR", "v1"},
		{"${VAR}", "v1"},
		{"$MISSING", "$MISSING"},
	}
	for _, c := range cases {
		if got := v.Resolve(c.in); got != c.want {
			t.Errorf("Variables.Resolve(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPerStoreDollarIsolation(t *testing.T) {
	secrets := &Secrets{Items: []*Secret{{Name: "ONLY_SECRET", Value: "s"}}}
	vars := &Variables{Items: []*Variable{{Name: "ONLY_VAR", Value: "v"}}}

	if got := secrets.Resolve("$ONLY_VAR"); got != "$ONLY_VAR" {
		t.Errorf("secrets.Resolve($ONLY_VAR) = %q, want $ONLY_VAR (secret store should not resolve variables across stores)", got)
	}

	if got := vars.Resolve("$ONLY_SECRET"); got != "$ONLY_SECRET" {
		t.Errorf("vars.Resolve($ONLY_SECRET) = %q, want $ONLY_SECRET (variable store should not resolve secrets across stores)", got)
	}
}

func TestResolveSecretsVarsDollarSyntax(t *testing.T) {
	secrets := &Secrets{Items: []*Secret{
		{Name: "WEREAD_API_KEY", Value: "wrk-secret"},
		{Name: "KEY", Value: "k1"},
	}}
	vars := &Variables{Items: []*Variable{{Name: "VAR", Value: "v1"}}}

	cases := []struct{ in, want string }{

		{"Bearer $WEREAD_API_KEY", "Bearer wrk-secret"},
		{"Bearer ${WEREAD_API_KEY}", "Bearer wrk-secret"},

		{"price $100", "price $100"},
		{"regex $1 group", "regex $1 group"},

		{"$A $B end", "$A $B end"},

		{"{{secrets.KEY}}-{{vars.VAR}}", "k1-v1"},

		{"$VAR", "v1"},

		{"mix $KEY and ${VAR} and literal", "mix k1 and v1 and literal"},
	}
	for _, c := range cases {
		if got := ResolveSecretsVars(secrets, vars, c.in); got != c.want {
			t.Errorf("ResolveSecretsVars(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestResolveSecretsVarsPriority(t *testing.T) {
	secrets := &Secrets{Items: []*Secret{{Name: "dup", Value: "from-secret"}}}
	vars := &Variables{Items: []*Variable{{Name: "dup", Value: "from-var"}}}

	if got := ResolveSecretsVars(secrets, vars, "$dup"); got != "from-secret" {
		t.Errorf("$dup = %q, want from-secret", got)
	}
	if got := ResolveSecretsVars(secrets, vars, "${dup}"); got != "from-secret" {
		t.Errorf("${dup} = %q, want from-secret", got)
	}

	if g := ResolveSecretsVars(secrets, vars, "{{secrets.dup}}"); g != "from-secret" {
		t.Errorf("{{secrets.dup}} = %q, want from-secret", g)
	}
	if g := ResolveSecretsVars(secrets, vars, "{{vars.dup}}"); g != "from-var" {
		t.Errorf("{{vars.dup}} = %q, want from-var", g)
	}
}

func TestResolveSecretsVarsNilSafe(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ResolveSecretsVars with nil panicked: %v", r)
		}
	}()

	if got := ResolveSecretsVars(nil, nil, "$X"); got != "$X" {
		t.Errorf("ResolveSecretsVars(nil,nil) = %q, want $X", got)
	}
	if got := ResolveSecretsVars(nil, nil, "plain text"); got != "plain text" {
		t.Errorf("ResolveSecretsVars(nil,nil,plain) = %q, want plain text", got)
	}
}

func TestSecretsLookup(t *testing.T) {
	s := &Secrets{Items: []*Secret{{Name: "a", Value: "1"}}}
	if v, ok := s.lookup("a"); !ok || v != "1" {
		t.Errorf("lookup(a) = %q,%v, want 1,true", v, ok)
	}
	if _, ok := s.lookup("missing"); ok {
		t.Error("lookup(missing) should be not ok")
	}
}
