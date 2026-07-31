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

package model

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/ascii85"
	"encoding/hex"
	"errors"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

var pdfPageTypePattern = regexp.MustCompile(`/Type\s*/Page\b`)

func countPDFPages(data []byte) int {
	return len(pdfPageTypePattern.FindAll(data, -1))
}

func extractPDFText(data []byte) (string, error) {
	if bytes.Contains(data, []byte("/Encrypt")) {
		return "", errors.New("encrypted PDF text extraction is not supported")
	}

	var text strings.Builder
	forEachPDFStream(data, func(dict, stream []byte) {
		decoded, ok := decodePDFStream(dict, stream)
		if !ok {
			return
		}
		streamText := extractPDFTextBlocks(decoded)
		if streamText == "" {
			return
		}
		if text.Len() > 0 {
			text.WriteByte(' ')
		}
		text.WriteString(streamText)
	})

	if text.Len() == 0 {
		return "", errors.New("no extractable PDF text found")
	}
	return text.String(), nil
}

func forEachPDFStream(data []byte, visit func(dict, stream []byte)) {
	const dictWindow = 8192
	streamToken := []byte("stream")
	endStreamToken := []byte("endstream")

	offset := 0
	for {
		streamIndex := bytes.Index(data[offset:], streamToken)
		if streamIndex < 0 {
			return
		}
		streamIndex += offset
		streamStart := streamIndex + len(streamToken)
		if streamStart < len(data) && data[streamStart] == '\r' {
			streamStart++
		}
		if streamStart < len(data) && data[streamStart] == '\n' {
			streamStart++
		}

		endIndex := bytes.Index(data[streamStart:], endStreamToken)
		if endIndex < 0 {
			return
		}
		endIndex += streamStart

		dictStartSearch := streamIndex - dictWindow
		if dictStartSearch < 0 {
			dictStartSearch = 0
		}
		dictStart := bytes.LastIndex(data[dictStartSearch:streamIndex], []byte("<<"))
		if dictStart < 0 {
			offset = endIndex + len(endStreamToken)
			continue
		}
		dictStart += dictStartSearch
		dict := data[dictStart:streamIndex]
		stream := bytes.TrimRight(data[streamStart:endIndex], "\r\n")
		visit(dict, stream)
		offset = endIndex + len(endStreamToken)
	}
}

func decodePDFStream(dict, stream []byte) ([]byte, bool) {
	if hasPDFName(dict, "Image") || hasPDFName(dict, "JPXDecode") || hasPDFName(dict, "DCTDecode") {
		return nil, false
	}

	filters := pdfStreamFilters(dict)
	decoded := append([]byte(nil), stream...)
	for _, filter := range filters {
		var err error
		switch filter {
		case "FlateDecode", "Fl":
			decoded, err = inflatePDFStream(decoded)
		case "ASCIIHexDecode", "AHx":
			decoded, err = decodeASCIIHex(decoded)
		case "ASCII85Decode", "A85":
			decoded, err = decodeASCII85(decoded)
		default:
			return nil, false
		}
		if err != nil {
			return nil, false
		}
	}
	return decoded, true
}

func pdfStreamFilters(dict []byte) []string {
	names := []string{"ASCIIHexDecode", "AHx", "ASCII85Decode", "A85", "FlateDecode", "Fl"}
	type foundFilter struct {
		name  string
		index int
	}

	var found []foundFilter
	for _, name := range names {
		index := indexPDFName(dict, name)
		if index >= 0 {
			found = append(found, foundFilter{name: name, index: index})
		}
	}
	sort.Slice(found, func(i, j int) bool {
		return found[i].index < found[j].index
	})

	ret := make([]string, 0, len(found))
	for _, filter := range found {
		ret = append(ret, filter.name)
	}
	return ret
}

func hasPDFName(data []byte, name string) bool {
	return indexPDFName(data, name) >= 0
}

func indexPDFName(data []byte, name string) int {
	needle := []byte("/" + name)
	offset := 0
	for {
		index := bytes.Index(data[offset:], needle)
		if index < 0 {
			return -1
		}
		index += offset
		afterIndex := index + len(needle)
		if afterIndex >= len(data) || isPDFDelimiter(data[afterIndex]) {
			return index
		}
		offset = afterIndex
	}
}

func inflatePDFStream(data []byte) ([]byte, error) {
	if reader, err := zlib.NewReader(bytes.NewReader(data)); err == nil {
		defer reader.Close()
		return io.ReadAll(reader)
	}

	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()
	return io.ReadAll(reader)
}

func decodeASCII85(data []byte) ([]byte, error) {
	data = bytes.TrimSpace(data)
	data = bytes.TrimPrefix(data, []byte("<~"))
	data = bytes.TrimSuffix(data, []byte("~>"))
	return io.ReadAll(ascii85.NewDecoder(bytes.NewReader(data)))
}

func decodeASCIIHex(data []byte) ([]byte, error) {
	var cleaned []byte
	for _, b := range data {
		if b == '>' {
			break
		}
		if isPDFWhitespace(b) {
			continue
		}
		cleaned = append(cleaned, b)
	}
	if len(cleaned)%2 == 1 {
		cleaned = append(cleaned, '0')
	}
	decoded := make([]byte, hex.DecodedLen(len(cleaned)))
	_, err := hex.Decode(decoded, cleaned)
	return decoded, err
}

func extractPDFTextBlocks(data []byte) string {
	var text strings.Builder
	inText := false
	for i := 0; i < len(data); {
		if isPDFOperatorAt(data, i, "BT") {
			inText = true
			i += 2
			continue
		}
		if isPDFOperatorAt(data, i, "ET") {
			inText = false
			i += 2
			continue
		}
		if !inText {
			i++
			continue
		}

		switch data[i] {
		case '(':
			raw, next := parsePDFLiteralString(data, i)
			writePDFText(&text, decodePDFTextBytes(raw))
			i = next
		case '<':
			if i+1 < len(data) && data[i+1] == '<' {
				i += 2
				continue
			}
			raw, next := parsePDFHexString(data, i)
			writePDFText(&text, decodePDFTextBytes(raw))
			i = next
		default:
			if isPDFOperatorAt(data, i, "T*") || isPDFOperatorAt(data, i, "Td") || isPDFOperatorAt(data, i, "TD") ||
				isPDFOperatorAt(data, i, "'") || isPDFOperatorAt(data, i, "\"") {
				text.WriteByte(' ')
			}
			i++
		}
	}
	return text.String()
}

func writePDFText(text *strings.Builder, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if text.Len() > 0 {
		text.WriteByte(' ')
	}
	text.WriteString(value)
}

func parsePDFLiteralString(data []byte, start int) ([]byte, int) {
	var ret []byte
	depth := 1
	for i := start + 1; i < len(data); i++ {
		b := data[i]
		if b == '\\' {
			if i+1 >= len(data) {
				return ret, len(data)
			}
			next := data[i+1]
			switch next {
			case 'n':
				ret = append(ret, '\n')
				i++
			case 'r':
				ret = append(ret, '\r')
				i++
			case 't':
				ret = append(ret, '\t')
				i++
			case 'b':
				ret = append(ret, '\b')
				i++
			case 'f':
				ret = append(ret, '\f')
				i++
			case '(', ')', '\\':
				ret = append(ret, next)
				i++
			case '\r', '\n':
				i++
				if next == '\r' && i+1 < len(data) && data[i+1] == '\n' {
					i++
				}
			default:
				if next >= '0' && next <= '7' {
					value := 0
					count := 0
					for i+1 < len(data) && count < 3 {
						c := data[i+1]
						if c < '0' || c > '7' {
							break
						}
						value = value*8 + int(c-'0')
						i++
						count++
					}
					ret = append(ret, byte(value))
				} else {
					ret = append(ret, next)
					i++
				}
			}
			continue
		}
		if b == '(' {
			depth++
		}
		if b == ')' {
			depth--
			if depth == 0 {
				return ret, i + 1
			}
		}
		ret = append(ret, b)
	}
	return ret, len(data)
}

func parsePDFHexString(data []byte, start int) ([]byte, int) {
	i := start + 1
	var cleaned []byte
	for ; i < len(data); i++ {
		if data[i] == '>' {
			i++
			break
		}
		if isPDFWhitespace(data[i]) {
			continue
		}
		cleaned = append(cleaned, data[i])
	}
	if len(cleaned)%2 == 1 {
		cleaned = append(cleaned, '0')
	}
	decoded := make([]byte, hex.DecodedLen(len(cleaned)))
	n, err := hex.Decode(decoded, cleaned)
	if err != nil {
		return nil, i
	}
	return decoded[:n], i
}

func decodePDFTextBytes(data []byte) string {
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		var runes []uint16
		for i := 2; i+1 < len(data); i += 2 {
			runes = append(runes, uint16(data[i])<<8|uint16(data[i+1]))
		}
		return string(utf16.Decode(runes))
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		var runes []uint16
		for i := 2; i+1 < len(data); i += 2 {
			runes = append(runes, uint16(data[i])|uint16(data[i+1])<<8)
		}
		return string(utf16.Decode(runes))
	}
	if utf8.Valid(data) {
		return string(data)
	}

	runes := make([]rune, 0, len(data))
	for _, b := range data {
		if b == 0 {
			continue
		}
		runes = append(runes, rune(b))
	}
	return string(runes)
}

func isPDFOperatorAt(data []byte, index int, op string) bool {
	if index < 0 || index+len(op) > len(data) || string(data[index:index+len(op)]) != op {
		return false
	}
	beforeOK := index == 0 || isPDFDelimiter(data[index-1])
	afterIndex := index + len(op)
	afterOK := afterIndex >= len(data) || isPDFDelimiter(data[afterIndex])
	return beforeOK && afterOK
}

func isPDFDelimiter(b byte) bool {
	return isPDFWhitespace(b) || strings.ContainsRune("[]<>()/{}", rune(b))
}

func isPDFWhitespace(b byte) bool {
	switch b {
	case 0, '\t', '\n', '\f', '\r', ' ':
		return true
	default:
		return false
	}
}
