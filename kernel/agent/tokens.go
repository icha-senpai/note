// Scribli - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
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

package agent

import (
	"encoding/json"
	"sync"

	"github.com/pkoukk/tiktoken-go"
	loader "github.com/pkoukk/tiktoken-go-loader"
	"github.com/sashabaranov/go-openai"
	tools "github.com/siyuan-note/siyuan/kernel/mcp/tools"
)

type tokenCounter struct {
	enc *tiktoken.Tiktoken
}

var (
	tokenCounterOnce sync.Once
	globalCounter    *tokenCounter
	tokenCounterErr  error
)

func getTokenCounter(modelName string) (*tokenCounter, error) {
	tokenCounterOnce.Do(func() {
		tiktoken.SetBpeLoader(loader.NewOfflineLoader())
		enc, err := tiktoken.EncodingForModel(modelName)
		if err != nil {

			enc, err = tiktoken.GetEncoding("cl100k_base")
			if err != nil {
				tokenCounterErr = err
				return
			}
		}
		globalCounter = &tokenCounter{enc: enc}
	})
	return globalCounter, tokenCounterErr
}

func (c *tokenCounter) count(text string) int {
	if c == nil || c.enc == nil {
		return estimateTokensByChars(text)
	}
	return len(c.enc.Encode(text, nil, nil))
}

func estimateTokensByChars(text string) int {
	if text == "" {
		return 0
	}
	cjk := 0
	other := 0
	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF || r >= 0x3040 && r <= 0x30FF || r >= 0xAC00 && r <= 0xD7AF {
			cjk++
		} else {
			other++
		}
	}
	return cjk*2/3 + other/4
}

func toolSource(name string) string {
	if t := tools.GetTool(name); t != nil {
		if t.Source != "" {
			return t.Source
		}
	}

	if len(name) > 8 && name[:8] == "plugin__" {
		return "plugin"
	}

	return "mcp"
}

func computeTokenBreakdown(counter *tokenCounter, messages []openai.ChatCompletionMessage, tools []openai.Tool, skillsTokens, realPromptTokens int) map[string]int {
	breakdown := map[string]int{
		"system":         0,
		"skills":         skillsTokens,
		"messages":       0,
		"nativeToolsDef": 0,
		"pluginToolsDef": 0,
		"mcpToolsDef":    0,
		"nativeTool":     0,
		"pluginTool":     0,
		"mcpTool":        0,
		"other":          0,
	}

	systemTotal := 0

	idToToolName := map[string]string{}
	const perMessageOverhead = 4
	for _, msg := range messages {
		switch msg.Role {
		case openai.ChatMessageRoleSystem:
			systemTotal += counter.count(msg.Content) + perMessageOverhead
		case openai.ChatMessageRoleUser:
			breakdown["messages"] += counter.count(msg.Content) + perMessageOverhead
		case openai.ChatMessageRoleAssistant:
			breakdown["messages"] += counter.count(msg.Content) + perMessageOverhead

			if msg.ReasoningContent != "" {
				breakdown["messages"] += counter.count(msg.ReasoningContent)
			}
			for _, tc := range msg.ToolCalls {
				name := tc.Function.Name
				idToToolName[tc.ID] = name

				callTokens := counter.count(name) + counter.count(tc.Function.Arguments) + 7
				switch toolSource(name) {
				case "native":
					breakdown["nativeTool"] += callTokens
				case "plugin":
					breakdown["pluginTool"] += callTokens
				default:
					breakdown["mcpTool"] += callTokens
				}
			}
		case openai.ChatMessageRoleTool:

			name := idToToolName[msg.ToolCallID]
			resultTokens := counter.count(msg.Content) + perMessageOverhead
			switch toolSource(name) {
			case "native":
				breakdown["nativeTool"] += resultTokens
			case "plugin":
				breakdown["pluginTool"] += resultTokens
			default:
				breakdown["mcpTool"] += resultTokens
			}
		}
	}
	breakdown["system"] = max(systemTotal-skillsTokens, 0)

	const perToolDefOverhead = 10
	for _, t := range tools {
		if t.Function == nil {
			continue
		}
		defText := t.Function.Name + " " + t.Function.Description
		if paramsJSON, err := json.Marshal(t.Function.Parameters); err == nil {
			defText += " " + string(paramsJSON)
		}
		defTokens := counter.count(defText) + perToolDefOverhead
		switch toolSource(t.Function.Name) {
		case "native":
			breakdown["nativeToolsDef"] += defTokens
		case "plugin":
			breakdown["pluginToolsDef"] += defTokens
		default:
			breakdown["mcpToolsDef"] += defTokens
		}
	}

	if len(messages) > 0 {
		breakdown["messages"] += 3
	}

	estimated := 0
	for k, v := range breakdown {
		if k == "other" {
			continue
		}
		estimated += v
	}
	if realPromptTokens > estimated {
		breakdown["other"] = realPromptTokens - estimated
	} else if estimated > realPromptTokens && realPromptTokens > 0 {
		scale := float64(realPromptTokens) / float64(estimated)
		allocated := 0

		keys := []string{"system", "skills", "messages",
			"nativeToolsDef", "pluginToolsDef", "mcpToolsDef",
			"nativeTool", "pluginTool", "mcpTool"}
		for _, k := range keys {
			scaled := int(float64(breakdown[k]) * scale)
			breakdown[k] = scaled
			allocated += scaled
		}

		breakdown["other"] = max(realPromptTokens-allocated, 0)
	}
	return breakdown
}
