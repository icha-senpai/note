// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/sqlrowserr"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(sqlrowserr.Analyzer) }
