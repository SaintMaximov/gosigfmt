package format

import (
	"strings"
	"testing"

	"github.com/SaintMaximov/gosigfmt/internal/config"
)

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
