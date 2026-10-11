// Scribli contributors. SPDX-License-Identifier: AGPL-3.0-or-later
package conf

import "testing"

func TestInterfaceColorsRejectInvalidRolesAndValues(t *testing.T) {
	if !ValidInterfaceColors(nil) || !ValidInterfaceColors(map[string]map[string]string{"0": {"accent": "#aabbcc"}, "1": {"background": "#123456"}}) {
		t.Fatal("valid overrides rejected")
	}
	for _, colors := range []map[string]map[string]string{{"2": {"accent": "#aabbcc"}}, {"0": {"unknown": "#aabbcc"}}, {"1": {"text": "red;display:none"}}} {
		if ValidInterfaceColors(colors) {
			t.Fatal("invalid overrides accepted", colors)
		}
	}
}
