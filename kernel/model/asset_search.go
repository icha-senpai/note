// Based on SiYuan; Copyright (c) 2020-present, b3log.org.
// SPDX-License-Identifier: AGPL-3.0-or-later

package model

import (
	"net/url"
	"strings"

	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/lute"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/parse"
)

// imageAssetMetadataHits merges image descriptions and titles, keeping each asset's strongest match.
func imageAssetMetadataHits(markdowns, keywords []string) map[string]int {
	ret := map[string]int{}
	luteEngine := lute.New()
	for _, markdown := range markdowns {
		tree := parse.Parse("", []byte(markdown), luteEngine.ParseOptions)
		ast.Walk(tree.Root, func(node *ast.Node, entering bool) ast.WalkStatus {
			if !entering || node.Type != ast.NodeImage {
				return ast.WalkContinue
			}
			dest := node.ChildByType(ast.NodeLinkDest)
			if dest == nil {
				return ast.WalkSkipChildren
			}
			assetPath := strings.SplitN(util.UnescapeHTML(dest.TokensStr()), "?", 2)[0]
			if decoded, err := url.PathUnescape(assetPath); err == nil {
				assetPath = decoded
			}
			if !strings.HasPrefix(assetPath, "assets/") {
				return ast.WalkSkipChildren
			}
			hits := assetMetadataTextHits(util.UnescapeHTML(node.Content()), keywords)
			if title := node.ChildByType(ast.NodeLinkTitle); title != nil {
				hits += assetMetadataTextHits(util.UnescapeHTML(title.TokensStr()), keywords)
			}
			ret[assetPath] = max(ret[assetPath], hits)
			return ast.WalkSkipChildren
		})
	}
	return ret
}

func assetMetadataTextHits(text string, keywords []string) int {
	text = strings.ToLower(text)
	hits := 0
	for i, keyword := range keywords {
		if strings.TrimSpace(keyword) == "" {
			continue
		}
		keyword = strings.ToLower(keyword)
		if strings.Contains(text, keyword) {
			if i == 0 {
				hits += 32
			}
			hits += strings.Count(text, keyword)
		}
	}
	return hits
}
