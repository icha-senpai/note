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

import "regexp"

type Variable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Variables struct {
	Items []*Variable `json:"items"`
}

func NewVariables() *Variables {
	return &Variables{Items: []*Variable{}}
}

var varPlaceholder = regexp.MustCompile(`\{\{vars\.([^}]+)\}\}`)

func (v *Variables) Resolve(in string) string {
	if v == nil {
		return in
	}
	in = varPlaceholder.ReplaceAllStringFunc(in, func(match string) string {
		sub := varPlaceholder.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		name := sub[1]
		for _, item := range v.Items {
			if item != nil && item.Name == name {
				return item.Value
			}
		}
		return match
	})
	return resolveDollar(in, v.lookup)
}

func (v *Variables) lookup(name string) (string, bool) {
	if v == nil {
		return "", false
	}
	for _, item := range v.Items {
		if item != nil && item.Name == name {
			return item.Value, true
		}
	}
	return "", false
}
