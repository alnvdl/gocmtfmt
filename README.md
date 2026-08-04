# gocmtfmt

`gocmtfmt` wraps the content of all `//` Go comment blocks. It mainly reformats
paragraphs and lists, keeping titles, code blocks and other elements as they
are.

`/* ... */` comments blocks are intentionally ignored, as a way to allow for
users to have comment blocks that are not subject to reformatting. This can be
useful when commenting out code blocks for example.

`gocmtfmt` exists for three reasons:
1. Limiting the number of columns makes text more readable.
2. Reformatting makes comments in code look a lot more like what tools like
   `go doc` and `pkgsite` produce, leading to fewer surprises
   (see [Notes](#Notes)).
3. Manually managing line breaks and identifying offending files is a pain.

## Using

Install with:
```sh
go install github.com/alnvdl/gocmtfmt@latest
```

And run with
```sh
gocmtfmt -w file.go
```

Or to run in all files in a module:
```sh
find . -name '*.go' | xargs cmtfmt -w
```

`gocmtfmt` does not support the `./...` syntax used by native `go fmt` tooling;
instead, it mimics plain `gofmt`.

## Comparison to gofmt

`gocmtfmt` is meant to be used as a complement to `gofmt` not as a replacement:
it will invoke the same behavior of `gofmt` before and after on every input it
processes, and you can still keep `gofmt` in your pipeline.

`gocmtfmt` supports the `-l` and `-w` flags like `gofmt`, and it also accepts a
column width (`-c`, defaults to 79) and a tab size (`-t`, defaults to 4).

Differently from `gofmt`, it reformats all `//` comment blocks as
[Go doc comments](https://go.dev/doc/comment), not just those tied to certain
language constructs.

## Notes

1. Unused [link definitions](https://go.dev/doc/comment#links) will be removed
   from comments.

2. Fake paragraphs are joined. Consider the comment block below:
   ```go
   // FunctionA does accepts one of 3 types of input.
   // Input A will cause it to do something.
   // Input B will cause it to do something else.
   ```

   It will get reformatted as:
   ```go
   // FunctionA does accepts one of 3 types of input. Input A will cause it to
   // do something. Input B will cause it to do something else.
   ```

   This is by design, as `gocmtfmt` formats documentation exactly like `go doc`
   would display them in text mode. In fact, it is based on (a slightly
   modified version of) the text comment printer of the standard library:
   https://pkg.go.dev/go/doc/comment.
