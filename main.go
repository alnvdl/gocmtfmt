package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var exit = os.Exit
var stdin = os.Stdin
var stderr = os.Stderr
var stdout = os.Stdout

const (
	defaultTabWidth    = 4
	defaultColumnWidth = 79
)

const (
	exitSuccess = 0
	exitError   = 1
	exitUsage   = 2
)

func main() {
	if status := run(os.Args[1:]); status != 0 {
		exit(status)
	}
}

func run(args []string) int {
	flags := flag.NewFlagSet("gocmtfmt", flag.ContinueOnError)
	flags.SetOutput(stderr)

	list := flags.Bool("l", false, "list files that would be reformatted")
	write := flags.Bool("w", false, "write result to source files instead of stdout")
	columnWidth := flags.Int("c", defaultColumnWidth, "maximum column width")
	tabWidth := flags.Int("t", defaultTabWidth, "tab size")

	if err := flags.Parse(args); err != nil {
		return exitUsage
	}
	if *columnWidth <= 0 || *tabWidth <= 0 {
		fmt.Fprintln(stderr, "error: -c and -t must be greater than zero")
		return exitUsage
	}

	var err error
	if flags.NArg() == 0 {
		if *write {
			fmt.Fprintln(stderr, "error: cannot use -w with standard input")
			return exitUsage
		}
		err = formatStdin(*list, *columnWidth, *tabWidth)
	} else {
		err = formatPaths(flags.Args(), *list, *write, *columnWidth, *tabWidth)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitError
	}
	return exitSuccess
}

func formatStdin(list bool, columnWidth, tabWidth int) error {
	source, err := io.ReadAll(stdin)
	if err != nil {
		return err
	}
	formatted, err := formatSource(source, "stdin", tabWidth, columnWidth)
	if err != nil {
		return err
	}
	if !list {
		_, err = stdout.Write(formatted)
	}
	return err
}

func formatPaths(paths []string, list, write bool, columnWidth, tabWidth int) error {
	for _, root := range paths {
		if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			switch {
			case err != nil:
				return err
			case entry.IsDir():
				return nil
			case path != root && !isGoFilename(entry.Name()):
				return nil
			}
			return formatPath(path, list, write, columnWidth, tabWidth)
		}); err != nil {
			return err
		}
	}
	return nil
}

func formatPath(path string, list, write bool, columnWidth, tabWidth int) error {
	source, formatted, info, err := formatFile(path, tabWidth, columnWidth)
	if err != nil {
		return err
	}
	changed := !bytes.Equal(source, formatted)
	if list && changed {
		if _, err := fmt.Fprintln(stdout, path); err != nil {
			return err
		}
	}
	if write && changed {
		if err := os.WriteFile(path, formatted, info.Mode()); err != nil {
			return err
		}
	}
	if !list && !write {
		if _, err := stdout.Write(formatted); err != nil {
			return err
		}
	}
	return nil
}

func isGoFilename(name string) bool {
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}
