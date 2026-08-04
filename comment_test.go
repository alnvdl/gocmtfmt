package main

import (
	"strings"
	"testing"
)

func TestFormatComment(t *testing.T) {
	tests := []struct {
		desc string

		sample string
		want   string
		width  int
	}{{
		desc: "single line",
		sample: `
// This is a single line comment.
`,
		want: `
// This is a single line comment.
`,
	}, {
		desc: "preserves italic text",
		sample: `
// A *formatted* word.
`,
		want: `
// A *formatted* word.
`,
	}, {
		desc: "prefers punctuation when reflowing an existing line break",
		sample: `
// This is just a bunch of senseless words, we are interested in a thing, which
// is verifying that the line break has a bias for punctuation: even though
// "which" would fit in the first line, it gets moves to the next line.
`,
		want: `
// This is just a bunch of senseless words, we are interested in a thing,
// which is verifying that the line break has a bias for punctuation: even
// though "which" would fit in the first line, it gets moves to the next line.
`,
	}, {
		desc: "single paragraph",
		sample: `
// This is a single line comment.
// And this line is part of the same paragraph.
`,
		want: `
// This is a single line comment. And this line is part of the same paragraph.
`,
	}, {
		desc: "a list without any surrounding content is treated as a paragraph",
		sample: `
//   - Item 1
//   - Item 2 is long and will cause the line to break for item 3
//   - Item 3
`,
		want: `
// - Item 1 - Item 2 is long and will cause the line to break for item 3 - Item
// 3
`,
	}, {
		desc: "a list with surrounding content is treated as a list",
		sample: `
// This is a list:
//   - Item 1
//   - Item 2 is much longer and requires line breaks because it exceeds the maximum allowed line length for comment content.
//   - Item 3
`,
		want: `
// This is a list:
//   - Item 1
//   - Item 2 is much longer and requires line breaks because it exceeds the
//     maximum allowed line length for comment content.
//   - Item 3
`,
	}, {
		desc: "preserves blank lines between list items",
		sample: `
// This is a list:
//   - Item 1
//
//   - Item 2
`,
		want: `
// This is a list:
//
//   - Item 1
//
//   - Item 2
`,
	}, {
		desc: "preserves multiple paragraphs in a list item",
		sample: `
// This is a list:
//   - First paragraph.
//
//     Second paragraph.
`,
		want: `
// This is a list:
//
//   - First paragraph.

//     Second paragraph.
`,
	}, {
		desc: "wraps varied word lengths through comment formatting",
		sample: `
// Alpha bravo charlie delta echo foxtrot golf hotel india juliett kilo lima mike november oscar papa quebec.
`,
		want: `
// Alpha bravo charlie delta echo foxtrot golf hotel india juliett kilo lima
// mike november oscar papa quebec.
`,
		width: 79,
	}, {
		desc: "code blocks do not have their lines broken",
		sample: `
//	func example(objects []object) {
//		for index, object in reallyLongFunctionToEnumerateObjectsPassedAsArguments(objects) {
//			// This is a comment inside the code block. It should not be broken into multiple lines.
//			fmt.Println(object)
//		}
//	}
`,
		want: `
//	func example(objects []object) {
//		for index, object in reallyLongFunctionToEnumerateObjectsPassedAsArguments(objects) {
//			// This is a comment inside the code block. It should not be broken into multiple lines.
//			fmt.Println(object)
//		}
//	}
`,
	}, {
		desc: "complex example",
		sample: `
// This is a very long comment that spans multiple lines.
// It is used to demonstrate how to handle long comments in Go code. The comment should be properly formatted and should not cause any issues with the code structure.
// This is just an example and does not contain any meaningful information. This is a code block:
//
//	func example() {
//		// This is an example function.
//		// Code blocks can exceed the maximum line width and should be left alone.
//		fmt.Println("Hello, World!")
//	}
//
// And this is a list:
// 	- Item 1
// 	- Item 2 with more details and explanations that go on for a while to show how to handle long lines in lists.
// 	- Item 3
//
// # Headings can also be present
//
// Short comments in quick succession.
// Are combined into a single paragraph.
// If they don't have line breaks between them.
//
// Overall, this comment is meant to illustrate how to manage long comments in Go code without breaking the flow of the code or making it difficult to read, including when it has references like this: [fmt.Println].
//
// See more details in [Go Doc comments].
//
// Direct links should be properly handled too if they are too long: https://example.com/this-is-a-very-long-url-that-should-be-broken-into-another-line-to-fit-within-the-specified-width-limit
//
// Testing when links are short: https://example.com/short-url
//
// See also indirect links: [Go Doc comments with a much longer link name to trigger breaks everywhere it gets used]
//
// Paragraph at the limit - just a bunch of random words until it hits the end.
//
// Paragraph past the limit - just a bunch of random words until it hits the end and...
//
// Also, this is a reference to a link that should be left alone as: [fmt.What].
//
// A complicated code block from the Match function in the standard library:
//
//	pattern:
//		{ term }
//	term:
//		'*'         matches any sequence of non-/ characters
//		'?'         matches any single non-/ character
//		'[' [ '^' ] { character-range } ']'
//		            character class (must be non-empty)
//		c           matches character c (c != '*', '?', '\\', '[')
//		'\\' c      matches character c
//
//	character-range:
//		c           matches character c (c != '\\', '-', ']')
//		'\\' c      matches character c
//		lo '-' hi   matches character c for lo <= c <= hi
//
// Match requires pattern to match all of name, not just a substring.
// The only possible returned error is [ErrBadPattern], when pattern
// is malformed.
//
// [Go Doc comments]: https://go.dev/doc/comment
// [Go Doc comments with a much longer link name to trigger breaks everywhere it gets used]: https://go.dev/doc/comment
`,
		want: `
// This is a very long comment that spans multiple lines. It is used to
// demonstrate how to handle long comments in Go code. The comment should be
// properly formatted and should not cause any issues with the code structure.
// This is just an example and does not contain any meaningful information.
// This is a code block:
//
//	func example() {
//		// This is an example function.
//		// Code blocks can exceed the maximum line width and should be left alone.
//		fmt.Println("Hello, World!")
//	}
//
// And this is a list:
//   - Item 1
//   - Item 2 with more details and explanations that go on for a while to show
//     how to handle long lines in lists.
//   - Item 3
//
// # Headings can also be present
//
// Short comments in quick succession. Are combined into a single paragraph.
// If they don't have line breaks between them.
//
// Overall, this comment is meant to illustrate how to manage long comments
// in Go code without breaking the flow of the code or making it difficult to
// read, including when it has references like this: [fmt.Println].
//
// See more details in [Go Doc comments].
//
// Direct links should be properly handled too if they are too long:
// https://example.com/this-is-a-very-long-url-that-should-be-broken-into-another-line-to-fit-within-the-specified-width-limit
//
// Testing when links are short: https://example.com/short-url
//
// See also indirect links: [Go Doc comments with a much longer link name to
// trigger breaks everywhere it gets used]
//
// Paragraph at the limit - just a bunch of random words until it hits the end.
//
// Paragraph past the limit - just a bunch of random words until it hits the
// end and...
//
// Also, this is a reference to a link that should be left alone as:
// [fmt.What].
//
// A complicated code block from the Match function in the standard library:
//
//	pattern:
//		{ term }
//	term:
//		'*'         matches any sequence of non-/ characters
//		'?'         matches any single non-/ character
//		'[' [ '^' ] { character-range } ']'
//		            character class (must be non-empty)
//		c           matches character c (c != '*', '?', '\\', '[')
//		'\\' c      matches character c
//
//	character-range:
//		c           matches character c (c != '\\', '-', ']')
//		'\\' c      matches character c
//		lo '-' hi   matches character c for lo <= c <= hi
//
// Match requires pattern to match all of name, not just a substring. The only
// possible returned error is [ErrBadPattern], when pattern is malformed.
//
// [Go Doc comments]: https://go.dev/doc/comment
// [Go Doc comments with a much longer link name to trigger breaks everywhere it gets used]: https://go.dev/doc/comment
`,
	}, {
		desc: "a blank comment line is preserved",
		sample: `
//
`,
		want: `
//
`,
	}, {
		desc: "allows a width narrower than the comment prefix",
		sample: `
// short text
`,
		want: `
// short text
`,
		width: 2,
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Remove the first empty line from the comments: they are just
			// used to make the test source code more readable.
			sample := strings.TrimPrefix(test.sample, "\n")
			want := strings.TrimPrefix(test.want, "\n")

			width := test.width
			if width == 0 {
				width = 79
			}
			got := string(formatComment(extractContent([]byte(sample)), width))
			if got != want {
				t.Errorf("formatComment(`%s`)\n\ngot=`%s`\n\nwant=`%s`", sample, got, want)
			}
		})
	}
}
