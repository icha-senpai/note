// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/mcp/tools"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/gulu"
)

func executableBlockCall(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}
	result, err := tools.ExecutableBlockTool.Handler(arg)
	writeToolCallResult(ret, result, err)
}

func canvasCall(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}
	result, err := tools.CanvasTool.Handler(arg)
	writeToolCallResult(ret, result, err)
}

func writeToolCallResult(ret *gulu.Result, result tools.CallToolResult, err error) {
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}
	if result.IsError {
		ret.Code = -1
	}
	if len(result.Content) > 0 && result.Content[0].Text != "" {
		ret.Msg = result.Content[0].Text
	}
	ret.Data = map[string]any{
		"content":           result.Content,
		"structuredContent": result.StructuredContent,
		"isError":           result.IsError,
	}
}
