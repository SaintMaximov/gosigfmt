package format

import (
	"bytes"
	"go/parser"
	"go/token"
	"testing"

	"github.com/SaintMaximov/gosigfmt/internal/config"
)

func FuzzFormat(f *testing.F) {
	seeds := []string{
		"package p\nfunc f() {}\n",
		"package p\nfunc f(a int) error { return nil }\n",
		"package p\nfunc f(\n\ta int,\n\tb int,\n) error {\n\treturn nil\n}\n",
		"package p\ntype I interface { Do(a int) error }\n",
		"package p\nfunc F[T any](x T) T { return x }\n",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	cfg := config.Defaults()
	f.Fuzz(func(t *testing.T, src []byte) {
		fset := token.NewFileSet()
		if _, err := parser.ParseFile(fset, "in.go", src, parser.ParseComments); err != nil {
			t.Skip("invalid Go source")
		}
		out, err := Format(src, cfg)
		if err != nil {
			t.Fatalf("Format: %v\nsrc:\n%s", err, src)
		}
		if _, err := parser.ParseFile(token.NewFileSet(), "out.go", out, parser.ParseComments); err != nil {
			t.Fatalf("output not valid Go: %v\nout:\n%s", err, out)
		}
		again, err := Format(out, cfg)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(out, again) {
			t.Errorf("not idempotent")
		}
	})
}
