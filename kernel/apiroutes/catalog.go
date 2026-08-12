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

package apiroutes

import (
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type Route struct {
	Method      string       `json:"method"`
	Path        string       `json:"path"`
	Family      string       `json:"family"`
	Handler     string       `json:"handler,omitempty"`
	EffectScope string       `json:"effectScope"`
	Effects     RouteEffects `json:"effects"`
	Risk        string       `json:"risk"`
}

type RouteEffects struct {
	LocalRead       bool `json:"localRead,omitempty"`
	LocalWrite      bool `json:"localWrite,omitempty"`
	LocalStateWrite bool `json:"localStateWrite,omitempty"`
	DataEgress      bool `json:"dataEgress,omitempty"`
	ExternalCost    bool `json:"externalCost,omitempty"`
}

var (
	mu     sync.RWMutex
	routes []Route
)

func SetFromGinRoutes(infos gin.RoutesInfo) {
	next := make([]Route, 0, len(infos))
	for _, info := range infos {
		if !isCatalogPath(info.Path) {
			continue
		}
		route := Route{
			Method:  info.Method,
			Path:    info.Path,
			Family:  routeFamily(info.Path),
			Handler: info.Handler,
		}
		route.EffectScope, route.Effects, route.Risk = inferEffects(route.Method, route.Path, route.Family)
		next = append(next, route)
	}
	sort.Slice(next, func(i, j int) bool {
		if next[i].Path == next[j].Path {
			return next[i].Method < next[j].Method
		}
		return next[i].Path < next[j].Path
	})

	mu.Lock()
	routes = next
	mu.Unlock()
}

func List() []Route {
	mu.RLock()
	defer mu.RUnlock()

	ret := make([]Route, len(routes))
	copy(ret, routes)
	return ret
}

func Match(method, path string) bool {
	_, ok := Find(method, path)
	return ok
}

func Find(method, path string) (Route, bool) {
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	if idx := strings.IndexByte(path, '?'); idx >= 0 {
		path = path[:idx]
	}
	for _, route := range List() {
		if route.Method != method && route.Method != "ANY" {
			continue
		}
		if route.Path == path || patternMatches(route.Path, path) {
			return route, true
		}
	}
	return Route{}, false
}

func FindByPath(path string) []Route {
	path = strings.TrimSpace(path)
	if idx := strings.IndexByte(path, '?'); idx >= 0 {
		path = path[:idx]
	}

	ret := []Route{}
	for _, route := range List() {
		if route.Path == path || patternMatches(route.Path, path) {
			ret = append(ret, route)
		}
	}
	return ret
}

func isCatalogPath(path string) bool {
	return strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/ws/") ||
		strings.HasPrefix(path, "/es/") ||
		path == "/mcp" ||
		strings.HasPrefix(path, "/plugin/private/")
}

func routeFamily(path string) string {
	if path == "/mcp" {
		return "/mcp"
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return path
	}
	return "/" + parts[0] + "/" + parts[1]
}

func patternMatches(pattern, path string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range patternParts {
		if strings.HasPrefix(part, "*") {
			return len(pathParts) >= i
		}
		if i >= len(pathParts) {
			return false
		}
		if strings.HasPrefix(part, ":") {
			continue
		}
		if part != pathParts[i] {
			return false
		}
	}
	return len(patternParts) == len(pathParts)
}

func inferEffects(method, path, family string) (string, RouteEffects, string) {
	method = strings.ToUpper(strings.TrimSpace(method))
	lowerPath := strings.ToLower(path)
	name := routeName(path)

	effects := RouteEffects{LocalRead: true}
	scope := "local"
	risk := "read"

	if routeHasNetworkEffect(lowerPath, family) {
		effects.DataEgress = true
		scope = "mixed"
		risk = "network"
	}
	if strings.HasPrefix(family, "/api/ai") {
		effects.DataEgress = true
		effects.ExternalCost = true
		scope = "mixed"
		risk = "external-cost"
	}

	if routeIsRead(method, name, lowerPath) {
		return scope, effects, risk
	}

	switch {
	case family == "/api/sync":
		if strings.Contains(lowerPath, "download") || strings.Contains(lowerPath, "perform") {
			effects.LocalWrite = true
		} else {
			effects.LocalStateWrite = true
		}
		effects.DataEgress = true
		scope = "mixed"
		risk = "sync"
	case family == "/api/export":
		effects.LocalStateWrite = true
		risk = "export"
	case family == "/api/repo":
		if strings.Contains(lowerPath, "checkout") || strings.Contains(lowerPath, "rollback") || strings.Contains(lowerPath, "resetrepo") || strings.Contains(lowerPath, "importrepokey") {
			effects.LocalWrite = true
			risk = "write"
		} else {
			effects.LocalStateWrite = true
			risk = "state"
		}
	case routeWritesWorkspace(family):
		effects.LocalWrite = true
		risk = "write"
	default:
		effects.LocalStateWrite = true
		risk = "state"
	}
	if effects.DataEgress {
		scope = "mixed"
	}
	return scope, effects, risk
}

func routeName(path string) string {
	path = strings.Trim(path, "/")
	if idx := strings.LastIndexByte(path, '/'); idx >= 0 {
		return strings.ToLower(path[idx+1:])
	}
	return strings.ToLower(path)
}

func routeIsRead(method, name, lowerPath string) bool {
	if method == "GET" && !strings.Contains(lowerPath, "/proxy") {
		return true
	}
	readPrefixes := []string{
		"get", "list", "ls", "search", "render", "query", "diff", "check",
		"version", "currenttime", "bootprogress", "stat", "mcpstatus",
	}
	for _, prefix := range readPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func routeHasNetworkEffect(lowerPath, family string) bool {
	if strings.HasPrefix(family, "/api/network") || strings.HasPrefix(family, "/ws/network") || strings.HasPrefix(family, "/es/network") {
		return true
	}
	if family == "/api/sync" {
		return strings.Contains(lowerPath, "sync") || strings.Contains(lowerPath, "cloud") || strings.Contains(lowerPath, "upload") || strings.Contains(lowerPath, "download")
	}
	return false
}

func routeWritesWorkspace(family string) bool {
	switch family {
	case "/api/archive", "/api/asset", "/api/attr", "/api/av", "/api/block", "/api/bookmark",
		"/api/file", "/api/filetree", "/api/format", "/api/history", "/api/import",
		"/api/notebook", "/api/ref", "/api/riff", "/api/search", "/api/snippet",
		"/api/tag", "/api/template", "/api/transactions":
		return true
	default:
		return false
	}
}
