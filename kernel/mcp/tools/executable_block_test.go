// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tools

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/icha-senpai/note/kernel/apiroutes"
	"github.com/icha-senpai/note/kernel/conf"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
)

func TestExecutableBlockRunJSReturnsStructuredOutput(t *testing.T) {
	result, err := executableBlockHandler(map[string]any{
		"action": "run_js",
		"code":   "console.log('adding', input.x); input.x + 2",
		"input":  map[string]any{"x": 40},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || !result.HasStructuredContent() {
		t.Fatalf("unexpected executable result: %#v", result)
	}
	content := result.StructuredContent.(map[string]any)
	if content["output"] != int64(42) && content["output"] != float64(42) {
		t.Fatalf("output = %#v, want 42", content["output"])
	}
	logs := content["logs"].([]string)
	if len(logs) != 1 || logs[0] != "adding 40" {
		t.Fatalf("unexpected logs: %#v", logs)
	}
}

func TestExecutableBlockRunJSTimesOut(t *testing.T) {
	result, err := executableBlockHandler(map[string]any{
		"action":    "run_js",
		"code":      "while (true) {}",
		"timeoutMs": float64(10),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("execution failures should be structured tool results, not transport errors: %#v", result)
	}
	content := result.StructuredContent.(map[string]any)
	if !strings.Contains(content["error"].(string), "execution timed out") {
		t.Fatalf("unexpected timeout error: %#v", content["error"])
	}
}

func TestExecutableBlockChartReturnsEChartsMarkdown(t *testing.T) {
	result, err := executableBlockHandler(map[string]any{
		"action": "chart",
		"chart": map[string]any{
			"title": map[string]any{"text": "Notes"},
			"xAxis": map[string]any{"type": "category", "data": []any{"A", "B"}},
			"yAxis": map[string]any{"type": "value"},
			"series": []any{
				map[string]any{"type": "bar", "data": []any{1, 2}},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || !result.HasStructuredContent() {
		t.Fatalf("unexpected chart result: %#v", result)
	}
	content := result.StructuredContent.(map[string]any)
	if content["renderer"] != "echarts" {
		t.Fatalf("renderer = %#v, want echarts", content["renderer"])
	}
	markdown := content["markdown"].(string)
	if !strings.HasPrefix(markdown, "```echarts\n") || !strings.HasSuffix(markdown, "\n```") {
		t.Fatalf("unexpected chart markdown: %s", markdown)
	}
}

func TestExecutableBlockRunAPIDelegatesToGuardedAPICall(t *testing.T) {
	oldConf := model.Conf
	oldServerURL := util.ServerURL
	t.Cleanup(func() {
		model.Conf = oldConf
		util.ServerURL = oldServerURL
	})

	model.Conf = &model.AppConf{Api: &conf.API{Token: "test-token"}}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token test-token" {
			t.Fatalf("Authorization header = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"","data":{"version":"test"}}`))
	}))
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	util.ServerURL = parsedURL
	method := http.MethodPost
	apiPath := "/api/block/getBlockKramdown"
	if routes := apiroutes.List(); len(routes) > 0 {
		method = routes[0].Method
		apiPath = routes[0].Path
	}

	result, err := executableBlockHandler(map[string]any{
		"action": "run_api",
		"method": method,
		"path":   apiPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || !result.HasStructuredContent() {
		t.Fatalf("unexpected API executable result: %#v", result)
	}
	content := result.StructuredContent.(map[string]any)
	if content["action"] != "run_api" || content["sourceAction"] != "api_call" {
		t.Fatalf("unexpected API executable content: %#v", content)
	}
	if !strings.Contains(content["body"].(string), `"version":"test"`) {
		t.Fatalf("unexpected API executable body: %#v", content["body"])
	}
}
