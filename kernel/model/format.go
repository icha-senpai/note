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
	"bytes"

	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/kernel/util"
	"github.com/icha-senpai/note/third_party/forks/logging"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
	"github.com/icha-senpai/note/third_party/forks/lute/editor"
	"github.com/icha-senpai/note/third_party/forks/lute/render"
)

func AutoSpace(rootID string) (err error) {
	tree, err := LoadTreeByBlockID(rootID)
	if err != nil {
		return
	}

	logging.LogInfof("formatting tree [%s]...", rootID)
	util.PushProtyleLoading(rootID, Conf.Language(116))
	defer ReloadProtyle(rootID)

	FlushTxQueue()

	generateOpTypeHistory(tree, HistoryOpFormat)
	luteEngine := NewLute()
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		switch n.Type {
		case ast.NodeTextMark:
			luteEngine.MergeSameTextMark(n)
		case ast.NodeCodeBlockCode:

			n.Tokens = bytes.ReplaceAll(n.Tokens, []byte(editor.Zwj+"```"), []byte("```"))
			n.Tokens = bytes.ReplaceAll(n.Tokens, []byte("```"), []byte(editor.Zwj+"```"))
		}
		return ast.WalkContinue
	})

	rootIAL := tree.Root.KramdownIAL
	addBlockIALNodes(tree, false)

	formatRenderer := render.NewFormatRenderer(tree, luteEngine.RenderOptions, luteEngine.ParseOptions)
	md := formatRenderer.Render()
	newTree := parseKTree(md)
	newTree.Root.Spec = treenode.CurrentSpec

	luteEngine.SetAutoSpace(true)
	formatRenderer = render.NewFormatRenderer(newTree, luteEngine.RenderOptions, luteEngine.ParseOptions)
	md = formatRenderer.Render()
	newTree = parseKTree(md)
	newTree.Root.Spec = treenode.CurrentSpec
	newTree.Root.ID = tree.ID
	newTree.Root.KramdownIAL = rootIAL
	newTree.ID = tree.ID
	newTree.Path = tree.Path
	newTree.HPath = tree.HPath
	newTree.Box = tree.Box
	err = writeTreeUpsertQueue(newTree)
	if err != nil {
		return
	}
	logging.LogInfof("formatted tree [%s]", rootID)
	util.RandomSleep(500, 700)
	return
}
