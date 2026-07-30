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

package agent

import (
	"embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed models.json
var modelsJSON embed.FS

type modelMetaEntry struct {
	ContextLength int `json:"contextLength"`
}

var (
	modelContextOnce  sync.Once
	modelContextLimit map[string]int
)

func loadModelContext() map[string]int {
	modelContextOnce.Do(func() {
		raw, err := modelsJSON.ReadFile("models.json")
		if err != nil {
			return
		}
		var decoded map[string]json.RawMessage
		if err := json.Unmarshal(raw, &decoded); err != nil {
			return
		}
		result := make(map[string]int, len(decoded))
		for name, value := range decoded {
			if name == "_meta" {
				continue
			}
			var entry modelMetaEntry
			if err := json.Unmarshal(value, &entry); err != nil {
				continue
			}
			if entry.ContextLength > 0 {

				result[strings.ToLower(name)] = entry.ContextLength
			}
		}
		modelContextLimit = result
	})
	return modelContextLimit
}

func GetModelContextLimit(model string) int {
	if model == "" {
		return 0
	}
	table := loadModelContext()
	if table == nil {
		return 0
	}
	lower := strings.ToLower(model)
	if limit, ok := table[lower]; ok {
		return limit
	}
	if idx := strings.LastIndexByte(lower, '/'); idx >= 0 {
		if limit, ok := table[lower[idx+1:]]; ok {
			return limit
		}
	}
	return 0
}
