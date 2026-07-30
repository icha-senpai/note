package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/icha-senpai/note/third_party/forks/github/pelletier/go-toml/v2/internal/testsuite"
)

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
	}

	err := testsuite.EncodeStdin()
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	log.Printf("Usage: %s < toml-file\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
	os.Exit(1)
}
