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

package treenode

import (
	"github.com/88250/lute/ast"
	"github.com/88250/lute/parse"
)

func MoveFoldHeading(updateNode, oldNode *ast.Node) {
	foldHeadings := map[string][]*ast.Node{}

	ast.Walk(oldNode, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		if ast.NodeHeading == n.Type && "1" == n.IALAttr("fold") {
			children := HeadingChildren(n)
			foldHeadings[n.ID] = children
		}
		return ast.WalkContinue
	})

	var updateFoldHeadings []*ast.Node
	ast.Walk(updateNode, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		if ast.NodeHeading == n.Type && "1" == n.IALAttr("fold") {
			updateFoldHeadings = append(updateFoldHeadings, n)
		}
		return ast.WalkContinue
	})
	for _, h := range updateFoldHeadings {
		children := foldHeadings[h.ID]
		for i := len(children) - 1; 0 <= i; i-- {
			h.Next.InsertAfter(children[i])
		}
	}
}

type FoldHeadingStack struct {
	levels []int
	last   *ast.Node
}

func (s *FoldHeadingStack) Enter(n *ast.Node) {
	s.last = n
	if ast.NodeHeading != n.Type {
		return
	}

	for 0 < len(s.levels) && s.levels[len(s.levels)-1] >= n.HeadingLevel {
		s.levels = s.levels[:len(s.levels)-1]
	}

	if "1" == n.IALAttr("fold") {
		s.levels = append(s.levels, n.HeadingLevel)
	}
}

func (s *FoldHeadingStack) Hidden() bool {
	depth := len(s.levels)
	if 0 == depth {
		return false
	}

	if n := s.last; nil != n && ast.NodeHeading == n.Type && "1" == n.IALAttr("fold") && s.levels[depth-1] == n.HeadingLevel {

		return 1 < depth
	}
	return true
}

func CollectFoldHiddenNodes(parent *ast.Node) (unlinks []*ast.Node) {
	if nil == parent {
		return
	}

	collectFoldHiddenNodes(parent, &unlinks)
	return
}

func collectFoldHiddenNodes(parent *ast.Node, unlinks *[]*ast.Node) {
	var stack FoldHeadingStack
	for n := parent.FirstChild; nil != n; n = n.Next {
		stack.Enter(n)
		if stack.Hidden() {
			*unlinks = append(*unlinks, n)
			continue
		}

		if n.IsContainerBlock() {
			collectFoldHiddenNodes(n, unlinks)
		}
	}
}

func IsInFoldedHeading(node, currentHeading *ast.Node) bool {
	if nil == node {
		return false
	}

	heading := HeadingParent(node)
	if nil == heading {
		return false
	}
	if ast.NodeHeading == heading.Type {
		if "1" == heading.IALAttr("heading-fold") || "1" == heading.IALAttr("fold") {
			return true
		}
		if heading == currentHeading {

			return false
		}
	}
	return IsInFoldedHeading(heading, currentHeading)
}

func GetHeadingFold(nodes []*ast.Node) (ret []*ast.Node) {
	for _, n := range nodes {
		if "1" == n.IALAttr("heading-fold") {
			ret = append(ret, n)
		}
	}
	return
}

func GetParentFoldedHeading(node *ast.Node) (parentFoldedHeading *ast.Node) {
	if nil == node {
		return
	}

	currentLevel := 7
	if ast.NodeHeading == node.Type {
		currentLevel = node.HeadingLevel
	}
	for n := node.Previous; nil != n; n = n.Previous {
		if ast.NodeHeading != n.Type {
			continue
		}

		if n.HeadingLevel >= currentLevel {
			continue
		}
		currentLevel = n.HeadingLevel

		if "1" == n.IALAttr("fold") {
			if ast.NodeHeading != node.Type {
				parentFoldedHeading = n
			}
			if n.HeadingLevel < node.HeadingLevel {
				parentFoldedHeading = n
			}
		}
	}
	return
}

func HeadingChildren(heading *ast.Node) (ret []*ast.Node) {
	start := heading.Next
	if nil == start {
		return
	}
	if ast.NodeKramdownBlockIAL == start.Type {
		start = start.Next
	}

	currentLevel := heading.HeadingLevel
	for n := start; nil != n; n = n.Next {
		if ast.NodeHeading == n.Type {
			if currentLevel >= n.HeadingLevel {
				break
			}
		}
		ret = append(ret, n)
	}
	return
}

func SuperBlockLastHeading(sb *ast.Node) *ast.Node {
	headings := sb.ChildrenByType(ast.NodeHeading)
	if 0 < len(headings) {
		return headings[len(headings)-1]
	}
	return nil
}

func HeadingParent(node *ast.Node) *ast.Node {
	if nil == node {
		return nil
	}

	currentLevel := 16
	if ast.NodeHeading == node.Type {
		currentLevel = node.HeadingLevel
	}

	for n := node.Previous; nil != n; n = n.Previous {
		if ast.NodeHeading == n.Type && n.HeadingLevel < currentLevel {
			return n
		}
	}
	return node.Parent
}

func HeadingLevel(node *ast.Node) int {
	if nil == node {
		return 0
	}

	for n := node; nil != n; n = n.Previous {
		if ast.NodeHeading == n.Type {
			return n.HeadingLevel
		}
	}
	return 0
}

func TopHeadingLevel(tree *parse.Tree) (ret int) {
	ret = 7
	for n := tree.Root.FirstChild; nil != n; n = n.Next {
		if ast.NodeHeading == n.Type {
			if ret > n.HeadingLevel {
				ret = n.HeadingLevel
			}
		}
	}
	if 7 == ret {
		ret = 0
	}
	return
}
