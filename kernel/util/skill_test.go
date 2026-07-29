// SiYuan - Refactor your thinking
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

package util

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func touchSkillMd(t *testing.T, parts ...string) {
	t.Helper()
	p := filepath.Join(parts...)
	if err := os.MkdirAll(p, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}
	skillMd := filepath.Join(p, "SKILL.md")
	if err := os.WriteFile(skillMd, []byte("---\nname: x\n---\nbody"), 0644); err != nil {
		t.Fatalf("write %s: %v", skillMd, err)
	}
}

func sorted(ss []string) []string {
	out := append([]string(nil), ss...)
	sort.Strings(out)
	return out
}

func toSlashSet(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = filepath.ToSlash(s)
	}
	sort.Strings(out)
	return out
}

func TestFindSkillDirsRootSkill(t *testing.T) {
	root := t.TempDir()
	touchSkillMd(t, root)
	got := findSkillDirs(root)
	want := []string{"."}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("findSkillDirs(root skill) = %v, want %v", got, want)
	}
}

func TestFindSkillDirsWrappedSingle(t *testing.T) {
	root := t.TempDir()
	touchSkillMd(t, root, "WeChatReading")
	got := toSlashSet(findSkillDirs(root))
	want := []string{"WeChatReading"}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("findSkillDirs(wrapped single) = %v, want %v", got, want)
	}
}

func TestFindSkillDirsWrappedCollection(t *testing.T) {
	root := t.TempDir()
	touchSkillMd(t, root, "myrepo", "skills", "foo")
	touchSkillMd(t, root, "myrepo", "skills", "bar")
	got := toSlashSet(findSkillDirs(root))
	want := []string{"myrepo/skills/bar", "myrepo/skills/foo"}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("findSkillDirs(wrapped collection) = %v, want %v", got, want)
	}
}

func TestFindSkillDirsCollectionNoWrap(t *testing.T) {
	root := t.TempDir()
	touchSkillMd(t, root, "skills", "baz")
	got := toSlashSet(findSkillDirs(root))
	want := []string{"skills/baz"}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("findSkillDirs(collection no wrap) = %v, want %v", got, want)
	}
}

func TestFindSkillDirsDoesNotDescendIntoSkillInternals(t *testing.T) {
	root := t.TempDir()

	skillDir := filepath.Join(root, "skill5", "SKILL.md")
	if err := os.MkdirAll(filepath.Join(root, "skill5", "references"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "skill5", "scripts"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillDir, []byte("---\nname: x\n---"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "skill5", "references", "a.md"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "skill5", "scripts", "b.py"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	got := toSlashSet(findSkillDirs(root))
	want := []string{"skill5"}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("should not descend into skill internals: got %v, want %v", got, want)
	}
}

func TestFindSkillDirsNoneFound(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "some", "dir"), 0755); err != nil {
		t.Fatal(err)
	}
	if got := findSkillDirs(root); len(got) != 0 {
		t.Fatalf("findSkillDirs(no skill) = %v, want empty", got)
	}
}

func TestFindSkillDirsSkipsVCSDirs(t *testing.T) {
	root := t.TempDir()
	touchSkillMd(t, root, "real-skill")

	if err := os.MkdirAll(filepath.Join(root, ".github", "workflows"), 0755); err != nil {
		t.Fatal(err)
	}
	got := toSlashSet(findSkillDirs(root))
	want := []string{"real-skill"}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("findSkillDirs(skip vcs) = %v, want %v", got, want)
	}
}

type urlCase struct {
	name       string
	in         string
	wantURL    string
	wantIsZip  bool
	wantBranch string
}

func TestNormalizeSkillURL(t *testing.T) {
	cases := []urlCase{

		{name: "shorthand", in: "Tencent/WeChatReading", wantURL: "https://codeload.github.com/Tencent/WeChatReading/zip/refs/heads/main", wantIsZip: true, wantBranch: "main"},

		{name: "npx command", in: "npx skills add Tencent/WeChatReading -g", wantURL: "https://codeload.github.com/Tencent/WeChatReading/zip/refs/heads/main", wantIsZip: true, wantBranch: "main"},

		{name: "npx scoped", in: "npx skills@latest add foo/bar", wantURL: "https://codeload.github.com/foo/bar/zip/refs/heads/main", wantIsZip: true, wantBranch: "main"},

		{name: "full repo", in: "https://github.com/Tencent/WeChatReading", wantURL: "https://codeload.github.com/Tencent/WeChatReading/zip/refs/heads/main", wantIsZip: true, wantBranch: "main"},

		{name: "tree branch", in: "https://github.com/foo/bar/tree/dev", wantURL: "https://codeload.github.com/foo/bar/zip/refs/heads/dev", wantIsZip: true, wantBranch: "dev"},

		{name: "tree branch path", in: "https://github.com/foo/bar/tree/dev/skills/x", wantURL: "https://codeload.github.com/foo/bar/zip/refs/heads/dev", wantIsZip: true, wantBranch: "dev"},
		// commit/<sha> → codeload zip/<sha>
		{name: "commit sha", in: "https://github.com/foo/bar/commit/abc123", wantURL: "https://codeload.github.com/foo/bar/zip/abc123", wantIsZip: true},

		{name: "blob file", in: "https://github.com/foo/bar/blob/main/SKILL.md", wantURL: "https://raw.githubusercontent.com/foo/bar/main/SKILL.md", wantIsZip: false},

		{name: "raw direct", in: "https://raw.githubusercontent.com/foo/bar/main/skills/x/SKILL.md", wantURL: "https://raw.githubusercontent.com/foo/bar/main/skills/x/SKILL.md", wantIsZip: false},

		{name: "release zip", in: "https://github.com/foo/bar/releases/download/v1.0/skill.zip", wantURL: "https://github.com/foo/bar/releases/download/v1.0/skill.zip", wantIsZip: true},

		{name: "release non-zip", in: "https://github.com/foo/bar/releases/download/v1.0/skill.tar.gz", wantURL: "https://github.com/foo/bar/releases/download/v1.0/skill.tar.gz", wantIsZip: false},

		{name: "third-party", in: "https://example.com/skill.zip", wantURL: "https://example.com/skill.zip"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := normalizeSkillURL(c.in)
			if err != nil {
				t.Fatalf("normalizeSkillURL(%q) error: %v", c.in, err)
			}
			if c.wantURL != "" && got.downloadURL != c.wantURL {
				t.Errorf("downloadURL = %q, want %q", got.downloadURL, c.wantURL)
			}
			if got.isZip != c.wantIsZip {
				t.Errorf("isZip = %v, want %v", got.isZip, c.wantIsZip)
			}
			if c.wantBranch != "" && got.branch != c.wantBranch {
				t.Errorf("branch = %q, want %q", got.branch, c.wantBranch)
			}
		})
	}
}

func TestNormalizeSkillURLOwnerRepoBoundary(t *testing.T) {
	bad := []string{
		"example.com",
		"a/b/c",
		"https://foo/bar",
		"/abs/path",
	}
	for _, in := range bad {

		got, err := normalizeSkillURL(in)
		if err != nil {
			continue
		}

		if in == "example.com" && strings.HasPrefix(got.downloadURL, "https://codeload.github.com/") {
			t.Errorf("input %q should not map to codeload, got %s", in, got.downloadURL)
		}
	}
}

func TestNormalizeSkillURLOwnerRepoValid(t *testing.T) {
	good := []string{"a/b", "Tencent/WeChatReading", "user-1/my.repo.v2"}
	for _, in := range good {
		got, err := normalizeSkillURL(in)
		if err != nil {
			t.Errorf("normalizeSkillURL(%q) unexpected error: %v", in, err)
			continue
		}
		if !got.isZip || got.branch != "main" {
			t.Errorf("normalizeSkillURL(%q) = %+v, want codeload main zip", in, got)
		}
	}
}

func TestParseSkillFrontmatter(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		wantName string
		wantDesc string
		wantBody string
	}{
		{
			name:     "standard",
			text:     "---\nname: my-skill\ndescription: does X\n---\nbody line",
			wantName: "my-skill",
			wantDesc: "does X",
			wantBody: "body line",
		},
		{
			name:     "no frontmatter",
			text:     "just body",
			wantName: "",
			wantDesc: "",
			wantBody: "just body",
		},
		{
			name:     "extra fields ignored",
			text:     "---\nname: a\nversion: \"1.0\"\nauthor: me\ndescription: d\n---\nb",
			wantName: "a",
			wantDesc: "d",
			wantBody: "b",
		},
		{
			name:     "crlf line endings",
			text:     "---\r\nname: cr\r\ndescription: crlf\r\n---\r\nbody",
			wantName: "cr",
			wantDesc: "crlf",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fm, body := parseSkillFrontmatter(c.text)
			if fm["name"] != c.wantName {
				t.Errorf("name = %q, want %q", fm["name"], c.wantName)
			}
			if fm["description"] != c.wantDesc {
				t.Errorf("description = %q, want %q", fm["description"], c.wantDesc)
			}
			if c.wantBody != "" && body != c.wantBody {
				t.Errorf("body = %q, want %q", body, c.wantBody)
			}
		})
	}
}

func TestFirstLine(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"single", "single"},
		{"first\nsecond", "first"},
		{"first\r\nsecond", "first"},
		{"   trimmed   ", "trimmed"},
	}
	for _, c := range cases {
		if got := firstLine(c.in); got != c.want {
			t.Errorf("firstLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFirstLineTruncation(t *testing.T) {
	long := make([]rune, 300)
	for i := range long {
		long[i] = 'x'
	}
	got := firstLine(string(long))
	if len([]rune(got)) != 203 { // 200 + "..."
		t.Errorf("firstLine(long) len = %d, want 203", len([]rune(got)))
	}
}
