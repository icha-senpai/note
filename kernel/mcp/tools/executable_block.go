// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
)

const (
	maxExecutableBlockCodeBytes = 64 * 1024
	defaultExecutableTimeoutMs  = 1000
	maxExecutableTimeoutMs      = 5000
)

var ExecutableBlockTool = &Tool{
	Name:        "executable_block",
	Title:       "Executable block",
	Description: "Run or validate permission-gated local executable note blocks. Actions: validate_js(code), run_js(code, input?, timeoutMs?), run_sql(stmt), run_api(path, method?, query?, body?), chart(chart or chartJSON). JavaScript has no filesystem, network, process, fetch, or require access. Execution is manual and local.",
	InputSchema: ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"action":    {Type: "string", Description: "Operation", Enum: []string{"validate_js", "run_js", "run_sql", "run_api", "chart"}},
			"code":      {Type: "string", Description: "JavaScript source code. The last expression is returned."},
			"input":     {Type: "object", Description: "Optional JSON-compatible value exposed to the block as global input."},
			"stmt":      {Type: "string", Description: "Read-only SQL SELECT statement for run_sql."},
			"method":    {Type: "string", Description: "HTTP method for run_api. Defaults to POST.", Enum: []string{"GET", "POST", "PUT", "DELETE", "PATCH"}},
			"path":      {Type: "string", Description: "Local API path for run_api, e.g. /api/block/getBlockKramdown. Full URLs are rejected."},
			"query":     {Type: "object", Description: "Optional run_api query parameters."},
			"body":      {Type: "object", Description: "Optional JSON request body for run_api."},
			"bodyText":  {Type: "string", Description: "Optional raw JSON request body for run_api."},
			"chart":     {Type: "object", Description: "ECharts-compatible chart option object for chart."},
			"chartJSON": {Type: "string", Description: "ECharts-compatible chart option object as JSON for chart."},
			"timeoutMs": {Type: "number", Description: "Execution timeout in milliseconds. Defaults to 1000, maximum 5000."},
		},
		Required: []string{"action"},
	},
	OutputSchema: structuredOutputSchema(),
	EffectScope:  EffectScopeMixed,
	ActionEffects: mergeEffectMaps(
		effectMap(ToolEffects{LocalRead: true}, "validate_js", "run_sql"),
		effectMap(ToolEffects{LocalStateWrite: true}, "run_js"),
		effectMap(ToolEffects{LocalRead: true, LocalWrite: true, DataEgress: true}, "run_api"),
		effectMap(ToolEffects{}, "chart"),
	),
	EffectHandler: executableBlockEffects,
	Handler:       executableBlockHandler,
}

func init() {
	register(ExecutableBlockTool)
}

func executableBlockHandler(args map[string]any) (CallToolResult, error) {
	action, _ := args["action"].(string)
	switch action {
	case "validate_js":
		return executableBlockValidateJS(args)
	case "run_js":
		return executableBlockRunJS(args)
	case "run_sql":
		return executableBlockRunSQL(args)
	case "run_api":
		return executableBlockRunAPI(args)
	case "chart":
		return executableBlockChart(args)
	}
	return errorResult("unknown action '" + action + "', expected one of: [validate_js, run_js, run_sql, run_api, chart]"), nil
}

func executableBlockValidateJS(args map[string]any) (CallToolResult, error) {
	code, err := executableBlockCode(args)
	if err != nil {
		return errorResult("validate_js error: " + err.Error()), nil
	}
	rt := goja.New()
	if _, err = goja.Compile("executable-block.js", code, false); err != nil {
		return structuredTextResult("JavaScript validation failed", map[string]any{
			"action": "validate_js",
			"valid":  false,
			"error":  err.Error(),
		}), nil
	}
	return structuredTextResult("JavaScript validation passed", map[string]any{
		"action": "validate_js",
		"valid":  true,
		"engine": "goja",
		"global": executableBlockGlobalSummary(rt),
	}), nil
}

func executableBlockRunJS(args map[string]any) (CallToolResult, error) {
	code, err := executableBlockCode(args)
	if err != nil {
		return errorResult("run_js error: " + err.Error()), nil
	}

	timeoutMs := intArg(args, "timeoutMs", defaultExecutableTimeoutMs)
	if timeoutMs <= 0 {
		timeoutMs = defaultExecutableTimeoutMs
	}
	if timeoutMs > maxExecutableTimeoutMs {
		timeoutMs = maxExecutableTimeoutMs
	}

	rt := goja.New()
	logs := []string{}
	if err = installExecutableConsole(rt, &logs); err != nil {
		return errorResult("run_js error: " + err.Error()), nil
	}
	if input, ok := args["input"]; ok {
		if err = rt.Set("input", input); err != nil {
			return errorResult("run_js error: set input failed: " + err.Error()), nil
		}
	}

	timer := time.AfterFunc(time.Duration(timeoutMs)*time.Millisecond, func() {
		rt.Interrupt("execution timed out")
	})
	value, runErr := rt.RunString(code)
	timer.Stop()

	if runErr != nil {
		return structuredTextResult("JavaScript execution failed", map[string]any{
			"action":    "run_js",
			"engine":    "goja",
			"timeoutMs": timeoutMs,
			"logs":      logs,
			"error":     runErr.Error(),
		}), nil
	}

	output := executableBlockExport(value)
	text := executableBlockText(output)
	return structuredTextResult(text, map[string]any{
		"action":    "run_js",
		"engine":    "goja",
		"timeoutMs": timeoutMs,
		"logs":      logs,
		"output":    output,
	}), nil
}

func executableBlockRunSQL(args map[string]any) (CallToolResult, error) {
	stmt, _ := args["stmt"].(string)
	if strings.TrimSpace(stmt) == "" {
		return errorResult("run_sql error: stmt is required"), nil
	}
	result, err := sqlQuery(map[string]any{"action": "query", "stmt": stmt})
	if err != nil || !result.HasStructuredContent() {
		return result, err
	}
	if content, ok := result.StructuredContent.(map[string]any); ok {
		content["action"] = "run_sql"
		content["sourceAction"] = "query"
	}
	return result, nil
}

func executableBlockRunAPI(args map[string]any) (CallToolResult, error) {
	result, err := apiCallHandler(args)
	if err != nil || !result.HasStructuredContent() {
		return result, err
	}
	if content, ok := result.StructuredContent.(map[string]any); ok {
		content["action"] = "run_api"
		content["sourceAction"] = "api_call"
	}
	return result, nil
}

func executableBlockChart(args map[string]any) (CallToolResult, error) {
	chart, err := executableBlockChartSpec(args)
	if err != nil {
		return errorResult("chart error: " + err.Error()), nil
	}
	data, err := json.MarshalIndent(chart, "", "  ")
	if err != nil {
		return errorResult("chart error: marshal chart failed: " + err.Error()), nil
	}
	markdown := "```echarts\n" + string(data) + "\n```"
	return structuredTextResult(markdown, map[string]any{
		"action":   "chart",
		"renderer": "echarts",
		"chart":    chart,
		"markdown": markdown,
	}), nil
}

func executableBlockCode(args map[string]any) (string, error) {
	code, _ := args["code"].(string)
	code = strings.TrimSpace(code)
	if code == "" {
		return "", errors.New("code is required")
	}
	if len([]byte(code)) > maxExecutableBlockCodeBytes {
		return "", fmt.Errorf("code is too large; maximum is %d bytes", maxExecutableBlockCodeBytes)
	}
	return code, nil
}

func executableBlockChartSpec(args map[string]any) (map[string]any, error) {
	if raw, ok := args["chart"].(map[string]any); ok && raw != nil {
		return cloneJSONMap(raw), nil
	}
	chartJSON, _ := args["chartJSON"].(string)
	chartJSON = strings.TrimSpace(chartJSON)
	if chartJSON == "" {
		return nil, errors.New("chart or chartJSON is required")
	}
	var chart map[string]any
	if err := json.Unmarshal([]byte(chartJSON), &chart); err != nil {
		return nil, err
	}
	if chart == nil {
		return nil, errors.New("chartJSON must decode to an object")
	}
	return chart, nil
}

func installExecutableConsole(rt *goja.Runtime, logs *[]string) error {
	console := rt.NewObject()
	if err := console.Set("log", func(call goja.FunctionCall) goja.Value {
		parts := make([]string, 0, len(call.Arguments))
		for _, arg := range call.Arguments {
			parts = append(parts, arg.String())
		}
		*logs = append(*logs, strings.Join(parts, " "))
		return goja.Undefined()
	}); err != nil {
		return err
	}
	return rt.Set("console", console)
}

func executableBlockExport(value goja.Value) any {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil
	}
	exported := value.Export()
	if _, err := json.Marshal(exported); err == nil {
		return exported
	}
	return value.String()
}

func executableBlockText(output any) string {
	if output == nil {
		return "JavaScript executed with no returned value"
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", output)
	}
	return string(data)
}

func executableBlockGlobalSummary(_ *goja.Runtime) []string {
	return []string{"input", "console.log"}
}

func executableBlockEffects(args map[string]any, action string) (ToolEffects, bool) {
	if action == "run_api" {
		return apiCallEffects(args, action)
	}
	switch action {
	case "validate_js", "run_sql":
		return ToolEffects{LocalRead: true}, true
	case "run_js":
		return ToolEffects{LocalStateWrite: true}, true
	case "chart":
		return ToolEffects{}, true
	}
	return ToolEffects{}, false
}
