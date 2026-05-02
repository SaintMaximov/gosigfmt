package config

import (
	"strings"
	"testing"
)

func TestValidate_OK(t *testing.T) {
	cfg := Defaults()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Defaults must validate: %v", err)
	}
}

func TestValidate_LineLengthZero(t *testing.T) {
	cfg := Defaults()
	cfg.LineLength = 0
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "line_length") {
		t.Errorf("want line_length error, got %v", err)
	}
}

func TestValidate_BadSplitResults(t *testing.T) {
	cfg := Defaults()
	cfg.SplitResults = "magic"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "split_results") {
		t.Errorf("want split_results error, got %v", err)
	}
}

func TestValidate_AllTargetsDisabled(t *testing.T) {
	cfg := Defaults()
	cfg.Targets = Targets{}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "target") {
		t.Errorf("want target error, got %v", err)
	}
}

func TestParseYAML_Empty(t *testing.T) {
	cfg, err := ParseYAML(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LineLength != 100 {
		t.Errorf("empty YAML must yield defaults; got LineLength=%d", cfg.LineLength)
	}
}

func TestParseYAML_Override(t *testing.T) {
	yaml := []byte("line_length: 80\ncollapse_short: false\n")
	cfg, err := ParseYAML(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LineLength != 80 {
		t.Errorf("LineLength: want 80, got %d", cfg.LineLength)
	}
	if cfg.CollapseShort {
		t.Error("CollapseShort: want false, got true")
	}
	if cfg.SplitResults != "auto" {
		t.Errorf("SplitResults must keep default 'auto', got %q", cfg.SplitResults)
	}
}

func TestParseYAML_UnknownKey(t *testing.T) {
	yaml := []byte("line_length: 80\nfoobar: 1\n")
	_, err := ParseYAML(yaml)
	if err == nil || !strings.Contains(err.Error(), "foobar") {
		t.Errorf("want unknown-key error mentioning 'foobar', got %v", err)
	}
}

func TestParseYAML_BadType(t *testing.T) {
	yaml := []byte("line_length: not-a-number\n")
	_, err := ParseYAML(yaml)
	if err == nil {
		t.Error("want error on malformed YAML")
	}
}

func TestParseYAML_ValidationFailure(t *testing.T) {
	// Valid YAML, parses cleanly, but fails validation (LineLength=0).
	yaml := []byte("line_length: 0\n")
	_, err := ParseYAML(yaml)
	if err == nil || !strings.Contains(err.Error(), "line_length") {
		t.Errorf("want validation error mentioning line_length, got %v", err)
	}
}
