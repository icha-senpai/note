// SPDX-License-Identifier: BSD-1-Clause

package generator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func NormalizeBaseURL(unicodeBaseURL string) (string, error) {
	baseURL := strings.TrimSpace(unicodeBaseURL)
	if baseURL == "" {
		return "", fmt.Errorf("unicode-base-url is required")
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return baseURL, nil
}

// Generate downloads mapping files and generates Go sources into outputDir.
func Generate(outputDir, unicodeBaseURL string, vendor Vendor) error {
	baseURL, err := NormalizeBaseURL(unicodeBaseURL)
	if err != nil {
		return err
	}
	if vendor == nil {
		return fmt.Errorf("vendor is required")
	}

	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}
	if err := os.MkdirAll(absOutput, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	fmt.Printf("Generating encoding tables to: %s\n", absOutput)

	var allEncodings []EncodingData
	for _, name := range vendor.MappingFiles() {
		fmt.Printf("Processing %s...\n", name)
		data, err := fetchAndParse(baseURL, name)
		if err != nil {
			log.Printf("Warning: Failed to process %s: %v", name, err)
			continue
		}
		allEncodings = append(allEncodings, data)
	}

	cfg := vendor.Config(baseURL)

	if err := generateTablesFile(absOutput, allEncodings, cfg); err != nil {
		return fmt.Errorf("generate tables file: %w", err)
	}
	if err := generateEncodingsFile(absOutput, allEncodings, cfg); err != nil {
		return fmt.Errorf("generate encodings file: %w", err)
	}

	fmt.Println("Generation complete!")
	return nil
}

func fetchAndParse(baseURL, name string) (EncodingData, error) {
	url := baseURL + name + ".TXT"
	content, err := fetchURL(url)
	if err != nil {
		return EncodingData{}, fmt.Errorf("fetch %s: %w", url, err)
	}

	mappings, err := parseMappingFile(content)
	if err != nil {
		return EncodingData{}, fmt.Errorf("parse %s: %w", name, err)
	}

	return EncodingData{
		Name:     name,
		Mappings: mappings,
	}, nil
}
