// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gofix defines an analyzer that checks go:fix directives.
package gofix

import (
	_ "embed"

	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/inspect"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/internal/gofixdirective"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/ast/inspector"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/internal/analysis/analyzerutil"
)

//go:embed doc.go
var doc string

var Analyzer = &analysis.Analyzer{
	Name:     "gofixdirective",
	Doc:      analyzerutil.MustExtractDoc(doc, "gofixdirective"),
	URL:      "https://pkg.go.dev/github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/gofix",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	root := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector).Root()
	gofixdirective.Find(pass, root, nil)
	return nil, nil
}
