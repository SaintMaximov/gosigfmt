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
		}
		return true
	})
	return out
}
