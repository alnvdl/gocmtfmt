package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStdin(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "formats input",
			args: []string{"-c", "50"},
			want: "package sample\n\n// This comment is long enough to wrap at a narrow\n// width.\nvar value int\n",
		},
		{
			name: "lists changed input without writing output",
			args: []string{"-l", "-c", "50"},
			// want is empty because the input is from stdin, so there is no
			// path to list.
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setStreams(t, "package sample\n\n// This comment is long enough to wrap at a narrow width.\nvar value int\n")

			if status := run(test.args); status != 0 {
				t.Fatalf("run() status = %d, want 0", status)
			}
			if got := readStream(t, stdout); got != test.want {
				t.Errorf("stdout = %q, want %q", got, test.want)
			}
		})
	}
}

func TestMainExitCodes(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantExitCalled bool
		wantStatus     int
	}{{
		name:           "success",
		args:           []string{"-c", "50"},
		wantExitCalled: false,
		wantStatus:     0,
	}, {
		name:           "runner error",
		args:           []string{"missing.go"},
		wantExitCalled: true,
		wantStatus:     1,
	}, {
		name:           "usage error",
		args:           []string{"-c", "0"},
		wantExitCalled: true,
		wantStatus:     2,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setStreams(t, "package sample\n")
			oldArgs, oldExit := os.Args, exit
			t.Cleanup(func() {
				os.Args, exit = oldArgs, oldExit
			})
			os.Args = append([]string{"gocmtfmt"}, test.args...)
			exitCalled := false
			gotStatus := 0
			exit = func(status int) {
				exitCalled = true
				gotStatus = status
			}

			main()

			if exitCalled != test.wantExitCalled {
				t.Errorf("exit called = %t, want %t", exitCalled, test.wantExitCalled)
			}
			if gotStatus != test.wantStatus {
				t.Errorf("exit status = %d, want %d", gotStatus, test.wantStatus)
			}
		})
	}
}

func TestRunFormatsPaths(t *testing.T) {
	setStreams(t, "")
	path := writeTempSource(t, "package sample\n\n// This comment is long enough to wrap at a narrow width.\nvar value int\n")

	if status := run([]string{"-c", "50", path}); status != 0 {
		t.Fatalf("run() status = %d, want 0", status)
	}
	if got := readStream(t, stdout); !strings.Contains(got, "// This comment is long enough to wrap at a narrow\n") {
		t.Errorf("stdout does not contain formatted source: %q", got)
	}

	resetStream(t, stdout)
	if status := run([]string{"-l", "-c", "50", path}); status != 0 {
		t.Fatalf("run() status = %d, want 0", status)
	}
	if got := readStream(t, stdout); got != path+"\n" {
		t.Errorf("listed paths = %q, want %q", got, path+"\n")
	}

	resetStream(t, stdout)
	if status := run([]string{"-w", "-c", "50", path}); status != 0 {
		t.Fatalf("run() status = %d, want 0", status)
	}
	if got := readFile(t, path); !strings.Contains(got, "// This comment is long enough to wrap at a narrow\n") {
		t.Errorf("written source is not formatted: %q", got)
	}
}

func TestRunFormatsDirectoryRecursively(t *testing.T) {
	setStreams(t, "")
	dir := t.TempDir()
	nestedDir := filepath.Join(dir, "nested")
	if err := os.Mkdir(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	rootPath := filepath.Join(dir, "root.go")
	nestedPath := filepath.Join(nestedDir, "nested.go")
	unchangedPath := filepath.Join(dir, "unchanged.go")
	nonGoPath := filepath.Join(nestedDir, "notes.txt")
	for path, source := range map[string]string{
		rootPath:      "package sample\n\n// This comment is long enough to wrap at a narrow width.\nvar value int\n",
		nestedPath:    "package sample\n\n// This comment is long enough to wrap at a narrow width.\nvar value int\n",
		unchangedPath: "package sample\n\nvar value int\n",
		nonGoPath:     "not Go\n",
	} {
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if status := run([]string{"-l", "-c", "50", "-t", "8", dir}); status != 0 {
		t.Fatalf("run() status = %d, want 0", status)
	}
	wantListed := nestedPath + "\n" + rootPath + "\n"
	if got := readStream(t, stdout); got != wantListed {
		t.Errorf("listed paths = %q, want %q", got, wantListed)
	}

	resetStream(t, stdout)
	if status := run([]string{"-w", "-c", "50", "-t", "8", dir}); status != 0 {
		t.Fatalf("run() status = %d, want 0", status)
	}
	if got := readStream(t, stdout); got != "" {
		t.Errorf("stdout = %q, want empty output", got)
	}
	wantContents := map[string]string{
		rootPath:      "package sample\n\n// This comment is long enough to wrap at a narrow\n// width.\nvar value int\n",
		nestedPath:    "package sample\n\n// This comment is long enough to wrap at a narrow\n// width.\nvar value int\n",
		unchangedPath: "package sample\n\nvar value int\n",
		nonGoPath:     "not Go\n",
	}
	for path, want := range wantContents {
		if got := readFile(t, path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
}

func TestRunReportsDirectoryWalkErrors(t *testing.T) {
	setStreams(t, "")
	path := filepath.Join(t.TempDir(), "missing")

	if status := run([]string{path}); status != 1 {
		t.Fatalf("run() status = %d, want 1", status)
	}
	if got, want := readStream(t, stderr), "lstat "+path+": no such file or directory\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestRunReportsCommandLineErrors(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStatus int
		want       string
	}{
		{name: "invalid width", args: []string{"-c", "0"}, wantStatus: 2, want: "error: -c and -t must be greater than zero\n"},
		{name: "write stdin", args: []string{"-w"}, wantStatus: 2, want: "error: cannot use -w with standard input\n"},
		{name: "missing file", args: []string{"missing.go"}, wantStatus: 1, want: "missing.go:"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setStreams(t, "package sample\n")
			if status := run(test.args); status != test.wantStatus {
				t.Fatalf("run() status = %d, want %d", status, test.wantStatus)
			}
			if got := readStream(t, stderr); !strings.Contains(got, test.want) {
				t.Errorf("stderr = %q, want substring %q", got, test.want)
			}
		})
	}
}

func setStreams(t *testing.T, input string) {
	t.Helper()
	oldStdin, oldStdout, oldStderr := stdin, stdout, stderr
	inputFile := writeTempFile(t, input)
	stdin = inputFile
	stdout = newTempFile(t)
	stderr = newTempFile(t)
	t.Cleanup(func() {
		stdin, stdout, stderr = oldStdin, oldStdout, oldStderr
	})
}

func writeTempSource(t *testing.T, source string) string {
	t.Helper()
	file := writeTempFile(t, source)
	path := file.Name()
	file.Close()
	return path
}

func writeTempFile(t *testing.T, content string) *os.File {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "gocmtfmt-test-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	return file
}

func newTempFile(t *testing.T) *os.File {
	t.Helper()
	return writeTempFile(t, "")
}

func resetStream(t *testing.T, file *os.File) {
	t.Helper()
	if err := file.Truncate(0); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
}

func readStream(t *testing.T, file *os.File) string {
	t.Helper()
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
