# HTML

![test](https://github.com/icha-senpai/note/third_party/forks/github/wmentor/html/workflows/test/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/wmentor/html/badge.svg?branch=master&v=1.0.1)](https://coveralls.io/github/wmentor/html?branch=master)
[![https://goreportcard.com/report/github.com/icha-senpai/note/third_party/forks/github/wmentor/html](https://goreportcard.com/badge/github.com/icha-senpai/note/third_party/forks/github/wmentor/html)](https://goreportcard.com/report/github.com/icha-senpai/note/third_party/forks/github/wmentor/html)
[![https://pkg.go.dev/github.com/icha-senpai/note/third_party/forks/github/wmentor/html](https://pkg.go.dev/badge/github.com/icha-senpai/note/third_party/forks/github/wmentor/html.svg)](https://pkg.go.dev/github.com/icha-senpai/note/third_party/forks/github/wmentor/html)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Simple HTML parser and data fetcher library written on Golang under MIT License.

## Require

* Golang (version >= 1.12)
* github.com/icha-senpai/note/third_party/forks/external/golang.org/x/net

## Install

```
go get github.com/icha-senpai/note/third_party/forks/github/wmentor/html
```

## Usage

### Fetch data from URL

```golang
package main

import (
  "fmt"
  "time"

  "github.com/icha-senpai/note/third_party/forks/github/wmentor/html"
)

func main() {

  src := "https://edition.cnn.com"

  parser := html.New()

  opts := &html.GetOpts{
    Agent:"Mozilla/5.0 (compatible; MSIE 10.0)",
    Timeout: time.Second*60,
  }

  parser.Get(src,opts)
  fmt.Println( string(parser.Text()) )

  parser.EachLink(func(link string) {
    fmt.Println("url=" + link)
  } )

  parser.EachImage(func(link string) {
    fmt.Println("img=" + link)
  } )

  parser.EachIframe(func(link string) {
    fmt.Println("iframe=" + link)
  } )
}
```

### Fetch data from file/stdin

```golang
package main

import (
  "fmt"
  "os"

  "github.com/icha-senpai/note/third_party/forks/github/wmentor/html"
)

func main() {

  parser := html.New()

  parser.Parse(os.Stdin) // io.Reader
  fmt.Println( string(parser.Text()) )

  parser.EachLink(func(link string) {
    fmt.Println("url=" + link)
  } )

  parser.EachImage(func(link string) {
    fmt.Println("img=" + link)
  } )

  parser.EachIframe(func(link string) {
    fmt.Println("iframe=" + link)
  } )
}
```
