package main

import (
	"bytes"
	"go/ast"
	"go/doc/comment"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"unicode/utf8"
)

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
	if hasDoNotEditMarker(source) {
		return source, nil
	}

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
	formatted := formatComments(source, fileSet, groups, tabWidth, columnWidth)
	return format.Source(formatted)
}

func hasDoNotEditMarker(source []byte) bool {
	firstLine := source
	if before, _, ok := bytes.Cut(source, []byte{'\n'}); ok {
		firstLine = before
	}
	return bytes.Contains(firstLine, []byte("DO NOT EDIT"))
}

func formatComments(source []byte, fileSet *token.FileSet, groups []*ast.CommentGroup, tabWidth, columnWidth int) []byte {
	var formatted bytes.Buffer
	lastEnd := 0
	for _, group := range groups {
		start, end := commentOffsets(group, fileSet)
		formatted.Write(source[lastEnd:start])
		formatted.Write(formatCommentGroup(source, group, fileSet, tabWidth, columnWidth))
		lastEnd = end
	}
	formatted.Write(source[lastEnd:])
	return formatted.Bytes()
}

func formatCommentGroup(source []byte, group *ast.CommentGroup, fileSet *token.FileSet, tabWidth, columnWidth int) []byte {
	start, end := commentOffsets(group, fileSet)
	linePrefix, indent := lineIndent(source, start)
	formatted := formatComment(
		extractContent(stripIndent(source[start:end])),
		columnWidth-indentWidth(linePrefix, tabWidth),
	)
	formatted = bytes.TrimSuffix(formatted, []byte("\n"))
	if indent == nil {
		indent = makeIndent(indentWidth(linePrefix, tabWidth), tabWidth)
	}
	return prependIndent(formatted, indent)
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

func lineCommentGroups(file *ast.File, fileSet *token.FileSet, source []byte) []*ast.CommentGroup {
	var groups []*ast.CommentGroup
	var current *ast.CommentGroup

	for _, group := range file.Comments {
		for _, comment := range group.List {
			if !isLineComment(comment, fileSet, source) {
				current = nil
				continue
			}
			if current == nil || !commentsAreAdjacent(current.List[len(current.List)-1], comment, fileSet, source) {
				current = &ast.CommentGroup{}
				groups = append(groups, current)
			}
			current.List = append(current.List, comment)
		}
	}
	return groups
}

func isLineComment(comment *ast.Comment, fileSet *token.FileSet, source []byte) bool {
	if !strings.HasPrefix(comment.Text, "//") {
		return false
	}
	if _, ok := ast.ParseDirective(comment.Slash, comment.Text); ok {
		return false
	}
	_, indent := lineIndent(source, fileSet.Position(comment.Pos()).Offset)
	return indent != nil
}

func commentsAreAdjacent(previous, current *ast.Comment, fileSet *token.FileSet, source []byte) bool {
	previousPosition := fileSet.Position(previous.End())
	currentPosition := fileSet.Position(current.Pos())
	if previousPosition.Line+1 != currentPosition.Line {
		return false
	}
	between := source[previousPosition.Offset:currentPosition.Offset]
	return len(bytes.TrimSpace(between)) == 0
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
