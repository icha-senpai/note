// Copyright (c) 2019-present, Scribli


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
