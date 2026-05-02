package format

import (
	"go/ast"
	"strings"
)

// blockToLine converts a /* ... */ comment to a // ... comment iff the comment
// is single-line. Multi-line block comments are returned unchanged.
func blockToLine(text string) string {
	if !strings.HasPrefix(text, "/*") || !strings.HasSuffix(text, "*/") {
		return text
	}
	if strings.Contains(text, "\n") {
		return text
	}
	body := strings.TrimSpace(text[2 : len(text)-2])
	return "// " + body
}

// commentsForField returns (leading, trailing) comment text associated with
// a field. Block comments are converted to line comments via blockToLine.
//   - leading: comments preceding the field (positioned before f.Pos())
//   - trailing: comments after the field (positioned at or after f.Pos())
func commentsForField(s signature, f *ast.Field) (leading []string, trailing []string) {
	if s.commentMap == nil {
		return nil, nil
	}
	groups, ok := s.commentMap[f]
	if !ok {
		return nil, nil
	}
	for _, g := range groups {
		for _, c := range g.List {
			text := blockToLine(c.Text)
			if c.Pos() < f.Pos() {
				leading = append(leading, text)
			} else {
				trailing = append(trailing, text)
			}
		}
	}
	return leading, trailing
}
