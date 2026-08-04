package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatSourceAndFile(t *testing.T) {
	tests := []struct {
		desc            string
		tabWidth        int
		columnWidth     int
		source          string
		want            string
		wantSourceError string
		filePath        func(dir string) string
		wantFileError   func(path string) string
	}{{
		desc:     "formats ordinary line comment blocks",
		tabWidth: 4,
		source: `
package sample

func Example() {}

// This ordinary comment is long enough to require wrapping even though it is not attached to a declaration.
var value = 1
`,
		want: `
package sample

func Example() {}

// This ordinary comment is long enough to require wrapping even though it is
// not attached to a declaration.
var value = 1
`,
	}, {
		desc:     "does not format a single trailing comment",
		tabWidth: 4,
		source: `
package sample

func Example() {} // This should be ignored because it is just a comment after a line.
`,
		want: `
package sample

func Example() {} // This should be ignored because it is just a comment after a line.
`,
	}, {
		desc:     "does not format multiple trailing comments",
		tabWidth: 4,
		source: `
package sample

func Example() {} // this is something quite long that goes beyond the column limit.
func Example2() {} // this is something else.
func Example3() {} // yet one more.
`,
		want: `
package sample

func Example()  {} // this is something quite long that goes beyond the column limit.
func Example2() {} // this is something else.
func Example3() {} // yet one more.
`,
	}, {
		desc:     "handles multi-line trailing comments",
		tabWidth: 4,
		source: `
package sample

func Example() {
	var value = 1 // First line of a trailing comment that is quite long indeed and exceeds the column limit.
	              // Second line of the trailing comment, also quite long here, passing the column limit.
	_ = value
}
`,
		want: `
package sample

func Example() {
	var value = 1 // First line of a trailing comment that is quite long indeed and exceeds the column limit.
	// Second line of the trailing comment, also quite long here, passing the
	// column limit.
	_ = value
}
`,
	}, {
		desc:     "preserves directives",
		tabWidth: 4,
		source: `
//go:generate go run example.go
//line:original.go:10
// Package abc does something.
//No space after the marker is insignificant here.
package sample

//go:generate go run example.go
`,
		want: `
// Package abc does something. No space after the marker is insignificant here.
//
//go:generate go run example.go
//line:original.go:10
package sample

//go:generate go run example.go
`,
	}, {
		desc:     "ignores block comments",
		tabWidth: 4,
		source: `
package sample

/* This block comment is deliberately long and should not be reformatted even though it exceeds the usual line width. */

// This line comment should be formatted because it uses the supported comment style and is not documentation.
var value = 1
`,
		want: `
package sample

/* This block comment is deliberately long and should not be reformatted even though it exceeds the usual line width. */

// This line comment should be formatted because it uses the supported comment
// style and is not documentation.
var value = 1
`,
	}, {
		desc:     "formats line comments adjacent to block comments",
		tabWidth: 4,
		source: `
package sample

/* Random comment that is very long and should be wrapped but won't because of the comment block style */
// Very long line comment that should be wrapped even though it is adjacent to a block comment.

// Very long line comment that should be wrapped even though it is adjacent to a block comment.
/* Random comment that is very long and should be wrapped but won't because of the comment block style */

/* Keep this block comment exactly as written. */
// This line comment is long enough to wrap and must not be skipped because it shares a parser comment group with a block comment.
var value = 1
`,
		want: `
package sample

/* Random comment that is very long and should be wrapped but won't because of the comment block style */
// Very long line comment that should be wrapped even though it is adjacent to
// a block comment.

// Very long line comment that should be wrapped even though it is adjacent to
// a block comment.
/* Random comment that is very long and should be wrapped but won't because of the comment block style */

/* Keep this block comment exactly as written. */
// This line comment is long enough to wrap and must not be skipped because it
// shares a parser comment group with a block comment.
var value = 1
`,
	}, {
		desc:     "formats package and indented comments",
		tabWidth: 4,
		source: `
// Package sample contains a deliberately long package comment that should be wrapped while retaining valid Go syntax.
package sample

// This is an unrelated comment and should stay exactly where it is.

func Example() {}

type Thing struct {
	// This comment is indented and should be wrapped to the remaining width available on each line.
	field int
}
`,
		want: `
// Package sample contains a deliberately long package comment that should be
// wrapped while retaining valid Go syntax.
package sample

// This is an unrelated comment and should stay exactly where it is.

func Example() {}

type Thing struct {
	// This comment is indented and should be wrapped to the remaining width
	// available on each line.
	field int
}
`,
	}, {
		desc:     "correctly applies indentation to multiple lines in complex cases",
		tabWidth: 4,
		source: `
package sample

type replacement struct {
	// start and end are the byte offsets in the original source file that
	// should be replaced with text. This is a list:
	//   - Item 1
	//   - Item 2 is much longer and requires line breaks because it exceeds the maximum allowed line length for comment content.
	//   - Item 3
	start int
}
`,
		want: `
package sample

type replacement struct {
	// start and end are the byte offsets in the original source file that
	// should be replaced with text. This is a list:
	//   - Item 1
	//   - Item 2 is much longer and requires line breaks because it exceeds
	//     the maximum allowed line length for comment content.
	//   - Item 3
	start int
}
`,
	}, {
		desc:     "comment unchanged when tab width is not long enough to wrap",
		tabWidth: 4,
		source: `
package sample

type sample struct {
	// This comment fits with four tabs but wraps with eight tabs in practice.
	field int
}
`,
		want: `
package sample

type sample struct {
	// This comment fits with four tabs but wraps with eight tabs in practice.
	field int
}
`,
	}, {
		desc:        "comment wraps with configured column width",
		tabWidth:    8,
		columnWidth: 50,
		source: `
package sample

	// This comment is long enough to wrap at the configured column width.
	var value int
`,
		want: `
package sample

// This comment is long enough to wrap at the
// configured column width.
var value int
`,
	}, {
		desc:     "formats comment blocks containing blank comment lines",
		tabWidth: 4,
		source: `
package sample

// This first paragraph is long enough that it really ought to be wrapped by the tool.
//
// This second paragraph is also long enough that it really ought to be wrapped too.
var value = 1
`,
		want: `
package sample

// This first paragraph is long enough that it really ought to be wrapped by
// the tool.
//
// This second paragraph is also long enough that it really ought to be wrapped
// too.
var value = 1
`,
	}, {
		desc:     "formats comment blocks containing code blocks",
		tabWidth: 4,
		source: `
package sample

// This prose paragraph is long enough that it really ought to be wrapped by the tool.
//
//	code()
//
// And this trailing paragraph is also long enough that it ought to be wrapped.
var value = 1
`,
		want: `
package sample

// This prose paragraph is long enough that it really ought to be wrapped by
// the tool.
//
//	code()
//
// And this trailing paragraph is also long enough that it ought to be wrapped.
var value = 1
`,
	}, {
		desc:     "formats prose that precedes a directive",
		tabWidth: 4,
		source: `
package sample

// This prose comment is long enough that it really ought to be wrapped by the tool.
//go:generate go run example.go
func Example() {}
`,
		want: `
package sample

// This prose comment is long enough that it really ought to be wrapped by the
// tool.
//
//go:generate go run example.go
func Example() {}
`,
	}, {
		desc:     "removes a comment holding nothing but a space",
		tabWidth: 4,
		source: `
package sample

//
var value = 1
`,
		want: `
package sample

var value = 1
`,
	}, {
		desc:     "removes carriage returns in CRLF sources",
		tabWidth: 4,
		source:   "package sample\r\n\r\n// This comment is long enough that it really ought to be wrapped by the tool here.\r\nvar value = 1\r\n",
		want:     "package sample\n\n// This comment is long enough that it really ought to be wrapped by the tool\n// here.\nvar value = 1\n",
	}, {
		desc:     "handles continuation lines indented differently from the first line",
		tabWidth: 4,
		source: `
package sample

func Example() {
	// This comment line is indented with a tab and is long enough to be wrapped.
    // This continuation line was indented with spaces instead of a tab character.
	_ = 1
}
`,
		want: `
package sample

func Example() {
	// This comment line is indented with a tab and is long enough to be
	// wrapped. This continuation line was indented with spaces instead of a
	// tab character.
	_ = 1
}
`,
	}, {
		desc:            "formatSource reports invalid Go source",
		source:          "package sample\nfunc {",
		wantSourceError: "2:6: expected 'IDENT', found '{'",
	}, {
		desc: "formatFile reports a missing file",
		filePath: func(dir string) string {
			return filepath.Join(dir, "sample.go")
		},
		wantFileError: func(path string) string {
			return fmt.Sprintf("lstat %s: no such file or directory", path)
		},
	}, {
		desc: "formatFile reports an unreadable directory",
		filePath: func(dir string) string {
			return dir
		},
		wantFileError: func(path string) string {
			return fmt.Sprintf("read %s: is a directory", path)
		},
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			source := strings.TrimPrefix(test.source, "\n")
			want := strings.TrimPrefix(test.want, "\n")
			columnWidth := test.columnWidth
			if columnWidth == 0 {
				columnWidth = 79
			}

			t.Run("source", func(t *testing.T) {
				if test.wantFileError != nil {
					return
				}
				got, err := formatSource([]byte(source), "sample.go", test.tabWidth, columnWidth)
				if test.wantSourceError != "" {
					if err == nil || err.Error() != test.wantSourceError {
						t.Fatalf("formatSource error = %v, want %q", err, test.wantSourceError)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != want {
					t.Errorf("formatted source mismatch:\ngot:\n%s\nwant:\n%s", got, want)
				}

				again, err := formatSource(got, "sample.go", test.tabWidth, columnWidth)
				if err != nil {
					t.Fatal(err)
				}
				if string(again) != string(got) {
					t.Errorf("formatting is not idempotent:\nfirst:\n%s\nsecond:\n%s", got, again)
				}
			})

			t.Run("file", func(t *testing.T) {
				if test.wantSourceError != "" {
					return
				}
				dir := t.TempDir()
				path := filepath.Join(dir, "sample.go")
				if test.filePath != nil {
					path = test.filePath(dir)
				}
				if test.wantFileError != nil {
					_, _, _, err := formatFile(path, test.tabWidth, columnWidth)
					wantError := test.wantFileError(path)
					if err == nil || err.Error() != wantError {
						t.Fatalf("formatFile error = %v, want %q", err, wantError)
					}
					return
				}
				if err := os.WriteFile(path, []byte(source), 0644); err != nil {
					t.Fatal(err)
				}

				_, got, _, err := formatFile(path, test.tabWidth, columnWidth)
				if err != nil {
					t.Fatal(err)
				}
				unchanged, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if string(unchanged) != source {
					t.Fatalf("formatFile rewrote its input")
				}
				if string(got) != want {
					t.Errorf("formatted file mismatch:\ngot:\n%s\nwant:\n%s", got, want)
				}
			})
		})
	}
}
