package format

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"

	"github.com/SaintMaximov/gosigfmt/internal/config"
)

// Format applies signature formatting to src per cfg. Returns the modified
// source. If src cannot be parsed, the parse error is returned and src is
// returned unchanged. After formatting, the result is reparsed; if reparse
// fails, an internal error is returned with the original src.
func Format(src []byte, cfg config.Config) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "input.go", src, parser.ParseComments)
	if err != nil {
		return src, err
	}

	sigs := signatures(fset, file, cfg)
	var edits []edit

	for _, s := range sigs {
		baseIndent := lineIndent(src, s.fullSpan.start)
		alreadyMultiLine := signatureSpansMultipleLines(src, s.fullSpan)

		a, err := decide(s, cfg, baseIndent, alreadyMultiLine)
		if err != nil {
			return src, err
		}
		if a == actionKeep {
			continue
		}
		newText, err := renderForAction(s, cfg, baseIndent, a)
		if err != nil {
			return src, err
		}
		oldText := string(src[s.fullSpan.start:s.fullSpan.end])
		if newText == oldText {
			continue
		}
		edits = append(edits, edit{start: s.fullSpan.start, end: s.fullSpan.end, text: newText})
	}

	out := applyEdits(src, edits)

	if _, err := parser.ParseFile(token.NewFileSet(), "result.go", out, parser.ParseComments); err != nil {
		return src, fmt.Errorf("internal error: formatted output is not valid Go: %w", err)
	}
	return out, nil
}

func renderForAction(s signature, cfg config.Config, baseIndent string, a action) (string, error) {
	switch a {
	case actionCollapse:
		return renderSingleLine(s, cfg)
	case actionExpandParamsOnly:
		return renderMultiLine(s, cfg, baseIndent, false)
	case actionExpandParamsAndResults:
		return renderMultiLine(s, cfg, baseIndent, true)
	}
	return "", fmt.Errorf("unexpected action: %v", a)
}

// lineIndent returns the leading whitespace of the line containing offset.
func lineIndent(src []byte, offset int) string {
	if offset > len(src) {
		offset = len(src)
	}
	start := offset
	for start > 0 && src[start-1] != '\n' {
		start--
	}
	end := start
	for end < len(src) && (src[end] == ' ' || src[end] == '\t') {
		end++
	}
	return string(src[start:end])
}

func signatureSpansMultipleLines(src []byte, sp span) bool {
	if sp.end > len(src) {
		sp.end = len(src)
	}
	return bytes.IndexByte(src[sp.start:sp.end], '\n') >= 0
}
