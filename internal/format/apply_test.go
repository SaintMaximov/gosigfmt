package format

import "testing"

func TestApplyEdits_Single(t *testing.T) {
	src := []byte("hello world")
	edits := []edit{{start: 6, end: 11, text: "Go"}}
	got := applyEdits(src, edits)
	if string(got) != "hello Go" {
		t.Errorf("got %q", string(got))
	}
}

func TestApplyEdits_MultipleReverseOrder(t *testing.T) {
	src := []byte("aaaa bbbb cccc")
	// Replace "aaaa" → "AA", "bbbb" → "B", "cccc" → "CCCC"
	// Offsets in original: aaaa[0:4], bbbb[5:9], cccc[10:14]
	edits := []edit{
		{start: 0, end: 4, text: "AA"},
		{start: 5, end: 9, text: "B"},
		{start: 10, end: 14, text: "CCCC"},
	}
	got := applyEdits(src, edits)
	if string(got) != "AA B CCCC" {
		t.Errorf("got %q", string(got))
	}
}
