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

var LangLegacyToBCP47 = map[string]string{
	"en_US": "en",
}

var langBCP47ToLegacy map[string]string

func init() {
	langBCP47ToLegacy = make(map[string]string, len(LangLegacyToBCP47))
	for legacy, bcp47 := range LangLegacyToBCP47 {
		langBCP47ToLegacy[bcp47] = legacy
	}
}

func LangToBCP47(lang string) string {
	if v, ok := LangLegacyToBCP47[lang]; ok {
		return v
	}
	return "en"
}

func LangToLegacy(lang string) string {
	if legacy, ok := langBCP47ToLegacy[lang]; ok {
		return legacy
	}
	return lang
}
