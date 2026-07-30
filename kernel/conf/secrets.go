// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
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

import (
	"encoding/hex"
	"regexp"

	"github.com/icha-senpai/note/kernel/util"
)

type Secret struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Secrets struct {
	Items []*Secret `json:"items"`
}

func NewSecrets() *Secrets {
	return &Secrets{Items: []*Secret{}}
}

func (s *Secrets) Encrypt() {
	if s == nil {
		return
	}
	for _, item := range s.Items {
		if item == nil || item.Value == "" {
			continue
		}
		item.Value = util.AESEncrypt(item.Value)
	}
}

func (s *Secrets) Decrypt() {
	if s == nil {
		return
	}
	for _, item := range s.Items {
		if item == nil || item.Value == "" {
			continue
		}
		dec := util.AESDecrypt(item.Value)
		if dec == nil {
			continue
		}
		if plain, err := hex.DecodeString(string(dec)); err == nil {
			item.Value = string(plain)
		}
	}
}

var secretPlaceholder = regexp.MustCompile(`\{\{secrets\.([^}]+)\}\}`)

func (s *Secrets) Resolve(in string) string {
	if s == nil {
		return in
	}
	in = secretPlaceholder.ReplaceAllStringFunc(in, func(match string) string {
		sub := secretPlaceholder.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		name := sub[1]
		for _, item := range s.Items {
			if item != nil && item.Name == name {
				return item.Value
			}
		}
		return match
	})
	return resolveDollar(in, s.lookup)
}

func (s *Secrets) lookup(name string) (string, bool) {
	if s == nil {
		return "", false
	}
	for _, item := range s.Items {
		if item != nil && item.Name == name {
			return item.Value, true
		}
	}
	return "", false
}

var dollarPlaceholder = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

func resolveDollar(in string, lookups ...func(string) (string, bool)) string {
	return dollarPlaceholder.ReplaceAllStringFunc(in, func(match string) string {
		sub := dollarPlaceholder.FindStringSubmatch(match)

		name := sub[1]
		if name == "" {
			name = sub[2]
		}
		if name == "" {
			return match
		}
		for _, lk := range lookups {
			if v, ok := lk(name); ok {
				return v
			}
		}
		return match
	})
}

func ResolveSecretsVars(secrets *Secrets, vars *Variables, in string) string {
	in = secrets.Resolve(in)
	return vars.Resolve(in)
}
