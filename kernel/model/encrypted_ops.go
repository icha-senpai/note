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
	"github.com/icha-senpai/note/kernel/sql"
	"github.com/icha-senpai/note/kernel/treenode"
	"github.com/icha-senpai/note/third_party/forks/lute/ast"
)

func GetBlockRefTextInBox(id, boxID string) string {
	FlushTxQueue()

	bt := treenode.GetBlockTreeInBox(id, boxID)
	if nil == bt {
		return ErrBlockNotFound.Error()
	}

	tree, err := loadTreeByBlockTree(bt)
	if err != nil {
		return ""
	}

	node := treenode.GetNodeInTree(tree, id)
	if nil == node {
		return ErrBlockNotFound.Error()
	}

	ast.Walk(node, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}
		if n.IsTextMarkType("inline-memo") {
			n.TextMarkInlineMemoContent = ""
			return ast.WalkContinue
		}
		return ast.WalkContinue
	})

	return getNodeRefText(node)
}

func GetRefTextInBox(defBlockID, boxID string) string {
	return sql.GetRefTextInBox(defBlockID, boxID)
}
