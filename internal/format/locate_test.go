package format

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/SaintMaximov/gosigfmt/internal/config"
)

func parseSrc(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return fset, file
}

func TestSignatures_FuncDecl(t *testing.T) {
	src := `package p

func add(a int, b int) int { return a + b }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	if len(sigs) != 1 {
		t.Fatalf("want 1 signature, got %d", len(sigs))
	}
	if sigs[0].kind != sigFuncDecl {
		t.Errorf("want sigFuncDecl, got %v", sigs[0].kind)
	}
}

func TestSignatures_TargetFunctionsDisabled(t *testing.T) {
	src := `package p

func add(a int, b int) int { return a + b }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	cfg.Targets.Functions = false
	cfg.Targets.Interfaces = true // keep config valid
	sigs := signatures(fset, file, cfg)
	if len(sigs) != 0 {
		t.Errorf("want 0 signatures when Functions disabled, got %d", len(sigs))
	}
}

func TestSignatures_Method(t *testing.T) {
	src := `package p

type T struct{}
func (t *T) Do(a int) error { return nil }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	if len(sigs) != 1 {
		t.Fatalf("want 1 signature (method), got %d", len(sigs))
	}
	if sigs[0].receiver == nil {
		t.Error("method signature must have receiver")
	}
}
