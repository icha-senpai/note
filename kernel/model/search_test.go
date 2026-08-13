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

package model

import (
	gosql "database/sql"
	"slices"
	"strings"
	"testing"

	"github.com/icha-senpai/note/kernel/conf"
)

func TestValidEmbedBlockIDs(t *testing.T) {
	firstID := "20260721120000-block01"
	secondID := "20260721120001-block02"
	thirdID := "20260721120002-block03"
	ids := validEmbedBlockIDs([]string{firstID, "invalid", firstID, secondID, thirdID}, 2)
	if !slices.Equal(ids, []string{firstID, secondID}) {
		t.Fatalf("embedded block IDs should preserve order, deduplicate, and enforce the limit: %v", ids)
	}
}

func TestIsValidSearchBoxPath(t *testing.T) {
	validBox := "20210808180117-6v0mkxr"

	validCases := []struct {
		name string
		box  string
		path string
	}{
		{"notebook scope only", validBox, ""},
		{"slash only", validBox, "/"},
		{"specific document", validBox, "/20210808180117-6v0mkxr.sy"},
		{"subtree directory scope", validBox, "/20210808180117-6v0mkxr"},
		{"full child document path", validBox, "/20210808180117-6v0mkxr/20210808180530-a1b2c3d.sy"},
	}
	for _, tc := range validCases {
		t.Run("valid/"+tc.name, func(t *testing.T) {
			if !IsValidSearchBoxPath(tc.box, tc.path) {
				t.Fatalf("expected valid: box=%q path=%q", tc.box, tc.path)
			}
		})
	}

	invalidCases := []struct {
		name string
		box  string
		path string
	}{

		{
			"SQL injection UNION projection",
			validBox,
			"/x%') UNION SELECT id,parent_id FROM blocks WHERE path='/hidden.sy' -- ",
		},
		{"single quote string break", validBox, "/doc'secret.sy"},
		{"leading percent", validBox, "/%abc"},
		{"comment marker", validBox, "/doc -- "},
		{"invalid short numeric box", "123", ""},
		{"invalid uppercase box", "20210808180117-6V0MKXR", ""},
		{"invalid empty box", "", "/20210808180117-6v0mkxr.sy"},
		{"path missing leading slash", validBox, "20210808180117-6v0mkxr.sy"},
		{"invalid path segment", validBox, "/notanid.sy"},
		{"invalid middle path segment", validBox, "/20210808180117-6v0mkxr/notanid.sy"},
	}
	for _, tc := range invalidCases {
		t.Run("invalid/"+tc.name, func(t *testing.T) {
			if IsValidSearchBoxPath(tc.box, tc.path) {
				t.Fatalf("expected invalid: box=%q path=%q", tc.box, tc.path)
			}
		})
	}
}

func TestBuildBoxesPathFiltersArgCount(t *testing.T) {
	boxes := []string{"20210808180117-6v0mkxr", "20210808180117-a1b2c3d"}
	clause, args := buildBoxesFilter(boxes)
	if countPlaceholder(clause) != len(args) {
		t.Fatalf("box filter placeholder/arg mismatch: %q vs %d args", clause, len(args))
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 box args, got %d", len(args))
	}

	paths := []string{"/20210808180117-6v0mkxr", "/20210808180117-a1b2c3d/20210808180530-e5f6g7h.sy"}
	clause, args = buildPathsFilter(paths)
	if countPlaceholder(clause) != len(args) {
		t.Fatalf("path filter placeholder/arg mismatch: %q vs %d args", clause, len(args))
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 path args, got %d", len(args))
	}
	for i, a := range args {
		s, ok := a.(string)
		if !ok || s != paths[i]+"%" {
			t.Fatalf("path arg %d should be %q%%, got %v", i, paths[i], a)
		}
	}
}

func TestBuildRootIDExclusionFilter(t *testing.T) {
	rootIDs := []string{"20260716120000-abcdefg", "20260716120001-hijklmn"}
	clause, args := buildRootIDExclusionFilter(rootIDs, "b.")
	if " AND b.root_id NOT IN (?, ?)" != clause {
		t.Fatalf("unexpected root ID exclusion filter: %q", clause)
	}
	if countPlaceholder(clause) != len(args) || len(rootIDs) != len(args) {
		t.Fatalf("root ID filter placeholder/arg mismatch: %q vs %d args", clause, len(args))
	}
	for i, arg := range args {
		if rootIDs[i] != arg {
			t.Fatalf("root ID arg %d should be %q, got %v", i, rootIDs[i], arg)
		}
	}

	clause, args = buildRootIDExclusionFilter(nil)
	if "" != clause || 0 != len(args) {
		t.Fatalf("empty root IDs should not generate a filter: %q, %v", clause, args)
	}
}

func TestNormalizeBoxName(t *testing.T) {
	name := "  notebook/name\x00  "
	if normalized := normalizeBoxName(name); "notebookname" != normalized {
		t.Fatalf("unexpected normalized notebook name: %q", normalized)
	}
}

func countPlaceholder(s string) (n int) {
	for i := 0; i < len(s); i++ {
		if s[i] == '?' {
			n++
		}
	}
	return
}

func TestBuildRefUsedOrderBy(t *testing.T) {
	newestID := "20260714120000-newest1"
	olderID := "20260714110000-older01"
	invalidID := "invalid-id' OR 1=1 --"
	orderBy := buildRefUsedOrderBy(map[string]int64{
		olderID:   100,
		newestID:  200,
		invalidID: 300,
	})

	newestPos := strings.Index(orderBy, newestID)
	olderPos := strings.Index(orderBy, olderID)
	if 0 > newestPos || 0 > olderPos || newestPos >= olderPos {
		t.Fatalf("newer referenced blocks should sort before older referenced blocks: %q", orderBy)
	}
	if strings.Contains(orderBy, invalidID) {
		t.Fatalf("order clause should not contain invalid block IDs: %q", orderBy)
	}
	if !strings.HasSuffix(orderBy, "END ASC, ") {
		t.Fatalf("order clause has the wrong format: %q", orderBy)
	}
}

func TestBuildRefUsedOrderByEmpty(t *testing.T) {
	if orderBy := buildRefUsedOrderBy(nil); "" != orderBy {
		t.Fatalf("empty records should not generate an order clause: %q", orderBy)
	}
}

func TestBuildOrderByPrioritizesExactDocumentAndHeading(t *testing.T) {
	setSearchCaseSensitive(t, true)
	orderBy := buildOrderBy("math", 0, 0)
	assertOrderBySequence(t, orderBy,
		"name = 'math'",
		"instr(',' || alias || ',', ',math,') > 0",
		"content = 'math' AND type = 'd'",
		"content LIKE '%math%' AND type = 'd'",
		"content = 'math' AND type = 'h'",
		"content LIKE '%math%' AND type = 'h'",
		"sort ASC",
	)

	orderBy = buildOrderBy("math", 0, 7)
	assertOrderBySequence(t, orderBy,
		"content = 'math' AND type = 'd'",
		"content = 'math' AND type = 'h'",
		"rank",
	)

	orderBy = buildOrderBy("math", 0, 6)
	if strings.Contains(orderBy, "content = 'math'") {
		t.Fatalf("ascending relevance should not put exact matches first: %q", orderBy)
	}
}

func TestBuildOrderByEscapesKeyword(t *testing.T) {
	setSearchCaseSensitive(t, true)
	orderBy := buildOrderBy("O'Reilly", 0, 7)
	if !strings.Contains(orderBy, "content = 'O''Reilly'") {
		t.Fatalf("keyword was not escaped correctly in order clause: %q", orderBy)
	}
}

func TestBuildExactSearchOrderConditionEscapesKeyword(t *testing.T) {
	setSearchCaseSensitive(t, true)
	condition := buildExactSearchOrderCondition("content", "O'Reilly%_\\")
	if expected := "content = 'O''Reilly%_\\'"; expected != condition {
		t.Fatalf("case-sensitive exact condition mismatch: got %q, want %q", condition, expected)
	}

	Conf.Search.CaseSensitive = false
	condition = buildExactSearchOrderCondition("content", "O'Reilly%_\\")
	if expected := "content LIKE 'O''Reilly\\%\\_\\\\' ESCAPE '\\'"; expected != condition {
		t.Fatalf("case-insensitive exact condition mismatch: got %q, want %q", condition, expected)
	}
}

func TestBuildOrderByPrioritizesCaseInsensitiveExactMatches(t *testing.T) {
	setSearchCaseSensitive(t, false)

	orderBy := buildOrderBy("seo", 0, 0)
	assertOrderBySequence(t, orderBy,
		"name LIKE 'seo' ESCAPE '\\'",
		"(',' || alias || ',') LIKE '%,seo,%' ESCAPE '\\'",
		"content LIKE 'seo' ESCAPE '\\' AND type = 'd'",
		"content LIKE '%seo%' AND type = 'd'",
		"content LIKE 'seo' ESCAPE '\\' AND type = 'h'",
		"content LIKE '%seo%' AND type = 'h'",
		"sort ASC",
	)

	orderBy = buildOrderBy("seo", 0, 7)
	assertOrderBySequence(t, orderBy,
		"content LIKE 'seo' ESCAPE '\\' AND type = 'd'",
		"content LIKE 'seo' ESCAPE '\\' AND type = 'h'",
		"rank",
	)

	orderBy = buildOrderBy("seo", 0, 6)
	if strings.Contains(orderBy, "content LIKE 'seo'") {
		t.Fatalf("ascending relevance should not put exact matches first: %q", orderBy)
	}
}

func TestBuildOrderByRanksCaseInsensitiveExactContentFirst(t *testing.T) {
	setSearchCaseSensitive(t, false)
	testDB, err := gosql.Open("sqlite3_extended", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		testDB.Close()
	})
	if _, err = testDB.Exec("CREATE TABLE blocks (name TEXT, alias TEXT, content TEXT, type TEXT, sort INTEGER, updated TEXT)"); err != nil {
		t.Fatal(err)
	}
	if _, err = testDB.Exec("INSERT INTO blocks VALUES ('', '', 'Learn seo', 'd', 0, ''), ('', '', 'SEO', 'd', 1, '')"); err != nil {
		t.Fatal(err)
	}

	row := testDB.QueryRow("SELECT content FROM blocks " + buildOrderBy("seo", 0, 0) + " LIMIT 1")
	var content string
	if err = row.Scan(&content); err != nil {
		t.Fatal(err)
	}
	if content != "SEO" {
		t.Fatalf("case-insensitive exact content should rank first: %q", content)
	}
}

func TestBuildOrderByRanksExactAliasSegmentFirst(t *testing.T) {
	setSearchCaseSensitive(t, false)
	testDB, err := gosql.Open("sqlite3_extended", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		testDB.Close()
	})
	if _, err = testDB.Exec("CREATE TABLE blocks (name TEXT, alias TEXT, content TEXT, type TEXT, sort INTEGER, updated TEXT)"); err != nil {
		t.Fatal(err)
	}
	if _, err = testDB.Exec("INSERT INTO blocks VALUES ('', '', 'SEO Guide', 'd', 0, ''), ('', 'AS,SEO', 'Other', 'd', 1, '')"); err != nil {
		t.Fatal(err)
	}

	row := testDB.QueryRow("SELECT content FROM blocks " + buildOrderBy("seo", 0, 0) + " LIMIT 1")
	var content string
	if err = row.Scan(&content); err != nil {
		t.Fatal(err)
	}
	if content != "Other" {
		t.Fatalf("exact alias segment should rank before partial document content: %q", content)
	}
}

func TestBuildExactAliasSearchOrderCondition(t *testing.T) {
	tests := []struct {
		name          string
		caseSensitive bool
		alias         string
		query         string
		matched       bool
	}{
		{name: "single alias", caseSensitive: true, alias: "technical", query: "technical", matched: true},
		{name: "first alias", caseSensitive: true, alias: "technical,writing", query: "technical", matched: true},
		{name: "middle alias", caseSensitive: true, alias: "docs,technical,writing", query: "technical", matched: true},
		{name: "last alias", caseSensitive: true, alias: "docs,technical", query: "technical", matched: true},
		{name: "partial alias", caseSensitive: true, alias: "technical-writing", query: "technical", matched: false},
		{name: "case sensitive mismatch", caseSensitive: true, alias: "SEO", query: "seo", matched: false},
		{name: "case insensitive match", alias: "AS,SEO", query: "seo", matched: true},
		{name: "escaped wildcard and backslash", alias: `other,100%_\\path,tail`, query: `100%_\\path`, matched: true},
		{name: "escaped quote", alias: "other,O'Reilly,tail", query: "O'Reilly", matched: true},
		{name: "comma query", alias: "foo,bar", query: "foo,bar", matched: false},
		{name: "empty query", alias: "", query: "", matched: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setSearchCaseSensitive(t, test.caseSensitive)
			testDB, err := gosql.Open("sqlite3_extended", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				testDB.Close()
			})

			condition := buildExactAliasSearchOrderCondition("alias", test.query)
			row := testDB.QueryRow("SELECT CASE WHEN "+condition+" THEN 1 ELSE 0 END FROM (SELECT ? AS alias)", test.alias)
			var matched int
			if err = row.Scan(&matched); err != nil {
				t.Fatal(err)
			}
			if test.matched != (matched == 1) {
				t.Fatalf("unexpected exact alias match state: condition=%q alias=%q query=%q result=%d", condition, test.alias, test.query, matched)
			}
		})
	}
}

func setSearchCaseSensitive(t *testing.T, caseSensitive bool) {
	t.Helper()
	oldConf := Conf
	Conf = NewAppConf()
	Conf.Search = conf.NewSearch()
	Conf.Search.CaseSensitive = caseSensitive
	t.Cleanup(func() {
		Conf = oldConf
	})
}

func assertOrderBySequence(t *testing.T, orderBy string, fragments ...string) {
	t.Helper()
	previous := -1
	for _, fragment := range fragments {
		current := strings.Index(orderBy, fragment)
		if 0 > current {
			t.Fatalf("order clause is missing %q: %q", fragment, orderBy)
		}
		if current <= previous {
			t.Fatalf("order priority is wrong; %q is not in the expected position: %q", fragment, orderBy)
		}
		previous = current
	}
}
