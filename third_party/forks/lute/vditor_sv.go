// Copyright (c) 2019-present, Scribli


package lute

import (
	"strings"

	"github.com/icha-senpai/note/third_party/forks/lute/editor"
	"github.com/icha-senpai/note/third_party/forks/lute/parse"
	"github.com/icha-senpai/note/third_party/forks/lute/render"
)

func (lute *Lute) SpinVditorSVDOM(markdown string) (ovHTML string) {
	if editor.Caret == strings.TrimSpace(markdown) {
		return "<span data-type=\"text\"><wbr></span>" + string(render.NewlineSV)
	}

	tree := parse.Parse("", []byte(markdown), lute.ParseOptions)

	renderer := render.NewVditorSVRenderer(tree, lute.RenderOptions, lute.ParseOptions)
	output := renderer.Render()
	ovHTML = strings.ReplaceAll(string(output), editor.Caret, "<wbr>")
	return
}

func (lute *Lute) HTML2VditorSVDOM(sHTML string) (vHTML string) {
	markdown, err := lute.HTML2Markdown(sHTML)
	if nil != err {
		vHTML = err.Error()
		return
	}

	tree := parse.Parse("", []byte(markdown), lute.ParseOptions)
	renderer := render.NewVditorSVRenderer(tree, lute.RenderOptions, lute.ParseOptions)
	for nodeType, rendererFunc := range lute.HTML2VditorSVDOMRendererFuncs {
		renderer.ExtRendererFuncs[nodeType] = rendererFunc
	}
	output := renderer.Render()
	vHTML = string(output)
	return
}

func (lute *Lute) Md2VditorSVDOM(markdown string) (vHTML string) {
	tree := parse.Parse("", []byte(markdown), lute.ParseOptions)
	renderer := render.NewVditorSVRenderer(tree, lute.RenderOptions, lute.ParseOptions)
	for nodeType, rendererFunc := range lute.Md2VditorSVDOMRendererFuncs {
		renderer.ExtRendererFuncs[nodeType] = rendererFunc
	}
	output := renderer.Render()
	vHTML = strings.ReplaceAll(string(output), editor.Caret, "<wbr>")
	return
}
