package format

import "sort"

type edit struct {
	start, end int
	text       string
}

// applyEdits applies a list of byte-offset edits to src and returns the result.
// Edits are sorted by start offset descending, so earlier offsets remain valid
// throughout. Edits MUST NOT overlap.
func applyEdits(src []byte, edits []edit) []byte {
	if len(edits) == 0 {
		return src
	}
	sorted := make([]edit, len(edits))
	copy(sorted, edits)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].start > sorted[j].start
	})
	out := make([]byte, len(src))
	copy(out, src)
	for _, e := range sorted {
		out = append(out[:e.start], append([]byte(e.text), out[e.end:]...)...)
	}
	return out
}
