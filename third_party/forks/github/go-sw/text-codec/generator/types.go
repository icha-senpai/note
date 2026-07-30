// SPDX-License-Identifier: BSD-1-Clause

package generator

// Mapping represents a single byte to Unicode mapping.
type Mapping struct {
	Byte    byte
	Unicode rune
	Name    string
}

// EncodingData holds all mappings for an encoding.
type EncodingData struct {
	Name     string
	Mappings []Mapping
}

// VendorConfig holds vendor-specific configuration for code generation.
type VendorConfig struct {
	// PackageName is the Go package name (e.g., "apple", "ibm", "msdos")
	PackageName string
	// VendorName is the human-readable vendor name for comments (e.g., "Apple", "IBM", "MS-DOS")
	VendorName string
	// SourceURL is the base URL where mapping files are sourced from
	SourceURL string
	// EncodingPrefix is the prefix for encoding String() method (e.g., "Mac", "IBM", "CP")
	EncodingPrefix string
	// NameReplacements maps encoding names to Go-friendly names (optional per-vendor overrides)
	NameReplacements map[string]string
}

// Vendor provides vendor-specific generation inputs.
//
// The CLI can implement this from data (e.g. JSON specs) so it doesn't need to
// import vendor Go packages.
type Vendor interface {
	MappingFiles() []string
	Config(sourceURL string) VendorConfig
}
