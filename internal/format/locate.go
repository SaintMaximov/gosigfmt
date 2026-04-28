package format

import (
	"go/ast"
	"go/token"

	"github.com/SaintMaximov/gosigfmt/internal/config"
)

type sigKind int

const (
	sigFuncDecl sigKind = iota
	sigInterfaceMethod
	sigFuncLit
)

type signature struct {
	kind        sigKind
	receiver    *ast.FieldList
	typeParams  *ast.FieldList
	params      *ast.FieldList // never nil
	results     *ast.FieldList // nil if no returns
	funcKeyword token.Pos      // start of "func" or interface method name
	bodyStart   token.Pos      // "{" or end of method-in-interface line
	nameSpan    span
	commentMap  ast.CommentMap // populated by signatures() for the file
	fset        *token.FileSet
}

type span struct{ start, end int }

// signatures returns all signatures in the file that should be considered for
// formatting based on cfg.Targets.
func signatures(fset *token.FileSet, file *ast.File, cfg config.Config) []signature {
	var out []signature
	cmap := ast.NewCommentMap(fset, file, file.Comments)

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if !cfg.Targets.Functions {
				return true
			}
			s := signature{
				kind:        sigFuncDecl,
				receiver:    x.Recv,
				params:      x.Type.Params,
				results:     x.Type.Results,
				funcKeyword: x.Pos(),
				nameSpan:    span{start: fset.Position(x.Name.Pos()).Offset, end: fset.Position(x.Name.End()).Offset},
				commentMap:  cmap,
				fset:        fset,
			}
			if x.Type.TypeParams != nil {
				s.typeParams = x.Type.TypeParams
			}
			if x.Body != nil {
				s.bodyStart = x.Body.Lbrace
			} else {
				s.bodyStart = x.End()
			}
			out = append(out, s)
		case *ast.InterfaceType:
			if !cfg.Targets.Interfaces {
				return true
			}
			if x.Methods == nil {
				return true
			}
			for _, field := range x.Methods.List {
				ft, ok := field.Type.(*ast.FuncType)
				if !ok {
					continue
				}
				if len(field.Names) == 0 {
					continue // embedded interface, not a method
				}
				s := signature{
					kind:        sigInterfaceMethod,
					params:      ft.Params,
					results:     ft.Results,
					funcKeyword: field.Names[0].Pos(),
					bodyStart:   field.End(),
					nameSpan:    span{start: fset.Position(field.Names[0].Pos()).Offset, end: fset.Position(field.Names[0].End()).Offset},
					commentMap:  cmap,
					fset:        fset,
				}
				if ft.TypeParams != nil {
					s.typeParams = ft.TypeParams
				}
				out = append(out, s)
			}
		}
		return true
	})
	return out
}
