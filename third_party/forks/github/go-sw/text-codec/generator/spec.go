// SPDX-License-Identifier: BSD-1-Clause

package generator

import (
	"encoding/json"
	"errors"
	"os"
)

// Spec describes a vendor's inputs and naming conventions.
// It is designed to be loaded from a JSON file so the CLI can remain
// independent of any vendor-specific Go packages.
type Spec struct {
	PackageName      string            `json:"packageName"`
	VendorName       string            `json:"vendorName"`
	SourceURL        string            `json:"sourceUrl,omitempty"`
	EncodingPrefix   string            `json:"encodingPrefix"`
	NameReplacements map[string]string `json:"nameReplacements"`
	MappingFiles     []string          `json:"mappingFiles"`
}

func LoadSpecFile(path string) (Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return Spec{}, err
	}
	if spec.PackageName == "" {
		return Spec{}, errors.New("spec packageName is required")
	}
	if spec.VendorName == "" {
		return Spec{}, errors.New("spec vendorName is required")
	}
	if len(spec.MappingFiles) == 0 {
		return Spec{}, errors.New("spec mappingFiles is required")
	}
	if spec.NameReplacements == nil {
		spec.NameReplacements = map[string]string{}
	}
	return spec, nil
}

// SpecVendor adapts a Spec to the Vendor interface.
type SpecVendor struct {
	Spec Spec
}

func (v SpecVendor) MappingFiles() []string { return v.Spec.MappingFiles }

func (v SpecVendor) Config(sourceURL string) VendorConfig {
	cfg := VendorConfig{
		PackageName:      v.Spec.PackageName,
		VendorName:       v.Spec.VendorName,
		SourceURL:        sourceURL,
		EncodingPrefix:   v.Spec.EncodingPrefix,
		NameReplacements: v.Spec.NameReplacements,
	}
	if v.Spec.SourceURL != "" {
		cfg.SourceURL = v.Spec.SourceURL
	}
	return cfg
}
