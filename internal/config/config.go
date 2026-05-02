package config

type Targets struct {
	Functions    bool `yaml:"functions"`
	Interfaces   bool `yaml:"interfaces"`
	Generics     bool `yaml:"generics"`
	FuncLiterals bool `yaml:"func_literals"`
}

type Config struct {
	LineLength          int     `yaml:"line_length"`
	CollapseShort       bool    `yaml:"collapse_short"`
	SplitResults        string  `yaml:"split_results"`
	ExpandGroupedParams bool    `yaml:"expand_grouped_params"`
	Targets             Targets `yaml:"targets"`
	FormatTestFiles     bool    `yaml:"format_test_files"`
	SkipGenerated       bool    `yaml:"skip_generated"`
	WarnOnSkip          bool    `yaml:"warn_on_skip"`
}

func Defaults() Config {
	return Config{
		LineLength:          100,
		CollapseShort:       true,
		SplitResults:        "auto",
		ExpandGroupedParams: false,
		Targets: Targets{
			Functions:    true,
			Interfaces:   true,
			Generics:     true,
			FuncLiterals: false,
		},
		FormatTestFiles: true,
		SkipGenerated:   true,
		WarnOnSkip:      true,
	}
}
