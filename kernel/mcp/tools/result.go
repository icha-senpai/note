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

package tools

func structuredOutputSchema() *ToolSchema {
	return &ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"action":  {Type: "string", Description: "Tool action that was performed"},
			"status":  {Type: "string", Description: "Result status, usually ok or empty"},
			"message": {Type: "string", Description: "Human-readable result summary"},
		},
	}
}

func structuredTextResult(message string, structured map[string]any) CallToolResult {
	if structured == nil {
		structured = map[string]any{}
	}
	if _, ok := structured["message"]; !ok {
		structured["message"] = message
	}
	if _, ok := structured["status"]; !ok {
		structured["status"] = "ok"
	}
	return CallToolResult{
		Content:              []ContentItem{{Type: "text", Text: message}},
		StructuredContent:    structured,
		StructuredContentSet: true,
	}
}

// sliceRunes returns text[offset:offset+maxChars] in runes. nextOffset is -1 when the end is reached.
func sliceRunes(text string, offset, maxChars int) (sliced string, total int, nextOffset int) {
	runes := []rune(text)
	total = len(runes)
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := total
	if maxChars > 0 && offset+maxChars < end {
		end = offset + maxChars
	}
	sliced = string(runes[offset:end])
	nextOffset = -1
	if end < total {
		nextOffset = end
	}
	return
}
