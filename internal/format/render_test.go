package format

import (
	"strings"
	"testing"

	"github.com/SaintMaximov/gosigfmt/internal/config"
)

func TestRenderSingleLine_Simple(t *testing.T) {
	src := `package p
func add(a int, b int) int { return 0 }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	if len(sigs) != 1 {
		t.Fatalf("want 1 sig, got %d", len(sigs))
	}
	got, err := renderSingleLine(sigs[0], cfg)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := "func add(a int, b int) int {"
	if got != want {
		t.Errorf("want %q,\n got %q", want, got)
	}
}

func TestRenderSingleLine_Method(t *testing.T) {
	src := `package p
type T struct{}
func (t *T) Do(a int) error { return nil }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	got, err := renderSingleLine(sigs[0], cfg)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(got, "func (t *T) Do(a int) error") {
		t.Errorf("missing receiver/method in: %q", got)
	}
}

func TestRenderSingleLine_Generics(t *testing.T) {
	src := `package p
func Map[T any, U any](xs []T, f func(T) U) []U { return nil }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	got, err := renderSingleLine(sigs[0], cfg)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(got, "[T any, U any]") {
		t.Errorf("missing type params: %q", got)
	}
}

func TestRenderSingleLine_InterfaceMethod(t *testing.T) {
	src := `package p
type Doer interface {
	Do(ctx int, n int) error
}
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	got, err := renderSingleLine(sigs[0], cfg)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := "Do(ctx int, n int) error"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestRenderMultiLine_Simple(t *testing.T) {
	src := `package p
func longFunc(a int, b int, c int) error { return nil }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	got, err := renderMultiLine(sigs[0], cfg, "" /* baseIndent */, false /* splitResults */)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := "func longFunc(\n\ta int,\n\tb int,\n\tc int,\n) error {"
	if got != want {
		t.Errorf("want %q,\n got %q", want, got)
	}
}

func TestRenderMultiLine_GroupedParamsKept(t *testing.T) {
	src := `package p
func f(a, b int, c string) error { return nil }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults() // ExpandGroupedParams=false
	sigs := signatures(fset, file, cfg)
	got, err := renderMultiLine(sigs[0], cfg, "", false)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := "func f(\n\ta, b int,\n\tc string,\n) error {"
	if got != want {
		t.Errorf("want %q,\n got %q", want, got)
	}
}

func TestRenderMultiLine_GroupedParamsExpanded(t *testing.T) {
	src := `package p
func f(a, b int, c string) error { return nil }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	cfg.ExpandGroupedParams = true
	sigs := signatures(fset, file, cfg)
	got, err := renderMultiLine(sigs[0], cfg, "", false)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := "func f(\n\ta int,\n\tb int,\n\tc string,\n) error {"
	if got != want {
		t.Errorf("want %q,\n got %q", want, got)
	}
}

func TestRenderMultiLine_SplitResults(t *testing.T) {
	src := `package p
func f(a int) (x int, y string, err error) { return }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	got, err := renderMultiLine(sigs[0], cfg, "", true /* split results */)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := "func f(\n\ta int,\n) (\n\tx int,\n\ty string,\n\terr error,\n) {"
	if got != want {
		t.Errorf("want %q,\n got %q", want, got)
	}
}
