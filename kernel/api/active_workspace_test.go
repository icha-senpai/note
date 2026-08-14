// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/conf"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
)

func TestActiveWorkspaceAPIBridgesExecutableBlockAndCanvas(t *testing.T) {
	gin.SetMode(gin.TestMode)

	oldConf := model.Conf
	oldDataDir := util.DataDir
	t.Cleanup(func() {
		model.Conf = oldConf
		util.DataDir = oldDataDir
	})

	model.Conf = &model.AppConf{Api: &conf.API{Token: "test-token"}}
	util.DataDir = filepath.Join(t.TempDir(), "data")

	server := gin.New()
	ServeAPI(server)

	execResponse := postJSON(t, server, "/api/executableBlock/call", map[string]any{
		"action": "chart",
		"chart": map[string]any{
			"xAxis":  map[string]any{"type": "category", "data": []any{"A"}},
			"yAxis":  map[string]any{"type": "value"},
			"series": []any{map[string]any{"type": "bar", "data": []any{1}}},
		},
	})
	if execResponse.Code != 0 {
		t.Fatalf("executableBlock response code = %d, msg=%s", execResponse.Code, execResponse.Msg)
	}
	execData := execResponse.Data.(map[string]any)
	execStructured := execData["structuredContent"].(map[string]any)
	if execStructured["renderer"] != "echarts" {
		t.Fatalf("unexpected executable structured content: %#v", execStructured)
	}

	canvasID := "20260814122000-cdefghi"
	canvasResponse := postJSON(t, server, "/api/canvas/call", map[string]any{
		"action": "create",
		"id":     canvasID,
		"title":  "API Board",
	})
	if canvasResponse.Code != 0 {
		t.Fatalf("canvas response code = %d, msg=%s", canvasResponse.Code, canvasResponse.Msg)
	}
	canvasData := canvasResponse.Data.(map[string]any)
	canvasStructured := canvasData["structuredContent"].(map[string]any)
	if canvasStructured["id"] != canvasID {
		t.Fatalf("unexpected canvas structured content: %#v", canvasStructured)
	}
}

type apiResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func postJSON(t *testing.T, server http.Handler, path string, body map[string]any) apiResult {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token test-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("%s status = %d, body=%s", path, resp.Code, resp.Body.String())
	}

	var result apiResult
	if err = json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}
