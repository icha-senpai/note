# snappy

The Snappy compression format in the Go programming language.

This is a drop-in replacement for `github.com/icha-senpai/note/third_party/forks/github/golang/snappy`.

It provides a full, compatible replacement of the Snappy package by simply changing imports.

See [Snappy Compatibility](https://github.com/icha-senpai/note/third_party/forks/github/klauspost/compress/tree/master/s2#snappy-compatibility) in the S2 documentation.

"Better" compression mode is used. For buffered streams concurrent compression is used.

For more options use the [s2 package](https://pkg.go.dev/github.com/icha-senpai/note/third_party/forks/github/klauspost/compress/s2).

# usage

Replace imports `github.com/icha-senpai/note/third_party/forks/github/golang/snappy` with `github.com/icha-senpai/note/third_party/forks/github/klauspost/compress/snappy`.
