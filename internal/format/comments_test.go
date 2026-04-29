package format

import "testing"

func TestBlockToLine_Single(t *testing.T) {
	in := "/* hello */"
	got := blockToLine(in)
	if got != "// hello" {
		t.Errorf("got %q", got)
	}
}

func TestBlockToLine_Trim(t *testing.T) {
	in := "/*  spaced  */"
	got := blockToLine(in)
	if got != "// spaced" {
		t.Errorf("got %q", got)
	}
}

func TestBlockToLine_Multiline(t *testing.T) {
	in := "/* line1\nline2 */"
	got := blockToLine(in)
	if got != in { // multi-line: leave unchanged
		t.Errorf("got %q", got)
	}
}
