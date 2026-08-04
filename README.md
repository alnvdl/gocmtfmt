# gocmtfmt

[![Go Reference](https://pkg.go.dev/badge/github.com/alnvdl/gocmtfmt.svg)](https://pkg.go.dev/github.com/alnvdl/gocmtfmt)
[![Test workflow](https://github.com/alnvdl/gocmtfmt/actions/workflows/test.yaml/badge.svg)](https://github.com/alnvdl/gocmtfmt/actions/workflows/test.yaml)

`gocmtfmt` formats and wraps `//` comments in Go code. It mainly reflows
paragraphs and lists while preserving headings, code blocks, and other
elements. It formats, but does **not** wrap regular code.

`gocmtfmt` has three goals:
1. Make comments more readable by limiting the number of columns.
2. Reduce surprises by making comments in code look more like the output of
   tools such as `go doc` and `pkgsite` (see [Notes](#Notes)).
3. Keep codebases more consistent by introducing an opinionated and unambiguous
   way of formatting comments.

## Using

Install it with:
```sh
go install github.com/alnvdl/gocmtfmt@latest
```

Then run it with:
```sh
gocmtfmt -w .       # Recursively formats all Go files in a directory.
gocmtfmt -w file.go # Formats a single file.
```

## Comparison with gofmt

`gocmtfmt` complements `gofmt`, it is not a replacement. It applies the same
formatting as `gofmt` before and after processing each input file, but you can
still continue to use `gofmt` in your pipeline if you like.

`gocmtfmt` supports the `-l` and `-w` flags like `gofmt`, and it also accepts a
column width (`-c`, defaults to 79) and a tab size (`-t`, defaults to 4).

Unlike `gofmt`/`go fmt`:
- it formats all `//` comment blocks as
  [Go doc comments](https://go.dev/doc/comment), not just those tied to certain
  language constructs.
- it does not format files marked with `DO NOT EDIT` in their first line.
- it does not support the `./...` syntax used by the native `go fmt` tool.
  Instead, it works like plain `gofmt`.

## Notes

`gocmtfmt` is based on a slightly modified version of the standard library's
[text comment printer](https://pkg.go.dev/go/doc/comment), which uses an
algorithm for nicely reflowing paragraphs without necessarily making use of all
available space.

This tool is opinionated in the following ways:

1. `/* ... */` comment blocks are intentionally ignored, so users can keep them
   for things that should not be formatted. This is useful, for example, when
   commenting out code blocks. It also means your code must use // comments for
   the majority of things.

2. Unused [link definitions](https://go.dev/doc/comment#links) will be removed
   from comments.

2. Sequential lines without line breaks and poorly formatted lists are joined.
   Consider the following comment block:
   ```go
   // FunctionA accepts two types of input.
   // Input A causes it to do one thing.
   // Input B causes it to do something else.
   //
   // - Not really a list item.
   // - Because there's no indent before the bullets.
   ```

   It gets formatted and wrapped as:
   ```go
   // FunctionA accepts two types of input. Input A causes it to do one thing.
   // Input B causes it to do something else.
   //
   // - Not really a list item. - Because there's no indent before the bullets.
   ```

   This is because `gocmtfmt` formats documentation similarly to how `go doc`
   and `pkgsite` present it. See [Go Doc Comments](https://go.dev/doc/comment)
   to learn how to format your comments the Go way.

3. Trailing comments are not formatted or wrapped:
   ```go
	var value = 1 // This comment will be ignored even though it is longer than the column limit.
   ```

   Thisi s because Go's default formatting does not deal with multiline
   trailing comments, indenting subsequent lines as unrelated comment blocks.
