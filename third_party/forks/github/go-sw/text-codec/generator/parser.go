// SPDX-License-Identifier: BSD-1-Clause

package generator

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// fetchURL downloads content from a URL.
func fetchURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// parseMappingFile parses Unicode mapping file content.
// Format: 0xNN\t0xNNNN\t# UNICODE NAME
func parseMappingFile(content string) ([]Mapping, error) {
	var mappings []Mapping

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Parse the line
		mapping, err := parseMappingLine(line)
		if err != nil {
			// Skip lines that don't match the expected format
			continue
		}

		mappings = append(mappings, mapping)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return mappings, nil
}

// parseMappingLine parses a single mapping line.
// Format: 0xNN\t0xNNNN\t# UNICODE NAME
// Some encodings may have multi-byte sequences: 0xNNNN\t0xNNNN\t# NAME
func parseMappingLine(line string) (Mapping, error) {
	// Remove comment
	if idx := strings.Index(line, "#"); idx >= 0 {
		comment := strings.TrimSpace(line[idx+1:])
		line = line[:idx]

		parts := strings.Fields(line)
		if len(parts) < 2 {
			return Mapping{}, errors.New("invalid line format")
		}

		// Parse byte value
		byteVal, err := parseHex(parts[0])
		if err != nil {
			return Mapping{}, err
		}

		// Skip multi-byte sequences for single-byte encodings
		if byteVal > 0xFF {
			return Mapping{}, errors.New("multi-byte sequence")
		}

		// Parse Unicode value
		unicodeVal, err := parseHex(parts[1])
		if err != nil {
			return Mapping{}, err
		}

		return Mapping{
			Byte:    byte(byteVal),
			Unicode: rune(unicodeVal),
			Name:    comment,
		}, nil
	}

	return Mapping{}, errors.New("no comment found")
}

// parseHex parses a hexadecimal value in the format 0xNNNN.
func parseHex(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "0x") && !strings.HasPrefix(s, "0X") {
		return 0, fmt.Errorf("invalid hex format: %s", s)
	}
	return strconv.ParseInt(s[2:], 16, 64)
}
