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
	"os"
	"path/filepath"
	"strings"

	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/filelock"
)

var ExportTool = &Tool{
	Name:        "export",
	Description: "Export operations. Actions: md(id), html(id), preview(id), docx(id, output file path or directory), sy(id) → .sy.zip, md-zip(id), data() → full workspace backup.",
	InputSchema: ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"action": {Type: "string", Description: "Operation", Enum: []string{"md", "html", "preview", "docx", "sy", "md-zip", "data"}},
			"id":     {Type: "string", Description: "Document block ID (for md, html, preview, docx, sy, md-zip)"},
			"output": {Type: "string", Description: "Output .docx file path or output directory (required for docx, optional for others)"},
			"redactSecrets": {
				Type:        "boolean",
				Description: "Redact likely credentials/tokens from inline md/html/preview responses. Defaults to true.",
			},
		},
		Required: []string{"action"},
	},
	OutputSchema: structuredOutputSchema(),
	EffectScope:  EffectScopeLocal,
	ActionEffects: effectMap(
		ToolEffects{LocalRead: true, LocalStateWrite: true},
		"md", "html", "preview", "docx", "sy", "md-zip", "data",
	),
	Handler: exportHandler,
}

func init() {
	register(ExportTool)
}

func exportHandler(args map[string]any) (CallToolResult, error) {
	action, _ := args["action"].(string)
	switch action {
	case "md":
		return exportMd(args)
	case "html":
		return exportHtml(args)
	case "preview":
		return exportPreview(args)
	case "docx":
		return exportDocx(args)
	case "sy":
		return exportSy(args)
	case "md-zip":
		return exportMdZip(args)
	case "data":
		return exportData(args)
	}
	return CallToolResult{
		Content: []ContentItem{{Type: "text", Text: "unknown action '" + action + "', expected one of: [md, html, preview, docx, sy, md-zip, data]"}},
		IsError: true,
	}, nil
}

func exportMd(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}

	hPath, content := model.ExportMarkdownContent(id, 4, 0, true, false, false, false, false)
	if content == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "export failed or empty"}}, IsError: true}, nil
	}
	content = maybeRedactSensitiveText(args, content)

	text := fmt.Sprintf("# %s\n\n%s", hPath, content)
	return structuredTextResult(text, map[string]any{
		"action":   "md",
		"id":       id,
		"hPath":    hPath,
		"markdown": content,
	}), nil
}

func exportHtml(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	_, dom, _ := model.ExportHTML(id, "", false, false, false)
	if dom == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "export failed or empty"}}, IsError: true}, nil
	}
	dom = maybeRedactSensitiveText(args, dom)
	return structuredTextResult(dom, map[string]any{
		"action": "html",
		"id":     id,
		"html":   dom,
	}), nil
}

func exportPreview(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	html := model.ExportPreview(id, false)
	if html == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "export preview failed or empty"}}, IsError: true}, nil
	}
	html = maybeRedactSensitiveText(args, html)
	return structuredTextResult(html, map[string]any{
		"action": "preview",
		"id":     id,
		"html":   html,
	}), nil
}

func exportDocx(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	output, _ := args["output"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	output = strings.TrimSpace(output)
	if output == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "output file path is required for docx"}}, IsError: true}, nil
	}
	fullPath, err := exportDocxToOutput(id, output)
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "export docx failed: " + err.Error()}}, IsError: true}, nil
	}
	return structuredTextResult(fmt.Sprintf("exported docx to: %s", fullPath), map[string]any{
		"action": "docx",
		"id":     id,
		"output": output,
		"path":   fullPath,
	}), nil
}

func exportDocxToOutput(id, output string) (string, error) {
	if strings.EqualFold(filepath.Ext(output), ".docx") {
		target := util.GetUniqueFilename(output)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", err
		}
		generated, err := model.ExportDocx(id, filepath.Dir(target), false, false)
		if err != nil {
			return "", err
		}
		if filepath.Clean(generated) == filepath.Clean(target) {
			return generated, nil
		}
		if err = os.Rename(generated, target); err == nil {
			return target, nil
		}
		if err = filelock.Copy(generated, target); err != nil {
			return "", err
		}
		_ = os.Remove(generated)
		return target, nil
	}
	return model.ExportDocx(id, output, false, false)
}

func exportSy(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	_, zipPath := model.ExportPandocConvertZip([]string{id}, "", ".sy")
	if zipPath == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "export sy failed"}}, IsError: true}, nil
	}
	return structuredTextResult(fmt.Sprintf("exported sy.zip to: %s", zipPath), map[string]any{
		"action": "sy",
		"id":     id,
		"path":   zipPath,
	}), nil
}

func exportMdZip(args map[string]any) (CallToolResult, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "id is required"}}, IsError: true}, nil
	}
	_, zipPath := model.ExportPandocConvertZip([]string{id}, "", ".md")
	if zipPath == "" {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "export md-zip failed"}}, IsError: true}, nil
	}
	return structuredTextResult(fmt.Sprintf("exported md zip to: %s", zipPath), map[string]any{
		"action": "md-zip",
		"id":     id,
		"path":   zipPath,
	}), nil
}

func exportData(args map[string]any) (CallToolResult, error) {
	zipPath, err := model.ExportData()
	if err != nil {
		return CallToolResult{Content: []ContentItem{{Type: "text", Text: "export data failed: " + err.Error()}}, IsError: true}, nil
	}
	return structuredTextResult(fmt.Sprintf("exported data backup to: %s", zipPath), map[string]any{
		"action": "data",
		"path":   zipPath,
	}), nil
}
