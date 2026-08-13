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

package conf

import "testing"

func TestNormalizeMCPServerIDs(t *testing.T) {
	ai := NewAI()
	ai.MCP.Servers = []MCPServer{
		{Name: "missing"},
		{ID: "duplicate", Name: "first"},
		{ID: "duplicate", Name: "second"},
	}
	ai.Normalize()
	seen := map[string]bool{}
	for _, server := range ai.MCP.Servers {
		if server.ID == "" || seen[server.ID] {
			t.Fatalf("unexpected MCP server ID: %#v", ai.MCP.Servers)
		}
		seen[server.ID] = true
	}
}

func TestNormalizeMCPExposurePolicy(t *testing.T) {
	ai := &AI{MCP: &MCP{ExposurePolicy: &CapabilityPolicy{
		Default: "DENY",
		Overrides: map[string]string{
			" native/backend/search ": "ALLOW",
			"":                        "allow",
			"native/backend/delete":   "maybe",
		},
	}}}
	ai.Normalize()

	if ai.MCP.ExposurePolicy.Default != CapabilityPolicyDeny {
		t.Fatalf("default policy = %q, want deny", ai.MCP.ExposurePolicy.Default)
	}
	if !ai.MCP.ExposurePolicy.Allows("native/backend/search") {
		t.Fatal("explicit allow override was not honored")
	}
	if ai.MCP.ExposurePolicy.Allows("native/backend/delete") {
		t.Fatal("invalid override should have been discarded")
	}
	if ai.MCP.ExposurePolicy.Allows("native/backend/list") {
		t.Fatal("default deny was not honored")
	}
}

func TestNilMCPExposurePolicyDefaultsToAllow(t *testing.T) {
	var policy *CapabilityPolicy
	if !policy.Allows("native/backend/read") {
		t.Fatal("nil policy should default to allow")
	}

	ai := &AI{MCP: &MCP{}}
	ai.Normalize()
	if ai.MCP.ExposurePolicy == nil || !ai.MCP.ExposurePolicy.Allows("native/backend/read") {
		t.Fatalf("normalize did not install default allow policy: %#v", ai.MCP.ExposurePolicy)
	}
}
