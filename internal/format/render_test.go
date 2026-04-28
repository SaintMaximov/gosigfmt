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
