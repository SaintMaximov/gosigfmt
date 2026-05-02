package config

import "testing"

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	if cfg.LineLength != 100 {
		t.Errorf("LineLength: want 100, got %d", cfg.LineLength)
	}
	if !cfg.CollapseShort {
		t.Error("CollapseShort: want true, got false")
	}
	if cfg.SplitResults != "auto" {
		t.Errorf("SplitResults: want auto, got %q", cfg.SplitResults)
	}
	if cfg.ExpandGroupedParams {
		t.Error("ExpandGroupedParams: want false, got true")
	}
	if !cfg.Targets.Functions {
		t.Error("Targets.Functions: want true")
	}
	if !cfg.Targets.Interfaces {
		t.Error("Targets.Interfaces: want true")
	}
	if !cfg.Targets.Generics {
		t.Error("Targets.Generics: want true")
	}
	if cfg.Targets.FuncLiterals {
		t.Error("Targets.FuncLiterals: want false")
	}
	if !cfg.FormatTestFiles {
		t.Error("FormatTestFiles: want true")
	}
	if !cfg.SkipGenerated {
		t.Error("SkipGenerated: want true")
	}
	if !cfg.WarnOnSkip {
		t.Error("WarnOnSkip: want true")
	}
}
