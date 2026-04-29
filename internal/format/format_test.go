package format

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SaintMaximov/gosigfmt/internal/config"
	"gopkg.in/yaml.v3"
)

var update = flag.Bool("update", false, "update golden files")

func TestFormat_NoChangeNeeded(t *testing.T) {
	src := []byte(`package p

func add(a, b int) int {
	return a + b
}
`)
	cfg := config.Defaults()
	out, err := Format(src, cfg)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if string(out) != string(src) {
		t.Errorf("expected no change\nwant: %q\n got: %q", string(src), string(out))
	}
}

func TestFormat_LongSignatureExpands(t *testing.T) {
	// 70 x's makes the signature 107 chars, exceeding the default 100-char limit.
	long := strings.Repeat("x", 70)
	src := []byte("package p\n\nfunc f(a int, " + long + " int, b string) error {\n\treturn nil\n}\n")
	cfg := config.Defaults()
	out, err := Format(src, cfg)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(string(out), "(\n\ta int,") {
		t.Errorf("expected each-param-per-line expansion; got:\n%s", string(out))
	}
}

func TestFormat_MultiLineCollapses(t *testing.T) {
	src := []byte(`package p

func f(
	a int,
	b int,
) int {
	return 0
}
`)
	cfg := config.Defaults()
	out, err := Format(src, cfg)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(string(out), "func f(a int, b int) int {") {
		t.Errorf("expected collapse to single line; got:\n%s", string(out))
	}
}

func TestFormat_Idempotent(t *testing.T) {
	long := strings.Repeat("x", 70)
	src := []byte("package p\n\nfunc f(a int, " + long + " int, b string) error {\n\treturn nil\n}\n")
	cfg := config.Defaults()
	once, err := Format(src, cfg)
	if err != nil {
		t.Fatalf("Format 1: %v", err)
	}
	twice, err := Format(once, cfg)
	if err != nil {
		t.Fatalf("Format 2: %v", err)
	}
	if string(once) != string(twice) {
		t.Errorf("not idempotent.\nfirst:\n%s\nsecond:\n%s", string(once), string(twice))
	}
}

func TestFormat_CommentLineAfterParam(t *testing.T) {
	// Force expand by making a single param very long.
	long := strings.Repeat("x", 70)
	src := []byte("package p\n\nfunc f(a int, b int /* primary */, c int, " + long + " int) error {\n\treturn nil\n}\n")
	cfg := config.Defaults()
	out, err := Format(src, cfg)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if !strings.Contains(string(out), "// primary") {
		t.Errorf("block comment must be converted to line comment in expanded form; got:\n%s", string(out))
	}
}

func TestFormat_LineCommentForbidsCollapse(t *testing.T) {
	src := []byte(`package p

func f(
	a int, // primary
	b int,
) int {
	return 0
}
`)
	cfg := config.Defaults()
	out, err := Format(src, cfg)
	if err != nil {
		t.Fatalf("Format: %v", err)
	}
	if strings.Contains(string(out), "func f(a int") {
		t.Errorf("must NOT collapse when line comments present; got:\n%s", string(out))
	}
}

func TestGolden(t *testing.T) {
	cases, err := filepath.Glob("../../testdata/golden/*")
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range cases {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			input, err := os.ReadFile(filepath.Join(dir, "input.go"))
			if err != nil {
				t.Fatal(err)
			}
			cfg := config.Defaults()
			if cfgBytes, err := os.ReadFile(filepath.Join(dir, "config.yaml")); err == nil {
				if err := yaml.Unmarshal(cfgBytes, &cfg); err != nil {
					t.Fatalf("parse config.yaml: %v", err)
				}
			}
			got, err := Format(input, cfg)
			if err != nil {
				t.Fatalf("Format: %v", err)
			}
			expectedPath := filepath.Join(dir, "expected.go")
			if *update {
				if err := os.WriteFile(expectedPath, got, 0644); err != nil {
					t.Fatal(err)
				}
				return
			}
			expected, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, expected) {
				t.Errorf("mismatch in %s\nexpected:\n%s\ngot:\n%s", name, string(expected), string(got))
			}
			again, err := Format(got, cfg)
			if err != nil {
				t.Fatalf("idempotent re-format: %v", err)
			}
			if !bytes.Equal(again, got) {
				t.Errorf("not idempotent in %s", name)
			}
		})
	}
}
