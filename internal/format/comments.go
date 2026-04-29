package format

import "strings"

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
