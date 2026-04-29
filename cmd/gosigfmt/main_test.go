package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_StdinPrint(t *testing.T) {
	in := bytes.NewBufferString("package p\nfunc f(a int, b int) error { return nil }\n")
	var out, errBuf bytes.Buffer
	exit := run([]string{"gosigfmt"}, in, &out, &errBuf)
	if exit != 0 {
		t.Errorf("exit: want 0, got %d. stderr=%s", exit, errBuf.String())
	}
	if !strings.Contains(out.String(), "func f(a int, b int) error") {
		t.Errorf("unexpected output:\n%s", out.String())
	}
}

func TestRun_ListMode_NeedsFormatting(t *testing.T) {
	dir := t.TempDir()
	long := strings.Repeat("a", 60)
	src := "package p\nfunc f(a int, " + long + " int, b int, c int, d int, e int) error { return nil }\n"
	path := filepath.Join(dir, "x.go")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	exit := run([]string{"gosigfmt", "-l", path}, nil, &out, &errBuf)
	if exit != 2 {
		t.Errorf("exit: want 2, got %d. stderr=%s", exit, errBuf.String())
	}
	if !strings.Contains(out.String(), "x.go") {
		t.Errorf("expected file path in stdout, got: %q", out.String())
	}
}

func TestRun_WriteMode(t *testing.T) {
	dir := t.TempDir()
	long := strings.Repeat("a", 60)
	src := "package p\nfunc f(a int, " + long + " int, b int, c int, d int) error { return nil }\n"
	path := filepath.Join(dir, "x.go")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	exit := run([]string{"gosigfmt", "-w", path}, nil, &out, &errBuf)
	if exit != 0 {
		t.Errorf("exit: want 0, got %d. stderr=%s", exit, errBuf.String())
	}
	formatted, _ := os.ReadFile(path)
	if !strings.Contains(string(formatted), "(\n\ta int,") {
		t.Errorf("file was not rewritten with multi-line signature:\n%s", string(formatted))
	}
}

func TestRun_IncompatibleFlags(t *testing.T) {
	var out, errBuf bytes.Buffer
	exit := run([]string{"gosigfmt", "-w", "-l", "x.go"}, nil, &out, &errBuf)
	if exit != 1 {
		t.Errorf("exit: want 1, got %d", exit)
	}
	if !strings.Contains(errBuf.String(), "mutually exclusive") {
		t.Errorf("missing error in stderr: %s", errBuf.String())
	}
}

func TestRun_DiffMode(t *testing.T) {
	dir := t.TempDir()
	long := strings.Repeat("a", 60)
	src := "package p\nfunc f(a int, " + long + " int, b int, c int) error { return nil }\n"
	path := filepath.Join(dir, "x.go")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	exit := run([]string{"gosigfmt", "-d", path}, nil, &out, &errBuf)
	if exit != 0 {
		t.Errorf("exit: want 0, got %d. stderr=%s", exit, errBuf.String())
	}
	if !strings.Contains(out.String(), "@@") {
		t.Errorf("expected diff hunk header @@; got: %q", out.String())
	}
}
