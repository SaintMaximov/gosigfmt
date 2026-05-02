package main

import (
	"fmt"
	"io"
	"os"

	"github.com/SaintMaximov/gosigfmt/internal/config"
	"github.com/SaintMaximov/gosigfmt/internal/runner"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args, os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	errf := func(format string, a ...any) {
		_, _ = fmt.Fprintf(stderr, "error: "+format+"\n", a...)
	}

	f, paths, err := parseFlags(args, stderr)
	if err != nil {
		return 1
	}

	if f.showVer {
		_, _ = fmt.Fprintln(stdout, "gosigfmt version", version)
		return 0
	}

	mode, err := f.resolveMode()
	if err != nil {
		errf("%s", err)
		return 1
	}
	if f.cfgPath != "" && f.noConfig {
		errf("--config and --no-config are mutually exclusive")
		return 1
	}

	cfg, err := loadConfig(f.cfgPath, f.noConfig)
	if err != nil {
		errf("%s", err)
		return 1
	}

	exit, err := runner.Process(paths, runner.Options{
		Mode:     mode,
		Parallel: f.parallel,
		Cfg:      cfg,
		Stdin:    stdin,
		Stdout:   stdout,
		Stderr:   stderr,
		NoConfig: f.noConfig,
		Config:   f.cfgPath,
	})
	if err != nil {
		errf("%s", err)
		if exit == 0 {
			exit = 1
		}
	}
	return exit
}

func loadConfig(cfgPath string, noConfig bool) (config.Config, error) {
	if cfgPath == "" || noConfig {
		return config.Defaults(), nil
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return config.Config{}, err
	}
	return config.ParseYAML(data)
}
