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

func TestRedactSensitiveTextAdditionalProviders(t *testing.T) {
	slackToken := strings.Join([]string{
		"xoxb",
		"123456789012",
		"123456789012",
		"abcdefghijklmnopqrstuvwxyz",
	}, "-")
	input := strings.Join([]string{
		`slack = "` + slackToken + `"`,
		`gitlab = "glpat-aaaaaaaaaaaaaaaaaaaaaaaa"`,
		`npm = "npm_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
		`google = "AIzaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`,
		"-----BEGIN PRIVATE KEY-----\nabc123\n-----END PRIVATE KEY-----",
	}, "\n")
	output := redactSensitiveText(input)
	for _, leaked := range []string{"xoxb-", "glpat-", "npm_", "AIza", "PRIVATE KEY"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("provider secret leaked %q in %s", leaked, output)
		}
	}
}

func TestMaybeRedactSensitiveTextDefaultsOn(t *testing.T) {
	input := `token = "sk-proj-dddddddddddddddddddddddddddddddd"`
	if output := maybeRedactSensitiveText(map[string]any{}, input); strings.Contains(output, "sk-proj") {
		t.Fatalf("default redaction leaked token: %s", output)
	}
	if output := maybeRedactSensitiveText(map[string]any{"redactSecrets": false}, input); !strings.Contains(output, "sk-proj") {
		t.Fatalf("explicit raw output was redacted: %s", output)
	}
}
