package runner

import (
	"bytes"
	"fmt"
	"strings"
)

// Diff returns a unified diff between original and modified for the given path.
// Returns "" when both inputs are equal.
// Uses a simple line-level LCS algorithm — sufficient for source files.
func Diff(path string, original, modified []byte) string {
	if bytes.Equal(original, modified) {
		return ""
	}
	a := splitLines(string(original))
	b := splitLines(string(modified))
	hunks := computeHunks(a, b)

	var out strings.Builder
	fmt.Fprintf(&out, "--- %s\n+++ %s\n", path, path)
	for _, h := range hunks {
		fmt.Fprintf(&out, "@@ -%d,%d +%d,%d @@\n", h.aStart+1, h.aLen, h.bStart+1, h.bLen)
		for _, line := range h.lines {
			out.WriteString(line)
			out.WriteString("\n")
		}
	}
	return out.String()
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	out := strings.Split(s, "\n")
	if len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}

type hunk struct {
	aStart, aLen int
	bStart, bLen int
	lines        []string
}

// computeHunks computes hunks via LCS. Emits one hunk per contiguous change
// region with up to 3 lines of context around it.
func computeHunks(a, b []string) []hunk {
	lcs := lcsTable(a, b)
	ops := backtrack(lcs, a, b)

	const ctx = 3
	var hunks []hunk
	i := 0
	for i < len(ops) {
		if ops[i].kind == opEqual {
			i++
			continue
		}
		// find boundaries with context
		start := i
		for start > 0 && i-start < ctx {
			start--
		}
		end := i
		for end < len(ops) && (ops[end].kind != opEqual || end-i < ctx) {
			end++
		}
		// extend end with up to ctx equal ops
		k := 0
		for end < len(ops) && ops[end].kind == opEqual && k < ctx {
			end++
			k++
		}
		// build hunk
		var h hunk
		var aStart, bStart = -1, -1
		for j := start; j < end; j++ {
			op := ops[j]
			if aStart < 0 && (op.kind == opEqual || op.kind == opDel) {
				aStart = op.aIdx
			}
			if bStart < 0 && (op.kind == opEqual || op.kind == opIns) {
				bStart = op.bIdx
			}
			switch op.kind {
			case opEqual:
				h.lines = append(h.lines, " "+a[op.aIdx])
				h.aLen++
				h.bLen++
			case opDel:
				h.lines = append(h.lines, "-"+a[op.aIdx])
				h.aLen++
			case opIns:
				h.lines = append(h.lines, "+"+b[op.bIdx])
				h.bLen++
			}
		}
		if aStart < 0 {
			aStart = 0
		}
		if bStart < 0 {
			bStart = 0
		}
		h.aStart = aStart
		h.bStart = bStart
		hunks = append(hunks, h)
		i = end
	}
	return hunks
}

func lcsTable(a, b []string) [][]int {
	n, m := len(a), len(b)
	t := make([][]int, n+1)
	for i := range t {
		t[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if a[i-1] == b[j-1] {
				t[i][j] = t[i-1][j-1] + 1
			} else if t[i-1][j] >= t[i][j-1] {
				t[i][j] = t[i-1][j]
			} else {
				t[i][j] = t[i][j-1]
			}
		}
	}
	return t
}

type opKind int

const (
	opEqual opKind = iota
	opDel
	opIns
)

type diffOp struct {
	kind       opKind
	aIdx, bIdx int
}

func backtrack(t [][]int, a, b []string) []diffOp {
	var ops []diffOp
	i, j := len(a), len(b)
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && a[i-1] == b[j-1]:
			ops = append(ops, diffOp{kind: opEqual, aIdx: i - 1, bIdx: j - 1})
			i--
			j--
		case j > 0 && (i == 0 || t[i][j-1] >= t[i-1][j]):
			ops = append(ops, diffOp{kind: opIns, aIdx: i, bIdx: j - 1})
			j--
		default:
			ops = append(ops, diffOp{kind: opDel, aIdx: i - 1, bIdx: j})
			i--
		}
	}
	for l, r := 0, len(ops)-1; l < r; l, r = l+1, r-1 {
		ops[l], ops[r] = ops[r], ops[l]
	}
	return ops
}
