package format

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"

	"github.com/SaintMaximov/gosigfmt/internal/config"
)

// renderSingleLine returns the canonical single-line form of a signature.
// For FuncDecl/FuncLit ends with " {"; for interface methods has no trailing.
func renderSingleLine(s signature, cfg config.Config) (string, error) {
	var buf bytes.Buffer
	switch s.kind {
	case sigFuncDecl, sigFuncLit:
		buf.WriteString("func")
		if s.receiver != nil {
			buf.WriteString(" (")
			if err := printFieldListInner(&buf, s.fset, s.receiver); err != nil {
				return "", err
			}
			buf.WriteString(")")
		}
		if s.name != "" {
			buf.WriteString(" ")
			buf.WriteString(s.name)
		}
		if s.typeParams != nil {
			buf.WriteString("[")
			if err := printFieldListInner(&buf, s.fset, s.typeParams); err != nil {
				return "", err
			}
			buf.WriteString("]")
		}
		buf.WriteString("(")
		if err := printFieldListInner(&buf, s.fset, s.params); err != nil {
			return "", err
		}
		buf.WriteString(")")
		if err := writeResults(&buf, s); err != nil {
			return "", err
		}
		buf.WriteString(" {")
	case sigInterfaceMethod:
		buf.WriteString(s.name)
		if s.typeParams != nil {
			buf.WriteString("[")
			if err := printFieldListInner(&buf, s.fset, s.typeParams); err != nil {
				return "", err
			}
			buf.WriteString("]")
		}
		buf.WriteString("(")
		if err := printFieldListInner(&buf, s.fset, s.params); err != nil {
			return "", err
		}
		buf.WriteString(")")
		if err := writeResults(&buf, s); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unknown sigKind: %v", s.kind)
	}
	_ = cfg
	return buf.String(), nil
}

func writeResults(buf *bytes.Buffer, s signature) error {
	if s.results == nil || len(s.results.List) == 0 {
		return nil
	}
	buf.WriteString(" ")
	parens := needParens(s.results)
	if parens {
		buf.WriteString("(")
	}
	if err := printFieldListInner(buf, s.fset, s.results); err != nil {
		return err
	}
	if parens {
		buf.WriteString(")")
	}
	return nil
}

// needParens returns true if a Results FieldList must be parenthesized
// (i.e., either has named fields or has more than one field).
func needParens(fl *ast.FieldList) bool {
	if fl == nil {
		return false
	}
	totalFields := 0
	for _, f := range fl.List {
		if len(f.Names) == 0 {
			totalFields++
		} else {
			totalFields += len(f.Names)
			return true // any names → parens
		}
	}
	return totalFields > 1
}

// printNode writes a single AST node via go/printer in single-line mode.
func printNode(buf *bytes.Buffer, fset *token.FileSet, n ast.Node) error {
	cfg := printer.Config{Mode: 0, Tabwidth: 8}
	return cfg.Fprint(buf, fset, n)
}

// printFieldListInner prints the comma-separated content of a FieldList
// without surrounding parentheses or brackets.
func printFieldListInner(buf *bytes.Buffer, fset *token.FileSet, fl *ast.FieldList) error {
	if fl == nil {
		return nil
	}
	for i, f := range fl.List {
		if i > 0 {
			buf.WriteString(", ")
		}
		// names
		for j, name := range f.Names {
			if j > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(name.Name)
		}
		if len(f.Names) > 0 {
			buf.WriteString(" ")
		}
		// type
		if err := printNode(buf, fset, f.Type); err != nil {
			return err
		}
	}
	return nil
}
