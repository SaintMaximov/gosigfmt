package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/SaintMaximov/gosigfmt/internal/runner"
)

type cliFlags struct {
	write    bool
	list     bool
	diff     bool
	cfgPath  string
	noConfig bool
	parallel int
	showVer  bool
}

func parseFlags(args []string, stderr io.Writer) (*cliFlags, []string, error) {
	fs := flag.NewFlagSet("gosigfmt", flag.ContinueOnError)
	fs.SetOutput(stderr)

	f := &cliFlags{}
	fs.BoolVar(&f.write, "w", false, "write changes back to files in place")
	fs.BoolVar(&f.list, "l", false, "list files needing formatting (CI mode)")
	fs.BoolVar(&f.diff, "d", false, "show unified diff of pending changes")
	fs.StringVar(&f.cfgPath, "config", "", "explicit path to .gosigfmt.yaml")
	fs.BoolVar(&f.noConfig, "no-config", false, "ignore .gosigfmt.yaml and use defaults")
	fs.IntVar(&f.parallel, "parallel", 0, "worker count (default: GOMAXPROCS)")
	fs.BoolVar(&f.showVer, "version", false, "print version and exit")

	fs.Usage = func() {
		_, _ = fmt.Fprintf(stderr, "Usage: gosigfmt [flags] [path...]\n\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args[1:]); err != nil {
		return nil, nil, err
	}
	return f, fs.Args(), nil
}

func (f *cliFlags) resolveMode() (runner.Mode, error) {
	mode, count := runner.ModePrint, 0
	if f.write {
		mode, count = runner.ModeWrite, count+1
	}
	if f.list {
		mode, count = runner.ModeList, count+1
	}
	if f.diff {
		mode, count = runner.ModeDiff, count+1
	}
	if count > 1 {
		return mode, fmt.Errorf("-w, -l, and -d are mutually exclusive")
	}
	return mode, nil
}
