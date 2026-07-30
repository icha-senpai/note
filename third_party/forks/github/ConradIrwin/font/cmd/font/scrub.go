package main

import (
	"os"

	"github.com/icha-senpai/note/third_party/forks/github/ConradIrwin/font/sfnt"
)

// Scrub remove the name table (saves significant space).
func Scrub(font *sfnt.Font) error {
	// TODO handle multiple fonts and collection as current implementation writes all into stdout
	// - if a single font is handle within the entire process, allow writing to stdout
	// - else needs to be written into separate files, if stdout redirection is specified return an error
	if font.HasTable(sfnt.TagName) {
		font.AddTable(sfnt.TagName, sfnt.NewTableName())
	}

	_, err := font.WriteOTF(os.Stdout)
	return err
}
