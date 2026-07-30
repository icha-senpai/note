//go:build js

package testing

import "github.com/icha-senpai/note/third_party/forks/github/gopherjs/gopherjs/js"

func init() {
	testBinary = js.Global.Get("$testBinary").String()
}
