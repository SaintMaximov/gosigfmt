package format

import (
	"strings"
	"testing"

	"github.com/SaintMaximov/gosigfmt/internal/config"
)

func TestDecide_Keep(t *testing.T) {
	src := `package p
func add(a, b int) int { return a + b }
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	sigs := signatures(fset, file, cfg)
	a, err := decide(sigs[0], cfg, "" /* baseIndent */, false /* alreadyMultiLine */)
	if err != nil {
		t.Fatal(err)
	}
	if a != actionKeep {
		t.Errorf("want actionKeep, got %v", a)
	}
}

func TestDecide_Collapse(t *testing.T) {
	// short signature that is currently multi-line → should collapse
	src := `package p
func add(
	a int,
	b int,
) int {
	return a + b
}
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults() // CollapseShort=true
	sigs := signatures(fset, file, cfg)
	a, err := decide(sigs[0], cfg, "", true /* alreadyMultiLine */)
	if err != nil {
		t.Fatal(err)
	}
	if a != actionCollapse {
		t.Errorf("want actionCollapse, got %v", a)
	}
}

func TestDecide_NoCollapseWhenDisabled(t *testing.T) {
	src := `package p
func add(
	a int,
	b int,
) int {
	return a + b
}
`
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	cfg.CollapseShort = false
	sigs := signatures(fset, file, cfg)
	a, err := decide(sigs[0], cfg, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if a != actionKeep {
		t.Errorf("want actionKeep, got %v", a)
	}
}

func TestDecide_ExpandParams(t *testing.T) {
	long := strings.Repeat("a", 30)
	src := "package p\nfunc f(" + long + " int, " + long + "2 int, " + long + "3 int) error { return nil }\n"
	fset, file := parseSrc(t, src)
	cfg := config.Defaults() // line_length=100
	sigs := signatures(fset, file, cfg)
	a, err := decide(sigs[0], cfg, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if a != actionExpandParamsOnly && a != actionExpandParamsAndResults {
		t.Errorf("want expand action, got %v", a)
	}
}

func TestDecide_SplitResultsAlways(t *testing.T) {
	long := strings.Repeat("a", 30)
	src := "package p\nfunc f(" + long + " int) (x int, y int) { return 0, 0 }\n"
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	cfg.SplitResults = "always"
	cfg.LineLength = 30 // force splitting
	sigs := signatures(fset, file, cfg)
	a, err := decide(sigs[0], cfg, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if a != actionExpandParamsAndResults {
		t.Errorf("want actionExpandParamsAndResults, got %v", a)
	}
}

func TestDecide_SplitResultsNever(t *testing.T) {
	long := strings.Repeat("a", 30)
	src := "package p\nfunc f(" + long + " int) (x int, y int, z int, w int) { return 0,0,0,0 }\n"
	fset, file := parseSrc(t, src)
	cfg := config.Defaults()
	cfg.SplitResults = "never"
	cfg.LineLength = 30
	sigs := signatures(fset, file, cfg)
	a, err := decide(sigs[0], cfg, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if a != actionExpandParamsOnly {
		t.Errorf("want actionExpandParamsOnly, got %v", a)
	}
}
