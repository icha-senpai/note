// font is a utility that can parse and print information about font files.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/icha-senpai/note/third_party/forks/github/ConradIrwin/font/sfnt"
)

func usage() {
	fmt.Println(`
Usage: font [-i font-index] <features|info|metrics|scrub|stats> font.[otf,ttf,ttc,woff,woff2] ...

features: prints the gpos/gsub tables (contains font features)
info: prints the name table (contains metadata)
metrics: prints the hhea table (contains font metrics)
scrub: remove the name table (saves significant space)
stats: prints each table and the amount of space used`)
}

func main() {
	fontIndex := flag.Int("i", -1, "select `font-index` for TrueType Collection (.ttc/.otc), starting from 0.")

	flag.Usage = func() {
		usage()
		flag.PrintDefaults()
	}
	flag.Parse()

	command := "help"
	if cmd := flag.Arg(0); len(cmd) > 0 {
		command = cmd
	}

	cmds := map[string]func(*sfnt.Font) error{
		"scrub":    Scrub,
		"info":     Info,
		"stats":    Stats,
		"metrics":  Metrics,
		"features": Features,
	}
	if _, found := cmds[command]; !found || len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(2)
	}

	filenames := flag.Args()[1:]
	exitCode := 0
	runCommand := func(font *sfnt.Font) {
		if err := cmds[command](font); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			exitCode = 1
		}
	}

	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open font: %s\n", err)
			exitCode = 1
			continue
		}
		defer file.Close()

		if len(filenames) > 1 {
			fmt.Println("==>", filename, "<==")
		}

		isCollection, err := sfnt.IsCollection(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to determine if font is collection: %s\n", err)
			exitCode = 1
			continue
		}

		if !isCollection {
			font, err := sfnt.Parse(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse font: %s\n", err)
				exitCode = 1
				continue
			}

			runCommand(font)
		} else {
			if *fontIndex >= 0 {
				font, err := sfnt.ParseCollectionIndex(file, uint32(*fontIndex))
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to parse font index %d from collection: %s\n", *fontIndex, err)
					exitCode = 1
					continue
				}

				runCommand(font)
			} else {
				fonts, err := sfnt.ParseCollection(file)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to parse fonts from collection: %s\n", err)
					exitCode = 1
					continue
				}

				for i, font := range fonts {
					fmt.Printf("==>font index: %d<==\n", i)

					runCommand(font)
				}
			}
		}
	}
	os.Exit(exitCode)
}
