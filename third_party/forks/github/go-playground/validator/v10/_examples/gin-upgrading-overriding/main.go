package main

import "github.com/icha-senpai/note/third_party/forks/github/gin-gonic/gin/binding"

func main() {

	binding.Validator = new(defaultValidator)

	// regular gin logic
}
