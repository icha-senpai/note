package main

import (
	"fmt"

	"github.com/icha-senpai/note/third_party/forks/clipboard"
)

func main() {
	text, err := clipboard.ReadAll()
	if err != nil {
		panic(err)
	}

	fmt.Print(text)
}
