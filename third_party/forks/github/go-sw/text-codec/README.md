# text-codec

A Go library providing legacy text encodings, implementing the `github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding` interface.

## Overview

This library provides text encoding transformers for legacy character sets, including Mac OS encodings and Korean Johab encoding. The encodings are derived from the official Unicode Consortium mapping files.

## Installation

```bash
go get github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec
```

## Supported Encodings

This library supports **24 Mac OS encodings** from the Unicode Consortium mapping files.

> **Note:** The `CORPCHAR.TXT` file from Unicode is not an encoding, it's a reference document listing Apple's corporate-zone Unicode characters (U+F8FF, etc.) and is not included as an encoding.

### European

| Encoding                | Description                | Variable                |
| ----------------------- | -------------------------- | ----------------------- |
| Mac OS Central European | Central European languages | `apple.CentralEuropean` |
| Mac OS Croatian         | Croatian                   | `apple.Croatian`        |
| Mac OS Greek            | Greek                      | `apple.Greek`           |
| Mac OS Icelandic        | Icelandic                  | `apple.Iceland`         |
| Mac OS Romanian         | Romanian                   | `apple.Romanian`        |
| Mac OS Turkish          | Turkish                    | `apple.Turkish`         |
| Mac OS Celtic           | Celtic languages           | `apple.Celtic`          |
| Mac OS Gaelic           | Gaelic                     | `apple.Gaelic`          |
| Mac OS Ukrainian        | Ukrainian                  | `apple.Ukraine`         |

### Middle Eastern

| Encoding      | Description   | Variable       |
| ------------- | ------------- | -------------- |
| Mac OS Arabic | Arabic        | `apple.Arabic` |
| Mac OS Farsi  | Farsi/Persian | `apple.Farsi`  |
| Mac OS Hebrew | Hebrew        | `apple.Hebrew` |

### Indic

| Encoding          | Description           | Variable           |
| ----------------- | --------------------- | ------------------ |
| Mac OS Devanagari | Hindi, Sanskrit, etc. | `apple.Devanagari` |
| Mac OS Gujarati   | Gujarati              | `apple.Gujarati`   |
| Mac OS Gurmukhi   | Punjabi               | `apple.Gurmukhi`   |

### East Asian (CJK)

| Encoding                   | Description         | Variable                   | Base Encoding                  |
| -------------------------- | ------------------- | -------------------------- | ------------------------------ |
| Mac OS Japanese            | Japanese            | `apple.Japanese`           | Shift-JIS with Apple overrides |
| Mac OS Korean              | Korean              | `apple.Korean`             | EUC-KR with Apple overrides    |
| Mac OS Chinese Simplified  | Simplified Chinese  | `apple.ChineseSimplified`  | GBK with Apple overrides       |
| Mac OS Chinese Traditional | Traditional Chinese | `apple.ChineseTraditional` | Big5 with Apple overrides      |

### Other

| Encoding        | Description       | Variable         |
| --------------- | ----------------- | ---------------- |
| Mac OS Thai     | Thai              | `apple.Thai`     |
| Mac OS Inuit    | Inuit languages   | `apple.Inuit`    |
| Mac OS Dingbats | Dingbat symbols   | `apple.Dingbats` |
| Mac OS Symbol   | Symbol characters | `apple.Symbol`   |
| Mac OS Keyboard | Keyboard glyphs   | `apple.Keyboard` |

### Korean Encodings

| Encoding | Description                    | Variable       |
| -------- | ------------------------------ | -------------- |
| Johab    | KS C 5601-1992 Annex 3 (Johab) | `korean.Johab` |

The Johab encoding uses **algorithmic bit manipulation** for Hangul syllables rather than lookup tables, supporting:

### NeXTSTEP Encoding

| Encoding | Description                    | Variable        |
| -------- | ------------------------------ | --------------- |
| NeXTSTEP | NeXT computer character set    | `next.NeXTSTEP` |

The NeXTSTEP encoding was used on NeXT computers (1988-1996) and is still encountered in legacy OpenStep property lists and NeXT-era documents.

### Atari ST Encoding

| Encoding | Description                     | Variable   |
| -------- | ------------------------------- | ---------- |
| Atari ST | Atari ST/TT (TOS) character set | `atari.ST` |

The Atari ST encoding was used on Atari ST series computers (1985-1993) and is still encountered in legacy Atari files, demoscene content, and retro computing applications.

### Hangul Syllables (Johab)

- Precomposed Hangul syllables (U+AC00-U+D7A3)
- Compatibility Jamo (U+3131-U+318E)
- Decomposed Jamo (U+1100-U+11FF)
- Hanja, Greek, Cyrillic, and other symbols via lookup tables

## Usage

### Basic Decoding (Legacy to UTF-8)

```go
package main

import (
    "fmt"
    "github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec/apple"
)

func main() {
    // Mac Central European encoded bytes
    macCentEuroBytes := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x80} // "Hello" + Ä

    // Decode to UTF-8
    decoder := apple.CentralEuropean.NewDecoder()
    utf8Bytes, err := decoder.Bytes(macCentEuroBytes)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(utf8Bytes)) // Output: HelloÄ
}
```

### Basic Encoding (UTF-8 to Legacy)

```go
package main

import (
    "fmt"
    "github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec/apple"
)

func main() {
    // UTF-8 string containing special characters
    utf8String := "Größe"

    // Encode to Mac Central European
    encoder := apple.CentralEuropean.NewEncoder()
    macCentEuroBytes, err := encoder.Bytes([]byte(utf8String))
    if err != nil {
        panic(err)
    }

    fmt.Printf("%v\n", macCentEuroBytes)
}
```

### Streaming with transform.Reader

```go
package main

import (
    "bytes"
    "io"
    "github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec/apple"
    "github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform"
)

func main() {
    // Legacy encoded data
    legacyData := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}

    // Create a transforming reader
    reader := transform.NewReader(bytes.NewReader(legacyData), apple.CentralEuropean.NewDecoder())

    // Read UTF-8 data
    utf8Data, _ := io.ReadAll(reader)
    println(string(utf8Data))
}
```

### Streaming with transform.Writer

```go
package main

import (
    "bytes"
    "github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec/apple"
    "github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/transform"
)

func main() {
    var buf bytes.Buffer

    // Create a transforming writer
    writer := transform.NewWriter(&buf, apple.CentralEuropean.NewEncoder())

    // Write UTF-8 data (automatically encoded to Mac Central European)
    writer.Write([]byte("Hello, World!"))
    writer.Close()

    // buf now contains Mac Central European encoded bytes
}
```

### Korean Johab Encoding

```go
package main

import (
    "fmt"
    "github.com/icha-senpai/note/third_party/forks/github/go-sw/text-codec/korean"
)

func main() {
    // Encode Korean text to Johab
    encoder := korean.Johab.NewEncoder()
    johabBytes, err := encoder.Bytes([]byte("조합"))
    if err != nil {
        panic(err)
    }
    fmt.Printf("Johab: % X\n", johabBytes) // Output: B9 A1 D0 73

    // Decode Johab to UTF-8
    decoder := korean.Johab.NewDecoder()
    utf8Bytes, err := decoder.Bytes(johabBytes)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(utf8Bytes)) // Output: 조합
}
```

## Encoding Implementation

### Single-Byte Encodings

Most Mac OS encodings (20 in this package) are single-byte character sets where each byte maps to a specific Unicode code point. These are implemented using simple 256-entry lookup tables.

### Extended (CJK) Encodings

The East Asian encodings (Japanese, Korean, Chinese Simplified, Chinese Traditional) are multi-byte encodings that extend standard encodings with Apple-specific byte overrides:

- **Japanese**: Extends Shift-JIS with overrides for yen sign (0x5C→U+00A5), overline, etc.
- **Korean**: Extends EUC-KR with overrides for won sign (0x81→U+20A9), etc.
- **Chinese Simplified**: Extends GBK with Apple-specific overrides
- **Chinese Traditional**: Extends Big5 with Apple-specific overrides

This follows the same pattern used by [fonttools](https://github.com/fonttools/fonttools) for handling Mac OS CJK encodings.

### Korean Johab Encoding

Johab (KS C 5601-1992 Annex 3) uses **algorithmic bit manipulation** to encode Hangul syllables:

```
byte1 = 0x80 | (cho << 2) | (jung >> 3)
byte2 = (jung << 5) | jong
```

Where `cho`, `jung`, and `jong` are 5-bit indices for the initial consonant (choseong), vowel (jungseong), and final consonant (jongseong).

This approach encodes all 11,172 modern Hangul syllables without requiring a full lookup table.

## Building from Source

### Prerequisites

- Go 1.16 or later
- Task: <https://github.com/go-task/task>
- Internet connection (for downloading Unicode mapping files)

Install Task using Go:

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

### Generate Encoding Tables

```bash
# Using task
task generate

# Or using go generate
go generate ./...
```

### Run Tests

```bash
task test
# or
go test -v ./...
```

### Run Benchmarks

```bash
go test -bench=. ./...
```

## Data Source

The encoding tables are derived from the Unicode Consortium mapping files:

- **Apple encodings**: <https://www.unicode.org/Public/MAPPINGS/VENDORS/APPLE/>
- **Korean Johab**: <https://www.unicode.org/Public/MAPPINGS/OBSOLETE/EASTASIA/KSC/JOHAB.TXT>
- **NeXTSTEP**: <https://www.unicode.org/Public/MAPPINGS/VENDORS/NEXT/NEXTSTEP.TXT>
- **Atari ST**: <https://www.unicode.org/Public/MAPPINGS/VENDORS/MISC/ATARIST.TXT>

Additional references:

- CPython cjkcodecs: <https://github.com/python/cpython/blob/main/Modules/cjkcodecs/_codecs_kr.c>
- ICU charset data: <https://github.com/unicode-org/icu-data>

The encoding tables are derived from the Unicode Consortium's archival character mapping files.

## License

This library is licensed under the **BSD-1-Clause** license. See [LICENSE.md](LICENSE.md) for details.

## Contributing

Contributions are welcome! Please ensure any changes:

1. Follow Go coding conventions
2. Include appropriate tests
3. Maintain compatibility with the `github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding` interface
4. Update documentation as needed

## Acknowledgments

- Unicode, Inc. for maintaining the character mapping files
- [CPython](https://github.com/python/cpython) project for Johab encoding implementation.
- The [fonttools](https://github.com/fonttools/fonttools) project for extended encoding implementation.
- The Go Authors for the `github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text` package

## See Also

- [github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding](https://pkg.go.dev/github.com/icha-senpai/note/third_party/forks/external/golang.org/x/text/encoding) - Go text encoding interface
- [Unicode MAPPINGS](https://www.unicode.org/Public/MAPPINGS/) - Unicode character mapping files
