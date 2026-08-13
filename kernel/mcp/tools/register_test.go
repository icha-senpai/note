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

import "testing"

func TestGetAllToolsSorted(t *testing.T) {
	allTools := GetAllTools()
	for i := 1; i < len(allTools); i++ {
		if allTools[i-1].Name > allTools[i].Name {
			t.Fatalf("tools are not sorted: %q appears before %q", allTools[i-1].Name, allTools[i].Name)
		}
	}
}

func TestNativeToolsDeclareEffectMetadata(t *testing.T) {
	validScopes := map[string]bool{
		EffectScopeLocal:    true,
		EffectScopeExternal: true,
		EffectScopeMixed:    true,
		EffectScopeUnknown:  true,
	}

	for _, tool := range GetAllTools() {
		if tool == nil || (tool.Source != "" && tool.Source != "native") {
			continue
		}
		if !validScopes[tool.EffectScope] {
			t.Errorf("%s must declare a valid effect scope, got %q", tool.Name, tool.EffectScope)
		}
		if len(tool.ActionEffects) == 0 {
			t.Errorf("%s must declare action effects", tool.Name)
			continue
		}

		actionProp, hasAction := tool.InputSchema.Properties["action"]
		if !hasAction || len(actionProp.Enum) == 0 {
			continue
		}

		actionSet := map[string]bool{}
		for _, action := range actionProp.Enum {
			actionSet[action] = true
			if _, ok := tool.EffectsFor(action); !ok {
				t.Errorf("%s action %q must declare effects", tool.Name, action)
			}
		}

		for action := range tool.ActionEffects {
			if action == "" {
				continue
			}
			if !actionSet[action] && tool.Name != "frontend" {
				t.Errorf("%s declares effects for unknown action %q", tool.Name, action)
			}
		}
	}
}

func TestObserveRegistry(t *testing.T) {
	const name = "registry_observer_test"
	RemoveTool(name)

	var events []*Tool
	stop := ObserveRegistry(func(changedName string, tool *Tool) {
		if changedName == name {
			events = append(events, tool)
		}
	})
	t.Cleanup(func() {
		stop()
		RemoveTool(name)
	})

	tool := &Tool{Name: name, InputSchema: ToolSchema{Type: "object"}}
	if err := SetTool(name, tool); err != nil {
		t.Fatal(err)
	}
	RemoveToolIf(name, &Tool{Name: name})
	RemoveToolIf(name, tool)

	if len(events) != 2 || events[0] != tool || events[1] != nil {
		t.Fatalf("unexpected registry events: %#v", events)
	}
}

func TestSetToolKeepsExistingToolWhenSchemaIsInvalid(t *testing.T) {
	const name = "registry_validation_test"
	original := &Tool{Name: name, InputSchema: ToolSchema{Type: "object"}}
	if err := SetTool(name, original); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		RemoveTool(name)
	})

	invalid := &Tool{Name: name, InputSchema: ToolSchema{Raw: map[string]any{}}}
	if err := SetTool(name, invalid); err == nil {
		t.Fatal("expected invalid schema")
	}
	if actual := LookupTool(name); actual != original {
		t.Fatalf("invalid replacement changed registry entry: %#v", actual)
	}
}
