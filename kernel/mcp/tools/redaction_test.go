// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tools

import (
	"strings"
	"testing"
)

func TestRedactSensitiveText(t *testing.T) {
	input := `api_key = "sk-proj-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" bearer Bearer ghp_bbbbbbbbbbbbbbbbbbbbbbbbbbbb`
	output := redactSensitiveText(input)
	if strings.Contains(output, "sk-proj-") || strings.Contains(output, "ghp_") {
		t.Fatalf("secret was not redacted: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Fatalf("redaction marker missing: %s", output)
	}
}

func TestRedactSensitiveTextWithSearchMarks(t *testing.T) {
	input := `token = "<mark>sk-proj</mark>-cccccccccccccccccccccccccccccccc"`
	output := redactSensitiveText(input)
	if strings.Contains(output, "sk-proj") || strings.Contains(output, "<mark>") {
		t.Fatalf("marked secret was not redacted: %s", output)
	}
}
