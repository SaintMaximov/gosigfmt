package format

import (
	"go/ast"
	"strings"

	"github.com/SaintMaximov/gosigfmt/internal/config"
)

type action int

const (
	actionKeep action = iota
	actionCollapse
	actionExpandParamsOnly
	actionExpandParamsAndResults
)

func (a action) String() string {
	switch a {
	case actionKeep:
		return "keep"
	case actionCollapse:
		return "collapse"
	case actionExpandParamsOnly:
		return "expandParamsOnly"
	case actionExpandParamsAndResults:
		return "expandParamsAndResults"
	}
	return "unknown"
}

// decide chooses an action for the given signature.
// baseIndent is the indentation of the line containing "func".
// alreadyMultiLine indicates whether the signature in source is already split across lines.
func decide(s signature, cfg config.Config, baseIndent string, alreadyMultiLine bool) (action, error) {
	single, err := renderSingleLine(s, cfg)
	if err != nil {
		return actionKeep, err
	}
	totalLen := len(baseIndent) + len(single)

	// Forced modes for results splitting take precedence when total exceeds line length.
	switch cfg.SplitResults {
	case "always":
		if totalLen > cfg.LineLength {
			return actionExpandParamsAndResults, nil
		}
	case "never":
		if totalLen > cfg.LineLength {
			return actionExpandParamsOnly, nil
		}
	}

	if totalLen <= cfg.LineLength {
		if alreadyMultiLine && cfg.CollapseShort && !signatureHasLineComment(s) {
			return actionCollapse, nil
		}
		return actionKeep, nil
	}

	// totalLen > line_length: at minimum, expand params.
	// In "auto" mode, also expand results if even with params split the longest
	// resulting line still exceeds line_length.
	if cfg.SplitResults == "auto" {
		multi, err := renderMultiLine(s, cfg, baseIndent, false)
		if err != nil {
			return actionKeep, err
		}
		if longestLineLen(multi) > cfg.LineLength {
			return actionExpandParamsAndResults, nil
		}
		return actionExpandParamsOnly, nil
	}
	return actionExpandParamsOnly, nil
}

// signatureHasLineComment reports whether any "//" line comment is associated
// with the parameters or results of this signature.
func signatureHasLineComment(s signature) bool {
	if s.commentMap == nil {
		return false
	}
	check := func(fl *ast.FieldList) bool {
		if fl == nil {
			return false
		}
		for _, f := range fl.List {
			if cs, ok := s.commentMap[f]; ok {
				for _, group := range cs {
					for _, c := range group.List {
						if strings.HasPrefix(c.Text, "//") {
							return true
						}
					}
				}
			}
		}
		return false
	}
	return check(s.params) || check(s.results)
}

func longestLineLen(s string) int {
	m := 0
	for _, line := range strings.Split(s, "\n") {
		if len(line) > m {
			m = len(line)
		}
	}
	return m
}
