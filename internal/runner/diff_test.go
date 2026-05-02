package runner

import (
	"strings"
	"testing"
)

func TestDiff_NoChange(t *testing.T) {
	got := Diff("foo.go", []byte("a\nb\n"), []byte("a\nb\n"))
	if got != "" {
		t.Errorf("want empty diff, got %q", got)
	}
}

func TestDiff_OneLine(t *testing.T) {
	got := Diff("foo.go", []byte("a\nb\n"), []byte("a\nB\n"))
	if !strings.Contains(got, "-b") || !strings.Contains(got, "+B") {
		t.Errorf("missing -b / +B in diff:\n%s", got)
	}
	if !strings.Contains(got, "foo.go") {
		t.Errorf("missing filename in diff header")
	}
}
