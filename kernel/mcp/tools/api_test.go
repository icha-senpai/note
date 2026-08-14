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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/apiroutes"
	"github.com/icha-senpai/note/kernel/conf"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
)

func TestAPICatalogFiltersLiveRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.POST("/api/block/getBlockKramdown", func(c *gin.Context) {})
	server.POST("/api/block/updateBlock", func(c *gin.Context) {})
	server.GET("/ws/broadcast", func(c *gin.Context) {})
	apiroutes.SetFromGinRoutes(server.Routes())

	result, err := apiCatalogHandler(map[string]any{"family": "/api/block", "keyword": "getBlockKramdown"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || len(result.Content) != 1 {
		t.Fatalf("unexpected catalog result: %#v", result)
	}

	var payload struct {
		MatchedRoutes int `json:"matchedRoutes"`
		Families      []struct {
			Family string `json:"family"`
			Routes int    `json:"routes"`
		} `json:"families"`
		Routes []struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		} `json:"routes"`
	}
	if err = json.Unmarshal([]byte(result.Content[0].Text), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.MatchedRoutes != 1 || len(payload.Routes) != 1 {
		t.Fatalf("unexpected matched routes: %#v", payload)
	}
	if len(payload.Families) != 1 || payload.Families[0].Family != "/api/block" || payload.Families[0].Routes != 1 {
		t.Fatalf("unexpected filtered families: %#v", payload.Families)
	}
	if payload.Routes[0].Method != http.MethodPost || payload.Routes[0].Path != "/api/block/getBlockKramdown" {
		t.Fatalf("unexpected route: %#v", payload.Routes[0])
	}
}

func TestAPICallAllowsSelfSignedLoopbackHTTPS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routeServer := gin.New()
	routeServer.POST("/api/test/tls", func(c *gin.Context) {})
	apiroutes.SetFromGinRoutes(routeServer.Routes())

	oldConf := model.Conf
	oldURL := util.ServerURL
	t.Cleanup(func() {
		model.Conf = oldConf
		util.ServerURL = oldURL
	})

	model.Conf = &model.AppConf{Api: &conf.API{Token: "secret-token"}}
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Token secret-token" {
			t.Fatalf("unexpected Authorization header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tls":true}`))
	}))
	defer tlsServer.Close()

	parsedURL, err := url.Parse(tlsServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	util.ServerURL = parsedURL

	result, err := apiCallHandler(map[string]any{
		"method": "POST",
		"path":   "/api/test/tls",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || len(result.Content) != 1 {
		t.Fatalf("unexpected api_call TLS result: %#v", result)
	}
	if !strings.Contains(result.Content[0].Text, `{"tls":true}`) {
		t.Fatalf("missing TLS response body: %s", result.Content[0].Text)
	}
}

func TestAPIRouteReturnsEffectsAndCallGuidance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.POST("/api/block/updateBlock", func(c *gin.Context) {})
	apiroutes.SetFromGinRoutes(server.Routes())

	result, err := apiRouteHandler(map[string]any{"method": "POST", "path": "/api/block/updateBlock"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || len(result.Content) != 1 {
		t.Fatalf("unexpected route result: %#v", result)
	}

	var payload struct {
		Routes []struct {
			EffectScope string `json:"effectScope"`
			Effects     struct {
				LocalWrite bool `json:"localWrite"`
			} `json:"effects"`
			Risk string `json:"risk"`
		} `json:"routes"`
		APICall map[string]any `json:"apiCall"`
	}
	if err = json.Unmarshal([]byte(result.Content[0].Text), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Routes) != 1 || payload.Routes[0].EffectScope != "local" || !payload.Routes[0].Effects.LocalWrite || payload.Routes[0].Risk != "write" {
		t.Fatalf("unexpected route effects: %#v", payload)
	}
	if payload.APICall["tool"] != "api_call" || payload.APICall["path"] != "/api/block/updateBlock" {
		t.Fatalf("unexpected api_call guidance: %#v", payload.APICall)
	}
}

func TestAPICallEffectsFollowRouteMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.POST("/api/block/getBlockKramdown", func(c *gin.Context) {})
	server.POST("/api/block/updateBlock", func(c *gin.Context) {})
	server.POST("/api/export/exportMd", func(c *gin.Context) {})
	server.POST("/api/ai/chatGPT", func(c *gin.Context) {})
	apiroutes.SetFromGinRoutes(server.Routes())

	effects, ok := apiCallEffects(map[string]any{"method": "POST", "path": "/api/block/getBlockKramdown"}, "")
	if !ok || !effects.LocalRead || effects.LocalWrite || effects.LocalStateWrite || effects.DataEgress {
		t.Fatalf("unexpected read effects: %#v, ok=%v", effects, ok)
	}

	effects, ok = apiCallEffects(map[string]any{"method": "POST", "path": "/api/block/updateBlock"}, "")
	if !ok || !effects.LocalWrite || effects.DataEgress {
		t.Fatalf("unexpected write effects: %#v, ok=%v", effects, ok)
	}

	effects, ok = apiCallEffects(map[string]any{"method": "POST", "path": "/api/export/exportMd"}, "")
	if !ok || !effects.LocalStateWrite || effects.LocalWrite {
		t.Fatalf("unexpected export effects: %#v, ok=%v", effects, ok)
	}

	effects, ok = apiCallEffects(map[string]any{"method": "POST", "path": "/api/ai/chatGPT"}, "")
	if !ok || !effects.DataEgress || !effects.ExternalCost {
		t.Fatalf("unexpected AI effects: %#v, ok=%v", effects, ok)
	}
}

func TestAPICallUsesInternalTokenAndLocalPathOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routeServer := gin.New()
	routeServer.POST("/api/test/echo", func(c *gin.Context) {})
	apiroutes.SetFromGinRoutes(routeServer.Routes())

	oldConf := model.Conf
	oldURL := util.ServerURL
	t.Cleanup(func() {
		model.Conf = oldConf
		util.ServerURL = oldURL
	})

	model.Conf = &model.AppConf{Api: &conf.API{Token: "secret-token"}}
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/test/echo" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Token secret-token" {
			t.Fatalf("unexpected Authorization header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer httpServer.Close()

	parsedURL, err := url.Parse(httpServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	util.ServerURL = parsedURL

	result, err := apiCallHandler(map[string]any{
		"method": "POST",
		"path":   "/api/test/echo",
		"body":   map[string]any{"hello": "world"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || len(result.Content) != 1 {
		t.Fatalf("unexpected api_call result: %#v", result)
	}
	if !strings.Contains(result.Content[0].Text, `{"ok":true}`) {
		t.Fatalf("missing response body: %s", result.Content[0].Text)
	}

	result, err = apiCallHandler(map[string]any{
		"method": "GET",
		"path":   "https://example.com/api/test/echo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError || !strings.Contains(result.Content[0].Text, "path must be a local path") {
		t.Fatalf("full URL should be rejected: %#v", result)
	}
}
