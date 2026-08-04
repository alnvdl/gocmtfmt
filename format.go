package main

import (
	"bytes"
	"go/ast"
	"go/doc/comment"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
)

type replacement struct {
	start int
	end   int
	text  []byte
}

func formatFile(path string, tabWidth, columnWidth int) ([]byte, []byte, os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, nil, nil, err
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, err
	}
	formatted, err := formatSource(source, path, tabWidth, columnWidth)
	return source, formatted, info, err
}

func formatSource(source []byte, filename string, tabWidth, columnWidth int) ([]byte, error) {
	source, err := format.Source(source)
	if err != nil {
		return nil, err
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	groups := lineCommentGroups(file, fileSet, source)
	replacements := make([]replacement, 0, len(groups))
	for _, group := range groups {
		start, end := commentOffsets(group, fileSet)
		linePrefix, indent := lineIndent(source, start)
		comment := stripIndent(source[start:end])
		formatted := formatComment(extractContent(comment), columnWidth-indentWidth(linePrefix, tabWidth))
		formatted = bytes.TrimSuffix(formatted, []byte("\n"))
		continuationIndent := indent
		if indent == nil {
			continuationIndent = makeIndent(indentWidth(linePrefix, tabWidth), tabWidth)
		}
		formatted = prependIndent(formatted, continuationIndent)
		replacements = append(replacements, replacement{start: start, end: end, text: formatted})
	}
	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].start < replacements[j].start
	})

	for _, replacement := range slices.Backward(replacements) {
		source = applyReplacement(source, replacement)
	}
	return format.Source(source)
}

func commentOffsets(group *ast.CommentGroup, fileSet *token.FileSet) (start, end int) {
	start = fileSet.Position(group.Pos()).Offset
	end = fileSet.Position(group.End()).Offset
	return start, end
}

func lineIndent(source []byte, commentStart int) (linePrefix, indent []byte) {
	lineStart := commentStart
	for lineStart > 0 && source[lineStart-1] != '\n' {
		lineStart--
	}
	linePrefix = source[lineStart:commentStart]
	indent = linePrefix
	for _, char := range indent {
		if char != ' ' && char != '\t' {
			return linePrefix, nil
		}
	}
	return linePrefix, indent
}

func applyReplacement(source []byte, replacement replacement) []byte {
	return append(source[:replacement.start], append(replacement.text, source[replacement.end:]...)...)
}

func isDirective(comment *ast.Comment) bool {
	return len(comment.Text) >= 2 && comment.Text[:2] == "//" &&
		len(comment.Text) > 2 && comment.Text[2] != ' '
}

func lineCommentGroups(file *ast.File, fileSet *token.FileSet, source []byte) []*ast.CommentGroup {
	var groups []*ast.CommentGroup
	var current []*ast.Comment
	flush := func() {
		if len(current) > 0 {
			groups = append(groups, &ast.CommentGroup{List: current})
			current = nil
		}
	}

	for _, group := range file.Comments {
		for _, comment := range group.List {
			if isLineComment(comment) {
				if len(current) > 0 && !commentsAreAdjacent(current[len(current)-1], comment, fileSet, source) {
					flush()
				}
				current = append(current, comment)
				continue
			}
			flush()
		}
	}
	flush()
	return groups
}

func isLineComment(comment *ast.Comment) bool {
	return strings.HasPrefix(comment.Text, "//") && !isDirective(comment)
}

func commentsAreAdjacent(previous, current *ast.Comment, fileSet *token.FileSet, source []byte) bool {
	previousPosition := fileSet.Position(previous.End())
	currentPosition := fileSet.Position(current.Pos())
	if previousPosition.Line+1 != currentPosition.Line {
		return false
	}
	between := source[previousPosition.Offset:currentPosition.Offset]
	return strings.TrimSpace(string(between)) == ""
}

func prependIndent(comment, indent []byte) []byte {
	lines := bytes.Split(comment, []byte("\n"))
	for i := 1; i < len(lines); i++ {
		lines[i] = append(append([]byte{}, indent...), lines[i]...)
	}
	return bytes.Join(lines, []byte("\n"))
}

func makeIndent(width, tabWidth int) []byte {
	indent := bytes.Repeat([]byte("\t"), width/tabWidth)
	return append(indent, bytes.Repeat([]byte(" "), width%tabWidth)...)
}

func stripIndent(comment []byte) []byte {
	lines := bytes.Split(comment, []byte("\n"))
	for i := 1; i < len(lines); i++ {
		lines[i] = bytes.TrimLeft(lines[i], " \t")
	}
	return bytes.Join(lines, []byte("\n"))
}

func indentWidth(indent []byte, tabWidth int) int {
	width := 0
	for len(indent) > 0 {
		char, size := utf8.DecodeRune(indent)
		if char == '\t' {
			width += tabWidth
		} else {
			width++
		}
		indent = indent[size:]
	}
	return width
}

func extractContent(cmt []byte) *comment.Doc {
	lines := strings.Split(string(cmt), "\n")
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], "// ")
		lines[i] = strings.TrimPrefix(lines[i], "//")
	}
	cleanCmt := strings.Join(lines, "\n")
	return (&comment.Parser{}).Parse(cleanCmt)
}
