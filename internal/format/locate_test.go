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

func TestSignatures_InterfaceMethod(t *testing.T) {
	src := `package p

type Doer interface {
	Do(ctx context.Context, n int) error
}
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	if len(sigs) != 1 {
		t.Fatalf("want 1 signature (interface method), got %d", len(sigs))
	}
	if sigs[0].kind != sigInterfaceMethod {
		t.Errorf("want sigInterfaceMethod, got %v", sigs[0].kind)
	}
}

func TestSignatures_InterfaceMethodDisabled(t *testing.T) {
	src := `package p

type Doer interface {
	Do(ctx context.Context, n int) error
}
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	cfg.Targets.Interfaces = false
	sigs := signatures(fset, file, cfg)
	if len(sigs) != 0 {
		t.Errorf("want 0 signatures when Interfaces disabled, got %d", len(sigs))
	}
}

func TestSignatures_InterfaceEmbedded(t *testing.T) {
	src := `package p

type Reader interface {
	Read(p []byte) (int, error)
}

type ReadWriter interface {
	Reader
	Write(p []byte) (int, error)
}
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	// Reader has 1 method (Read), ReadWriter has Reader (embedded — skipped) + Write (1 method).
	// Total expected: 2 signatures (Read and Write); the embedded Reader is NOT a signature.
	if len(sigs) != 2 {
		t.Fatalf("want 2 signatures (Read + Write), got %d", len(sigs))
	}
	for _, s := range sigs {
		if s.kind != sigInterfaceMethod {
			t.Errorf("expected sigInterfaceMethod, got %v", s.kind)
		}
	}
}

func TestSignatures_FuncLit_DisabledByDefault(t *testing.T) {
	src := `package p

var f = func(a int, b int) int { return a + b }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults() // FuncLiterals=false
	sigs := signatures(fset, file, cfg)
	if len(sigs) != 0 {
		t.Errorf("want 0 sigs when FuncLiterals=false, got %d", len(sigs))
	}
}

func TestSignatures_FuncLit_Enabled(t *testing.T) {
	src := `package p

var f = func(a int, b int) int { return a + b }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	cfg.Targets.FuncLiterals = true
	sigs := signatures(fset, file, cfg)
	if len(sigs) != 1 {
		t.Fatalf("want 1 sig when FuncLiterals=true, got %d", len(sigs))
	}
	if sigs[0].kind != sigFuncLit {
		t.Errorf("want sigFuncLit, got %v", sigs[0].kind)
	}
}
