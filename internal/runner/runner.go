package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"github.com/SaintMaximov/gosigfmt/internal/config"
	"github.com/SaintMaximov/gosigfmt/internal/format"
)

type Mode int

const (
	ModePrint Mode = iota
	ModeWrite
	ModeList
	ModeDiff
)

type Options struct {
	Mode     Mode
	Parallel int // 0 = GOMAXPROCS
	Cfg      config.Config
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	NoConfig bool
	Config   string // explicit path; empty = discovery
}

// Process reads paths and runs the formatter according to opts.
// Returns:
//   - exitCode: 0 success, 2 if Mode==ModeList found files needing formatting,
//     1+ on errors.
//   - error: terminal error (e.g., usage error). main should print and exit.
func Process(paths []string, opts Options) (int, error) {
	if len(paths) == 0 {
		return processStdin(opts)
	}
	files, err := walk(paths, opts.Cfg)
	if err != nil {
		return 1, err
	}
	if len(files) == 0 {
		return 0, nil
	}

	parallel := opts.Parallel
	if parallel <= 0 {
		parallel = runtime.GOMAXPROCS(0)
	}

	type job struct {
		index int
		path  string
	}
	type result struct {
		index    int
		path     string
		original []byte
		output   []byte
		needsFmt bool
		diff     string
		err      error
		warns    []string
	}

	jobs := make(chan job)
	results := make(chan result)
	var wg sync.WaitGroup

	for w := 0; w < parallel; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				r := result{index: j.index, path: j.path}
				cfg, _, err := opts.resolveConfig(j.path)
				if err != nil {
					r.err = err
					results <- r
					continue
				}
				src, err := os.ReadFile(j.path)
				if err != nil {
					r.err = err
					results <- r
					continue
				}
				r.original = src
				out, err := format.Format(src, cfg)
				if err != nil {
					if opts.Cfg.WarnOnSkip {
						r.warns = append(r.warns, fmt.Sprintf("skipping %s: %v", j.path, err))
					}
					r.output = src
					results <- r
					continue
				}
				r.output = out
				r.needsFmt = !bytes.Equal(src, out)
				if opts.Mode == ModeDiff && r.needsFmt {
					r.diff = Diff(j.path, src, out)
				}
				results <- r
			}
		}()
	}

	go func() {
		for i, p := range files {
			jobs <- job{index: i, path: p}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	collected := make([]result, 0, len(files))
	for r := range results {
		collected = append(collected, r)
	}
	sort.Slice(collected, func(i, j int) bool { return collected[i].index < collected[j].index })

	exit := 0
	for _, r := range collected {
		for _, w := range r.warns {
			_, _ = fmt.Fprintln(opts.Stderr, w)
		}
		if r.err != nil {
			_, _ = fmt.Fprintf(opts.Stderr, "error %s: %v\n", r.path, r.err)
			exit = 1
			continue
		}
		switch opts.Mode {
		case ModePrint:
			_, _ = opts.Stdout.Write(r.output)
		case ModeWrite:
			if r.needsFmt {
				if err := atomicWrite(r.path, r.output); err != nil {
					_, _ = fmt.Fprintf(opts.Stderr, "write %s: %v\n", r.path, err)
					exit = 1
				}
			}
		case ModeList:
			if r.needsFmt {
				_, _ = fmt.Fprintln(opts.Stdout, r.path)
				if exit == 0 {
					exit = 2
				}
			}
		case ModeDiff:
			if r.needsFmt {
				_, _ = opts.Stdout.Write([]byte(r.diff))
			}
		}
	}
	return exit, nil
}

// resolveConfig returns the config for a given file path, applying NoConfig and
// explicit Config overrides if set. Otherwise walks up from filePath's dir.
func (o Options) resolveConfig(filePath string) (config.Config, string, error) {
	if o.NoConfig {
		return config.Defaults(), "", nil
	}
	if o.Config != "" {
		data, err := os.ReadFile(o.Config)
		if err != nil {
			return config.Config{}, "", err
		}
		cfg, err := config.ParseYAML(data)
		return cfg, o.Config, err
	}
	return config.FindConfig(filepath.Dir(filePath))
}

// atomicWrite writes data to path via temp file + rename, preserving file mode.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".gosigfmt-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if info, err := os.Stat(path); err == nil {
		_ = os.Chmod(tmpPath, info.Mode())
	}
	return os.Rename(tmpPath, path)
}
