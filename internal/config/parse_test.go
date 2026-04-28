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
