package config

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func (c Config) Validate() error {
	if c.LineLength < 1 {
		return fmt.Errorf("config error: line_length must be >= 1, got %d", c.LineLength)
	}
	switch c.SplitResults {
	case "auto", "always", "never":
	default:
		return fmt.Errorf("config error: invalid value for split_results: %q (want auto|always|never)", c.SplitResults)
	}
	if !c.Targets.Functions && !c.Targets.Interfaces && !c.Targets.Generics && !c.Targets.FuncLiterals {
		return fmt.Errorf("config error: at least one target must be enabled")
	}
	return nil
}

// ParseYAML parses a .gosigfmt.yaml file. nil or empty input returns Defaults().
// Unknown keys produce an error to protect against typos.
// All values are validated before returning.
func ParseYAML(data []byte) (Config, error) {
	cfg := Defaults()
	if len(bytes.TrimSpace(data)) == 0 {
		return cfg, nil
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("config parse error: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
