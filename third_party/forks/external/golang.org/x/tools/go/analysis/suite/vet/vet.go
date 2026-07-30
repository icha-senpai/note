// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The vet package defines the suite of analyzers used by cmd/vet,
// the default analysis tool run by "go vet".
// Its behavior is equivalent to:
//
//	func main() { unitchecker.Main(vet.Suite...) }
//
// If you need a different suite, define your own tool
// and run "go vet -vettool=mytool".
package vet

import (
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/appends"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/asmdecl"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/assign"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/atomic"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/bools"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/buildtag"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/cgocall"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/composite"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/copylock"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/defers"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/directive"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/errorsas"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/framepointer"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/hostport"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/httpresponse"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/ifaceassert"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/loopclosure"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/lostcancel"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/nilfunc"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/printf"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/shift"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/slog"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/stdmethods"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/stdversion"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/stringintconv"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/structtag"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/tests"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/timeformat"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/unmarshal"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/unreachable"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/unsafeptr"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/unusedresult"
	"github.com/icha-senpai/note/third_party/forks/external/golang.org/x/tools/go/analysis/passes/waitgroup"
)

// Suite is the suite of analyzers run by cmd/vet.
//
// The vet suite analyzers report diagnostics.
// (Diagnostics must describe real problems, but need not
// suggest fixes, and fixes are not necessarily safe to apply.)
var Suite = []*analysis.Analyzer{
	appends.Analyzer,
	asmdecl.Analyzer,
	assign.Analyzer,
	atomic.Analyzer,
	bools.Analyzer,
	buildtag.Analyzer,
	cgocall.Analyzer,
	composite.Analyzer,
	copylock.Analyzer,
	defers.Analyzer,
	directive.Analyzer,
	errorsas.Analyzer,
	// fieldalignment.Analyzer omitted: too noisy
	framepointer.Analyzer,
	httpresponse.Analyzer,
	hostport.Analyzer,
	ifaceassert.Analyzer,
	loopclosure.Analyzer,
	lostcancel.Analyzer,
	nilfunc.Analyzer,
	printf.Analyzer,
	// scannererr.Analyzer, // TODO(adonovan): add to go vet for 1.28 after the freeze (#17747)
	// shadow.Analyzer omitted: too noisy
	shift.Analyzer,
	sigchanyzer.Analyzer,
	slog.Analyzer,
	// sqlrowserr.Analyzer, // TODO(adonovan): add to go vet for 1.28 after the freeze (#17747)
	stdmethods.Analyzer,
	stdversion.Analyzer,
	stringintconv.Analyzer,
	structtag.Analyzer,
	tests.Analyzer,
	testinggoroutine.Analyzer,
	timeformat.Analyzer,
	unmarshal.Analyzer,
	unreachable.Analyzer,
	unsafeptr.Analyzer,
	unusedresult.Analyzer,
	waitgroup.Analyzer,
}
