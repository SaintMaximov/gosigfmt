package runner

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func Diff(path string, original, modified []byte) string {
	if bytes.Equal(original, modified) {
		return ""
	}
	lines := diffLines(string(original), string(modified))
	hunks := buildHunks(lines)
	return formatDiff(path, hunks)
}

// line is one line of a diff with its position in both files.
// kind is ' ' (equal), '-' (delete), or '+' (insert) — also used as the unified diff prefix.
type line struct {
	kind       byte
	text       string
	aIdx, bIdx int
}

// diffLines runs go-diff on orig and mod, returning per-line diff entries
// with their kinds and 0-based positions in each file.
func diffLines(orig, mod string) []line {
	dmp := diffmatchpatch.New()
	c1, c2, la := dmp.DiffLinesToChars(orig, mod)
	diffs := dmp.DiffMain(c1, c2, false)
	diffs = dmp.DiffCharsToLines(diffs, la)

	var lines []line
	ai, bi := 0, 0
	for _, d := range diffs {
		for _, s := range splitLines(d.Text) {
			ln := line{text: s, aIdx: ai, bIdx: bi}
			switch d.Type {
			case diffmatchpatch.DiffEqual:
				ln.kind = ' '
				ai++
				bi++
			case diffmatchpatch.DiffDelete:
				ln.kind = '-'
				ai++
			case diffmatchpatch.DiffInsert:
				ln.kind = '+'
				bi++
			}
			lines = append(lines, ln)
		}
	}
	return lines
}

type hunk struct {
	aStart, aLen int
	bStart, bLen int
	lines        []string // already prefixed with ' ', '-', or '+'
}

// buildHunks finds changed regions in lines, merges those close enough that their
// context windows overlap, and returns the resulting hunks.
func buildHunks(lines []line) []hunk {
	const ctx = 3

	// Step 1: find all change regions — contiguous runs of non-equal lines.
	type region struct{ lo, hi int } // [lo, hi] inclusive
	var regions []region
	for i := 0; i < len(lines); {
		if lines[i].kind == ' ' {
			i++
			continue
		}
		hi := i
		for hi+1 < len(lines) && lines[hi+1].kind != ' ' {
			hi++
		}
		regions = append(regions, region{i, hi})
		i = hi + 1
	}
	if len(regions) == 0 {
		return nil
	}

	// Step 2: merge regions whose context windows overlap.
	// Two windows touch when the equal-line gap between regions is ≤ 2*ctx.
	var merged []region
	merged = append(merged, regions[0])
	for _, r := range regions[1:] {
		prev := &merged[len(merged)-1]
		if r.lo-prev.hi <= 2*ctx {
			prev.hi = r.hi
		} else {
			merged = append(merged, r)
		}
	}

	// Step 3: build a hunk from each merged region, adding context lines.
	var hunks []hunk
	for _, r := range merged {
		lo := max(0, r.lo-ctx)
		hi := min(len(lines), r.hi+1+ctx)
		hunks = append(hunks, hunkFromLines(lines[lo:hi]))
	}
	return hunks
}

// hunkFromLines builds a single hunk from a contiguous slice of diff lines.
func hunkFromLines(lines []line) hunk {
	h := hunk{aStart: -1, bStart: -1}
	for _, ln := range lines {
		// Capture start positions from the first line that exists in each file.
		if h.aStart < 0 && ln.kind != '+' {
			h.aStart = ln.aIdx
		}
		if h.bStart < 0 && ln.kind != '-' {
			h.bStart = ln.bIdx
		}
		// Store the line with its diff prefix, count it for each side.
		h.lines = append(h.lines, string(ln.kind)+ln.text)
		if ln.kind != '+' {
			h.aLen++
		}
		if ln.kind != '-' {
			h.bLen++
		}
	}
	// For pure insertions or deletions, the missing side starts at line 0.
	if h.aStart < 0 {
		h.aStart = 0
	}
	if h.bStart < 0 {
		h.bStart = 0
	}
	return h
}

func formatDiff(path string, hunks []hunk) string {
	var out strings.Builder
	fmt.Fprintf(&out, "--- %s\n+++ %s\n", path, path)
	for _, h := range hunks {
		fmt.Fprintf(&out, "@@ -%d,%d +%d,%d @@\n", h.aStart+1, h.aLen, h.bStart+1, h.bLen)
		for _, l := range h.lines {
			fmt.Fprintf(&out, "%s\n", l)
		}
	}
	return out.String()
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	out := strings.Split(s, "\n")
	if out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}
