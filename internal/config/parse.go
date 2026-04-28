package config

import "fmt"

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
