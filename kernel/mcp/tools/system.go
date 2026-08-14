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

import (
	"fmt"
	"time"

	"github.com/icha-senpai/note/kernel/util"
)

var SystemTool = &Tool{
	Name:        "system",
	Description: "System info. Actions: version(), current_time(), workspace().",
	InputSchema: ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"action": {Type: "string", Description: "Operation", Enum: []string{"version", "current_time", "workspace"}},
		},
		Required: []string{"action"},
	},
	OutputSchema:  structuredOutputSchema(),
	EffectScope:   EffectScopeLocal,
	ActionEffects: effectMap(ToolEffects{LocalRead: true}, "version", "current_time", "workspace"),
	ReadOnlyHint:  true,
	Handler:       systemHandler,
}

func init() {
	register(SystemTool)
}

func systemHandler(args map[string]any) (CallToolResult, error) {
	action, _ := args["action"].(string)
	switch action {
	case "version":
		return systemVersion(args)
	case "current_time":
		return systemCurrentTime(args)
	case "workspace":
		return systemWorkspace(args)
	}
	return CallToolResult{
		Content: []ContentItem{{Type: "text", Text: "unknown action '" + action + "', expected one of: [version, current_time, workspace]"}},
		IsError: true,
	}, nil
}

func systemVersion(args map[string]any) (CallToolResult, error) {
	return structuredTextResult(util.Ver, map[string]any{
		"action":  "version",
		"version": util.Ver,
	}), nil
}

func systemCurrentTime(args map[string]any) (CallToolResult, error) {
	ms := util.CurrentTimeMillis()
	t := time.UnixMilli(ms)
	formatted := t.Format(time.RFC3339)
	return structuredTextResult(formatted, map[string]any{
		"action":       "current_time",
		"unixMillis":   ms,
		"rfc3339":      formatted,
		"timezoneName": t.Location().String(),
	}), nil
}

func systemWorkspace(args map[string]any) (CallToolResult, error) {
	message := fmt.Sprintf("Workspace: %s\nVersion: %s\nContainer: %s", util.WorkspaceDir, util.Ver, util.Container)
	return structuredTextResult(message, map[string]any{
		"action":    "workspace",
		"workspace": util.WorkspaceDir,
		"version":   util.Ver,
		"container": util.Container,
	}), nil
}
