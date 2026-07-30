// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The protoreflect build tag disables use of fast-path methods.
//go:build protoreflect
// +build protoreflect

package proto

import (
	"github.com/icha-senpai/note/third_party/forks/external/google.golang.org/protobuf/reflect/protoreflect"
	"github.com/icha-senpai/note/third_party/forks/external/google.golang.org/protobuf/runtime/protoiface"
)

const hasProtoMethods = false

func protoMethods(m protoreflect.Message) *protoiface.Methods {
	return nil
}
