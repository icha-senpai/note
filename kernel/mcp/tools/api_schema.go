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
	"net/http"
	"strings"

	"github.com/icha-senpai/note/kernel/apiroutes"
)

func apiRouteDetail(route apiroutes.Route) map[string]any {
	return map[string]any{
		"method":        route.Method,
		"path":          route.Path,
		"family":        route.Family,
		"handler":       route.Handler,
		"effectScope":   route.EffectScope,
		"effects":       route.Effects,
		"risk":          route.Risk,
		"requestSchema": apiRequestSchema(route),
	}
}

func apiRequestSchema(route apiroutes.Route) map[string]any {
	if schema, ok := exactAPIRequestSchema(route.Path); ok {
		return schema
	}

	body := inferredAPIRequestBodySchema(route)
	confidence := "pattern"
	source := "route naming heuristic"
	if body == nil {
		body = objectSchema(nil, map[string]any{}, true)
		confidence = "unknown"
		source = "generic JSON object fallback"
	}

	return map[string]any{
		"contentType": "application/json",
		"body":        body,
		"confidence":  confidence,
		"source":      source,
		"notes":       []string{"Inferred schema; confirm against docs/API.md or source if the route rejects the payload."},
	}
}

func exactAPIRequestSchema(path string) (map[string]any, bool) {
	switch path {
	case "/api/block/getBlockKramdown":
		return exactBodySchema(
			objectSchema([]string{"id"}, map[string]any{
				"id":       stringSchema("Block ID."),
				"mode":     enumSchema("Kramdown return mode.", "md", "textmark"),
				"notebook": stringSchema("Optional encrypted notebook ID when reading from an encrypted notebook."),
			}, false),
			map[string]any{"id": "20201225220954-dlgzk1o", "mode": "md"},
			"docs/API.md and kernel/api/block.go",
		), true
	case "/api/block/getBlockKramdowns":
		return exactBodySchema(
			objectSchema([]string{"ids"}, map[string]any{
				"ids":      arraySchema(stringSchema("Block ID."), "Block IDs."),
				"mode":     enumSchema("Kramdown return mode.", "md", "textmark"),
				"notebook": stringSchema("Optional encrypted notebook ID when reading from an encrypted notebook."),
			}, false),
			map[string]any{"ids": []string{"20201225220954-dlgzk1o"}, "mode": "md"},
			"kernel/api/block.go",
		), true
	case "/api/block/updateBlock":
		return exactBodySchema(
			objectSchema([]string{"id", "dataType", "data"}, map[string]any{
				"id":       stringSchema("Block ID to update."),
				"dataType": enumSchema("Input data format.", "markdown", "dom"),
				"data":     stringSchema("Replacement Markdown or DOM content."),
			}, false),
			map[string]any{"id": "20211230161520-querkps", "dataType": "markdown", "data": "Updated text"},
			"docs/API.md and kernel/api/block_op.go",
		), true
	case "/api/attr/getBlockAttrs":
		return exactBodySchema(
			objectSchema([]string{"id"}, map[string]any{
				"id": stringSchema("Block ID."),
			}, false),
			map[string]any{"id": "20210912214605-uhi5gco"},
			"docs/API.md and kernel/api/attr.go",
		), true
	case "/api/attr/batchGetBlockAttrs":
		return exactBodySchema(
			objectSchema([]string{"ids"}, map[string]any{
				"ids": arraySchema(stringSchema("Block ID."), "Block IDs."),
			}, false),
			map[string]any{"ids": []string{"20210912214605-uhi5gco"}},
			"kernel/api/attr.go",
		), true
	case "/api/attr/setBlockAttrs":
		return exactBodySchema(
			objectSchema([]string{"id", "attrs"}, map[string]any{
				"id":    stringSchema("Block ID."),
				"attrs": stringMapSchema("Block attributes. Custom attributes must use the custom- prefix."),
			}, false),
			map[string]any{"id": "20210912214605-uhi5gco", "attrs": map[string]string{"custom-attr1": "value"}},
			"docs/API.md and kernel/api/attr.go",
		), true
	case "/api/file/readDir":
		return exactBodySchema(
			objectSchema([]string{"path"}, map[string]any{
				"path": stringSchema("Directory path under the workspace, e.g. /data/<notebook>/<doc>."),
			}, false),
			map[string]any{"path": "/data/20210808180117-6v0mkxr/20200923234011-ieuun1p"},
			"docs/API.md and kernel/api/file.go",
		), true
	case "/api/filetree/createDocWithMd":
		return exactBodySchema(
			objectSchema([]string{"notebook", "path", "markdown"}, map[string]any{
				"notebook":     stringSchema("Notebook ID."),
				"path":         stringSchema("Document hpath. Start with / and separate levels with /."),
				"markdown":     stringSchema("GFM Markdown content."),
				"tags":         stringSchema("Optional tag string."),
				"parentID":     stringSchema("Optional parent document ID."),
				"id":           stringSchema("Optional explicit document ID."),
				"withMath":     boolSchema("Whether to import Markdown with math handling."),
				"clippingHref": stringSchema("Optional source URL for clipped content."),
			}, false),
			map[string]any{"notebook": "20210817205410-2kvfpfn", "path": "/foo/bar", "markdown": ""},
			"docs/API.md and kernel/api/filetree.go",
		), true
	}
	return nil, false
}

func inferredAPIRequestBodySchema(route apiroutes.Route) map[string]any {
	if route.Method == http.MethodGet || strings.HasPrefix(route.Path, "/ws/") || strings.HasPrefix(route.Path, "/es/") {
		return objectSchema(nil, map[string]any{}, true)
	}

	name := strings.ToLower(route.Path)
	routeBase := apiRouteBaseName(route.Path)
	properties := map[string]any{}
	required := []string{}

	if strings.Contains(routeBase, "byid") || strings.HasSuffix(routeBase, "id") || strings.Contains(name, "block") || strings.Contains(name, "doc") {
		properties["id"] = stringSchema("ID used by this route.")
		required = append(required, "id")
	}
	if strings.Contains(routeBase, "ids") || strings.Contains(routeBase, "batch") {
		properties["ids"] = arraySchema(stringSchema("ID."), "IDs used by this route.")
		required = append(required, "ids")
		delete(properties, "id")
		required = removeString(required, "id")
	}
	if strings.Contains(routeBase, "path") || strings.Contains(routeBase, "dir") || strings.Contains(routeBase, "file") {
		properties["path"] = stringSchema("Workspace-relative path used by this route.")
		if !containsString(required, "path") {
			required = append(required, "path")
		}
	}
	if strings.Contains(routeBase, "notebook") {
		properties["notebook"] = stringSchema("Notebook ID.")
		if !containsString(required, "notebook") {
			required = append(required, "notebook")
		}
	}
	if len(properties) == 0 {
		return nil
	}
	return objectSchema(required, properties, true)
}

func apiCallArguments(route apiroutes.Route, apiPath string) map[string]any {
	args := map[string]any{
		"method": route.Method,
		"path":   apiPath,
	}
	schema := apiRequestSchema(route)
	if example, ok := schema["exampleBody"].(map[string]any); ok && route.Method != http.MethodGet {
		args["body"] = example
	}
	return args
}

func exactBodySchema(body map[string]any, example map[string]any, source string) map[string]any {
	return map[string]any{
		"contentType": "application/json",
		"body":        body,
		"exampleBody": example,
		"confidence":  "exact",
		"source":      source,
	}
}

func objectSchema(required []string, properties map[string]any, additionalProperties bool) map[string]any {
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": additionalProperties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func boolSchema(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}

func enumSchema(description string, values ...string) map[string]any {
	enum := make([]any, 0, len(values))
	for _, value := range values {
		enum = append(enum, value)
	}
	return map[string]any{"type": "string", "description": description, "enum": enum}
}

func arraySchema(items map[string]any, description string) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": items}
}

func stringMapSchema(description string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"description":          description,
		"additionalProperties": map[string]any{"type": "string"},
	}
}

func apiRouteBaseName(path string) string {
	path = strings.Trim(path, "/")
	if idx := strings.LastIndexByte(path, '/'); idx >= 0 {
		return strings.ToLower(path[idx+1:])
	}
	return strings.ToLower(path)
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func removeString(values []string, needle string) []string {
	ret := values[:0]
	for _, value := range values {
		if value != needle {
			ret = append(ret, value)
		}
	}
	return ret
}
