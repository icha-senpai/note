// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tools

import (
	"regexp"
	"strings"
)

var sensitiveTextRedactors = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bsk-(?:proj-)?[A-Za-z0-9_-]{20,}`),
	regexp.MustCompile(`(?i)\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{20,}`),
	regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
	regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`),
	regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{16,}`),
	regexp.MustCompile(`(?i)\b(api[_-]?key|access[_-]?token|refresh[_-]?token|auth[_-]?token|token|secret|password|passwd|authorization)\b\s*[:=]\s*["']?[^"',;\s]+`),
}

func redactSensitiveText(text string) string {
	redacted := redactSensitiveText0(text)
	if redacted != text {
		return redacted
	}
	unmarked := strings.ReplaceAll(text, "<mark>", "")
	unmarked = strings.ReplaceAll(unmarked, "</mark>", "")
	redacted = redactSensitiveText0(unmarked)
	if redacted != unmarked {
		return redacted
	}
	return text
}

func redactSensitiveText0(text string) string {
	for _, redactor := range sensitiveTextRedactors {
		text = redactor.ReplaceAllStringFunc(text, func(match string) string {
			if idx := strings.IndexAny(match, "=:"); idx >= 0 {
				return match[:idx+1] + " [REDACTED]"
			}
			if strings.HasPrefix(strings.ToLower(match), "bearer ") {
				return "Bearer [REDACTED]"
			}
			return "[REDACTED]"
		})
	}
	return text
}
