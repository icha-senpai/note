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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/icha-senpai/note/kernel/apiroutes"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
)

const (
	maxAPICallResponseBytes = 5 * 1024 * 1024
	maxAPICallResponseChars = 50000
)

var APICatalogTool = &Tool{
	Name:         "api_catalog",
	Title:        "Scribli API catalog",
	Description:  "Discover Scribli's authenticated local HTTP/API route catalog, including inferred effect metadata. Optional filters: family (e.g. /api/block), keyword, limit (default 100, max 500). Use this before api_route or api_call when MCP lacks a specialized tool.",
	ReadOnlyHint: true,
	InputSchema: ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"family":  {Type: "string", Description: "Optional route family filter such as /api/block, /api/av, /api/setting, /ws/plugin, or /mcp."},
			"keyword": {Type: "string", Description: "Optional case-insensitive search across method, path, family, and handler."},
			"limit":   {Type: "number", Description: "Maximum number of routes to return. Defaults to 100; maximum 500."},
		},
	},
	EffectScope:   EffectScopeLocal,
	ActionEffects: effectMap(ToolEffects{LocalRead: true}, ""),
	Handler:       apiCatalogHandler,
}

var APIRouteTool = &Tool{
	Name:         "api_route",
	Title:        "Scribli API route detail",
	Description:  "Get exact local API route details, including method, family, handler, inferred effects, and api_call guidance. Provide path and optionally method.",
	ReadOnlyHint: true,
	InputSchema: ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"method": {Type: "string", Description: "Optional HTTP method.", Enum: []string{"GET", "POST", "PUT", "DELETE", "PATCH"}},
			"path":   {Type: "string", Description: "Route path, e.g. /api/block/getBlockKramdown."},
		},
		Required: []string{"path"},
	},
	EffectScope:   EffectScopeLocal,
	ActionEffects: effectMap(ToolEffects{LocalRead: true}, ""),
	Handler:       apiRouteHandler,
}

var APICallTool = &Tool{
	Name:        "api_call",
	Title:       "Scribli API fallback call",
	Description: "Call one authenticated local Scribli API endpoint by path. Use only after checking api_catalog/api_route and only when no purpose-built MCP tool fits. method defaults to POST. path must be a local /api/... path. body may be an object/array/string; query is an optional object of query parameters. The API token is attached internally and never returned. Confirmation and snapshot behavior are inferred from the selected route.",
	InputSchema: ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"method": {Type: "string", Description: "HTTP method. Defaults to POST.", Enum: []string{"GET", "POST", "PUT", "DELETE", "PATCH"}},
			"path":   {Type: "string", Description: "Local API path, e.g. /api/block/getBlockKramdown. Full URLs are rejected."},
			"query": {
				Type:        "object",
				Description: "Optional query parameters.",
				Properties:  map[string]Property{},
			},
			"body": {
				Type:        "object",
				Description: "Optional JSON request body for POST/PUT/PATCH/DELETE.",
				AnyOf: []Property{
					{Type: "object", Properties: map[string]Property{}},
					{Type: "array"},
					{Type: "string"},
				},
			},
			"bodyText": {Type: "string", Description: "Optional raw JSON request body. Use body for ordinary object payloads."},
		},
		Required: []string{"path"},
	},
	EffectScope: EffectScopeMixed,
	ActionEffects: effectMap(
		ToolEffects{LocalRead: true, LocalWrite: true, DataEgress: true},
		"", "GET", "POST", "PUT", "DELETE", "PATCH",
	),
	EffectHandler: apiCallEffects,
	Handler:       apiCallHandler,
}

func init() {
	register(APICatalogTool)
	register(APIRouteTool)
	register(APICallTool)
}

func apiCatalogHandler(args map[string]any) (CallToolResult, error) {
	family, _ := args["family"].(string)
	keyword, _ := args["keyword"].(string)
	limit := intArg(args, "limit", 100)
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	family = strings.TrimSpace(family)
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	filtered := make([]apiroutes.Route, 0)
	familyCounts := map[string]int{}
	for _, route := range apiroutes.List() {
		familyCounts[route.Family]++
		if family != "" && route.Family != family {
			continue
		}
		if keyword != "" && !routeMatchesKeyword(route, keyword) {
			continue
		}
		filtered = append(filtered, route)
	}

	families := make([]map[string]any, 0, len(familyCounts))
	for name, count := range familyCounts {
		families = append(families, map[string]any{"family": name, "routes": count})
	}
	sort.Slice(families, func(i, j int) bool {
		return families[i]["family"].(string) < families[j]["family"].(string)
	})

	return jsonToolResult(map[string]any{
		"totalRoutes":    len(apiroutes.List()),
		"matchedRoutes":  len(filtered),
		"returnedRoutes": min(limit, len(filtered)),
		"families":       families,
		"routes":         filtered[:min(limit, len(filtered))],
	})
}

func apiRouteHandler(args map[string]any) (CallToolResult, error) {
	apiPath, _ := args["path"].(string)
	apiPath = strings.TrimSpace(apiPath)
	if apiPath == "" {
		return errorResult("api_route error: path is required"), nil
	}
	method, _ := args["method"].(string)
	method = strings.ToUpper(strings.TrimSpace(method))

	var routes []apiroutes.Route
	if method != "" {
		if route, ok := apiroutes.Find(method, apiPath); ok {
			routes = append(routes, route)
		}
	} else {
		routes = apiroutes.FindByPath(apiPath)
	}
	if len(routes) == 0 {
		return errorResult("api_route error: route not found in api_catalog"), nil
	}

	return jsonToolResult(map[string]any{
		"routes": routes,
		"apiCall": map[string]any{
			"tool":   "api_call",
			"path":   apiPath,
			"method": routes[0].Method,
		},
		"note": "Payload schemas are not generated yet; use docs/API.md, existing MCP tools, or frontend/API source examples when payload shape is unclear.",
	})
}

func apiCallHandler(args map[string]any) (CallToolResult, error) {
	method, _ := args["method"].(string)
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodPost
	}
	if !map[string]bool{http.MethodGet: true, http.MethodPost: true, http.MethodPut: true, http.MethodDelete: true, http.MethodPatch: true}[method] {
		return errorResult("api_call error: unsupported method " + method), nil
	}

	apiPath, _ := args["path"].(string)
	apiPath = strings.TrimSpace(apiPath)
	if err := validateAPIPath(method, apiPath); err != nil {
		return errorResult("api_call error: " + err.Error()), nil
	}

	requestURL, err := localAPIURL(apiPath, args["query"])
	if err != nil {
		return errorResult("api_call error: " + err.Error()), nil
	}

	body, err := apiCallBody(args)
	if err != nil {
		return errorResult("api_call error: " + err.Error()), nil
	}

	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return errorResult("api_call error: create request failed: " + err.Error()), nil
	}
	req.Header.Set("Accept", "application/json, text/plain;q=0.9")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if model.Conf == nil || model.Conf.Api == nil || model.Conf.Api.Token == "" {
		return errorResult("api_call error: configured API token is unavailable"), nil
	}
	req.Header.Set("Authorization", "Token "+model.Conf.Api.Token)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return errorResult("api_call error: request failed: " + err.Error()), nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxAPICallResponseBytes))
	if err != nil {
		return errorResult("api_call error: read response failed: " + err.Error()), nil
	}
	text := truncateRunes(string(respBody), maxAPICallResponseChars)
	if len(respBody) == maxAPICallResponseBytes {
		text += "\n\n[response truncated at byte limit]"
	}

	result := fmt.Sprintf("HTTP %d %s\n%s\n\n%s", resp.StatusCode, resp.Status, resp.Header.Get("Content-Type"), text)
	return CallToolResult{Content: []ContentItem{{Type: "text", Text: result}}, IsError: resp.StatusCode >= 400}, nil
}

func apiCallEffects(args map[string]any, action string) (ToolEffects, bool) {
	method, _ := args["method"].(string)
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodPost
	}
	apiPath, _ := args["path"].(string)
	apiPath = strings.TrimSpace(apiPath)
	if route, ok := apiroutes.Find(method, apiPath); ok {
		return routeEffectsToToolEffects(route.Effects), true
	}
	return ToolEffects{LocalRead: true, LocalWrite: true, DataEgress: true}, true
}

func routeEffectsToToolEffects(e apiroutes.RouteEffects) ToolEffects {
	return ToolEffects{
		LocalRead:       e.LocalRead,
		LocalWrite:      e.LocalWrite,
		LocalStateWrite: e.LocalStateWrite,
		DataEgress:      e.DataEgress,
		ExternalCost:    e.ExternalCost,
	}
}

func validateAPIPath(method, apiPath string) error {
	if apiPath == "" {
		return errors.New("path is required")
	}
	if strings.Contains(apiPath, "://") || !strings.HasPrefix(apiPath, "/") {
		return errors.New("path must be a local path, not a full URL")
	}
	if !strings.HasPrefix(apiPath, "/api/") {
		return errors.New("path must start with /api/")
	}
	if strings.Contains(apiPath, "#") {
		return errors.New("path must not include a fragment")
	}
	pathOnly := apiPath
	if idx := strings.IndexByte(pathOnly, '?'); idx >= 0 {
		pathOnly = pathOnly[:idx]
	}
	if len(apiroutes.List()) > 0 && !apiroutes.Match(method, pathOnly) {
		return fmt.Errorf("route %s %s is not in api_catalog", method, pathOnly)
	}
	return nil
}

func localAPIURL(apiPath string, rawQuery any) (string, error) {
	if util.ServerURL == nil {
		if util.ServerPort == "" || util.ServerPort == "0" {
			return "", errors.New("local server URL is not available yet")
		}
		util.ServerURL = &url.URL{Scheme: "http", Host: "127.0.0.1:" + util.ServerPort}
	}

	u := *util.ServerURL
	u.Path = apiPath
	u.RawQuery = ""
	if idx := strings.IndexByte(apiPath, '?'); idx >= 0 {
		u.Path = apiPath[:idx]
		u.RawQuery = apiPath[idx+1:]
	}
	query := u.Query()
	if q, ok := rawQuery.(map[string]any); ok {
		for k, v := range q {
			query.Set(k, fmt.Sprintf("%v", v))
		}
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func apiCallBody(args map[string]any) (io.Reader, error) {
	if bodyText, _ := args["bodyText"].(string); strings.TrimSpace(bodyText) != "" {
		if !json.Valid([]byte(bodyText)) {
			return nil, errors.New("bodyText must be valid JSON")
		}
		return strings.NewReader(bodyText), nil
	}
	body, ok := args["body"]
	if !ok || body == nil {
		return nil, nil
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, errors.New("marshal body failed: " + err.Error())
	}
	return bytes.NewReader(data), nil
}

func routeMatchesKeyword(route apiroutes.Route, keyword string) bool {
	haystack := strings.ToLower(route.Method + " " + route.Path + " " + route.Family + " " + route.Handler)
	return strings.Contains(haystack, keyword)
}

func intArg(args map[string]any, key string, fallback int) int {
	switch v := args[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return int(i)
		}
	}
	return fallback
}

func jsonToolResult(v any) (CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errorResult("marshal result failed: " + err.Error()), nil
	}
	return CallToolResult{Content: []ContentItem{{Type: "text", Text: string(data)}}}, nil
}

func errorResult(message string) CallToolResult {
	return CallToolResult{Content: []ContentItem{{Type: "text", Text: message}}, IsError: true}
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "\n\n[response truncated]"
}
