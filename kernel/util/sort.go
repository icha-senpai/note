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

package util

import (
	"strings"

	"github.com/facette/natsort"
)

func NaturalCompare(str1, str2 string) bool {
	str1 = RemoveEmojiInvisible(str1)
	str2 = RemoveEmojiInvisible(str2)
	return natsort.Compare(str1, str2)
}

func EmojiLexicalCompare(str1, str2 string) bool {
	str1_ := strings.TrimSpace(RemoveEmojiInvisible(str1))
	str2_ := strings.TrimSpace(RemoveEmojiInvisible(str2))
	if str1_ == str2_ && 0 == len(str1_) {

		return strings.Compare(str1, str2) < 0
	}
	return LexicalCompare(str1, str2)
}

func LexicalCompare(str1, str2 string) bool {
	str1 = RemoveEmojiInvisible(str1)
	str2 = RemoveEmojiInvisible(str2)

	// Doc tree, backlinks, tags and templates ignores case when sorting alphabetically by name
	str1 = strings.ToLower(str1)
	str2 = strings.ToLower(str2)

	return strings.Compare(str1, str2) < 0
}

func LexicalCompare4FileTree(str1, str2 string) bool {

	// Improve doc tree Name Alphabet sorting

	str1 = RemoveEmojiInvisible(str1)
	str1 = strings.TrimSuffix(str1, ".sy")
	str2 = RemoveEmojiInvisible(str2)
	str2 = strings.TrimSuffix(str2, ".sy")

	// Doc tree, backlinks, tags and templates ignores case when sorting alphabetically by name
	str1 = strings.ToLower(str1)
	str2 = strings.ToLower(str2)

	return strings.Compare(str1, str2) < 0
}

const (
	SortModeNameASC = iota
	SortModeNameDESC
	SortModeUpdatedASC
	SortModeUpdatedDESC
	SortModeAlphanumASC
	SortModeAlphanumDESC
	SortModeCustom
	SortModeRefCountASC
	SortModeRefCountDESC
	SortModeCreatedASC
	SortModeCreatedDESC
	SortModeSizeASC
	SortModeSizeDESC
	SortModeSubDocCountASC
	SortModeSubDocCountDESC
	SortModeFileTree

	SortModeUnassigned = 256
)
