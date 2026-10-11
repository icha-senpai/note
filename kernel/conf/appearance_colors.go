// Scribli contributors. SPDX-License-Identifier: AGPL-3.0-or-later
package conf

import "regexp"

var interfaceColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func ValidInterfaceColors(colors map[string]map[string]string) bool {
	for mode, values := range colors {
		if mode != "0" && mode != "1" {
			return false
		}
		for role, color := range values {
			if role != "accent" && role != "background" && role != "surface" && role != "text" {
				return false
			}
			if !interfaceColorPattern.MatchString(color) {
				return false
			}
		}
	}
	return true
}
