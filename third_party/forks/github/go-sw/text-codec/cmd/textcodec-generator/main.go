// SPDX-License-Identifier: BSD-1-Clause

// Package main implements a vendor-agnostic code generator that downloads Unicode
// mapping files and generates Go lookup tables for legacy encodings.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec/generator"
)

func main() {
	outputDir := flag.String("output", "", "Output directory for generated files (defaults to spec packageName)")
	unicodeBaseURL := flag.String("unicode-base-url", "", "Base URL for Unicode mapping files (e.g. https://www.unicode.org/Public/MAPPINGS/VENDORS/APPLE/)")
	specPath := flag.String("spec", "", "Path to a vendor spec JSON file")
	flag.Parse()

	if *specPath == "" {
		log.Fatalf("spec is required")
	}

	spec, err := generator.LoadSpecFile(*specPath)
	if err != nil {
		log.Fatalf("failed to load spec: %v", err)
	}

	out := *outputDir
	if out == "" {
		out = spec.PackageName
	}

	vendor := generator.SpecVendor{Spec: spec}
	if err := generator.Generate(out, *unicodeBaseURL, vendor); err != nil {
		log.Fatalf("generation failed: %v", err)
	}

	fmt.Println("OK")
}
