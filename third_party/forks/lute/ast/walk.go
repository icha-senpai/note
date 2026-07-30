// Copyright (c) 2019-present, Scribli
//
// Lute is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package ast

type WalkStatus int

const (
	WalkStop = iota
	WalkSkipChildren
	WalkContinue
)

type Walker func(n *Node, entering bool) WalkStatus

func Walk(n *Node, walker Walker) {
	walk(n, walker)
}

func walk(n *Node, walker Walker) (ret WalkStatus) {
	ret = walker(n, true)
	if ret == WalkStop {
		return
	}

	if ret != WalkSkipChildren {
		for c := n.FirstChild; nil != c; c = c.Next {
			if ret = walk(c, walker); WalkStop == ret {
				return WalkStop
			}
		}
	}

	ret = walker(n, false)
	return
}
