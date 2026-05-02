package format

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"

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

// writeResults appends the results FieldList to buf, with parentheses if needed.
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
			return fmt.Errorf("render field type: %w", err)
		}
	}
	return nil
}

// renderMultiLine returns the multi-line form: each parameter on its own line.
// baseIndent is the indentation of the line containing "func".
// splitResults indicates whether the results FieldList should also be split.
//
// The output for FuncDecl/FuncLit ends with " {".
// For interface methods there is no trailing.
func renderMultiLine(
	s signature,
	cfg config.Config,
	baseIndent string,
	splitResults bool,
) (string, error) {
	paramIndent := baseIndent + "\t"
	var buf bytes.Buffer

	// Prefix
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
	case sigInterfaceMethod:
		buf.WriteString(s.name)
		if s.typeParams != nil {
			buf.WriteString("[")
			if err := printFieldListInner(&buf, s.fset, s.typeParams); err != nil {
				return "", err
			}
			buf.WriteString("]")
		}
	}

	// Params: each on its own line
	buf.WriteString("(\n")
	if err := writeFieldsMultiLine(&buf, s, s.params, paramIndent, cfg.ExpandGroupedParams); err != nil {
		return "", err
	}
	buf.WriteString(baseIndent)
	buf.WriteString(")")

	// Results
	if s.results != nil && len(s.results.List) > 0 {
		buf.WriteString(" ")
		if splitResults && hasMultipleResultFields(s.results) {
			buf.WriteString("(\n")
			if err := writeFieldsMultiLine(&buf, s, s.results, paramIndent, cfg.ExpandGroupedParams); err != nil {
				return "", err
			}
			buf.WriteString(baseIndent)
			buf.WriteString(")")
		} else {
			parens := needParens(s.results)
			if parens {
				buf.WriteString("(")
			}
			if err := printFieldListInner(&buf, s.fset, s.results); err != nil {
				return "", err
			}
			if parens {
				buf.WriteString(")")
			}
		}
	}

	if s.kind != sigInterfaceMethod {
		buf.WriteString(" {")
	}
	return buf.String(), nil
}

// writeFieldsMultiLine writes each field of fl on its own line, preceded by
// `indent` and followed by ",\n". Comments associated with each field are
// preserved: leading comments (positioned before the field) appear on their
// own line above; trailing comments appear at end of the field's line.
func writeFieldsMultiLine(
	buf *bytes.Buffer,
	s signature,
	fl *ast.FieldList,
	indent string,
	expandGrouped bool,
) error {
	if fl == nil {
		return nil
	}
	for _, f := range fl.List {
		leading, trailing := commentsForField(s, f)
		for _, c := range leading {
			buf.WriteString(indent)
			buf.WriteString(c)
			buf.WriteString("\n")
		}

		var typeBuf bytes.Buffer
		if err := printNode(&typeBuf, fset(s), f.Type); err != nil {
			return fmt.Errorf("render field type: %w", err)
		}
		typeStr := typeBuf.String()

		writeOne := func(nameStr string) {
			buf.WriteString(indent)
			if nameStr != "" {
				buf.WriteString(nameStr)
				buf.WriteString(" ")
			}
			buf.WriteString(typeStr)
			buf.WriteString(",")
			if len(trailing) > 0 {
				buf.WriteString(" ")
				buf.WriteString(strings.Join(trailing, " "))
				trailing = nil // attach to first emitted line
			}
			buf.WriteString("\n")
		}

		if len(f.Names) == 0 {
			writeOne("")
			continue
		}
		if expandGrouped {
			for _, name := range f.Names {
				writeOne(name.Name)
			}
		} else {
			names := make([]string, len(f.Names))
			for i, n := range f.Names {
				names[i] = n.Name
			}
			writeOne(strings.Join(names, ", "))
		}
	}
	return nil
}

// fset returns the file set associated with a signature (small adapter for clarity).
func fset(s signature) *token.FileSet {
	return s.fset
}

func hasMultipleResultFields(fl *ast.FieldList) bool {
	if fl == nil {
		return false
	}
	count := 0
	for _, f := range fl.List {
		if len(f.Names) == 0 {
			count++
		} else {
			count += len(f.Names)
		}
		if count > 1 {
			return true
		}
	}
	return false
}
