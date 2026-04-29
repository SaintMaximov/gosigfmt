package main

import (
	"flag"
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
	fs := flag.NewFlagSet("gosigfmt", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		write    = fs.Bool("w", false, "write changes back to files in place")
		list     = fs.Bool("l", false, "list files needing formatting (CI mode)")
		diff     = fs.Bool("d", false, "show unified diff of pending changes")
		cfgPath  = fs.String("config", "", "explicit path to .gosigfmt.yaml")
		noConfig = fs.Bool("no-config", false, "ignore .gosigfmt.yaml and use defaults")
		parallel = fs.Int("parallel", 0, "worker count (default: GOMAXPROCS)")
		showVer  = fs.Bool("version", false, "print version and exit")
	)

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: gosigfmt [flags] [path...]\n\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	if *showVer {
		fmt.Fprintln(stdout, "gosigfmt version", version)
		return 0
	}

	// validate flag combinations
	modeCount := 0
	if *write {
		modeCount++
	}
	if *list {
		modeCount++
	}
	if *diff {
		modeCount++
	}
	if modeCount > 1 {
		fmt.Fprintln(stderr, "error: -w, -l, and -d are mutually exclusive")
		return 1
	}
	if *cfgPath != "" && *noConfig {
		fmt.Fprintln(stderr, "error: --config and --no-config are mutually exclusive")
		return 1
	}

	mode := runner.ModePrint
	switch {
	case *write:
		mode = runner.ModeWrite
	case *list:
		mode = runner.ModeList
	case *diff:
		mode = runner.ModeDiff
	}

	// Pre-resolve cfg for stdin path (where there's no file dir to walk up from).
	cfg := config.Defaults()
	if *cfgPath != "" && !*noConfig {
		data, err := os.ReadFile(*cfgPath)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		c, err := config.ParseYAML(data)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 1
		}
		cfg = c
	}

	opts := runner.Options{
		Mode:     mode,
		Parallel: *parallel,
		Cfg:      cfg,
		Stdin:    stdin,
		Stdout:   stdout,
		Stderr:   stderr,
		NoConfig: *noConfig,
		Config:   *cfgPath,
	}

	exit, err := runner.Process(fs.Args(), opts)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		if exit == 0 {
			exit = 1
		}
	}
	return exit
}
