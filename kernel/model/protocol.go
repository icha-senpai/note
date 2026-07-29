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

package model

import "strings"

const (
	blockProtocolPrefix       = "scribli://blocks/"
	legacyBlockProtocolPrefix = "siyuan://blocks/"
)

func makeBlockProtocolURL(id string) string {
	return blockProtocolPrefix + id
}

func cutBlockProtocolURL(url string) (string, bool) {
	if after, ok := strings.CutPrefix(url, blockProtocolPrefix); ok {
		return after, true
	}
	return strings.CutPrefix(url, legacyBlockProtocolPrefix)
}

func trimBlockProtocolURL(url string) string {
	if after, ok := cutBlockProtocolURL(url); ok {
		return after
	}
	return url
}
