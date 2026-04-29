package runner

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/SaintMaximov/gosigfmt/internal/config"
)

// walk expands path arguments (including "./..." patterns) into a list of
// .go files to format, applying skip filters from cfg.
func walk(args []string, cfg config.Config) ([]string, error) {
	var out []string
	for _, arg := range args {
		recursive := false
		path := arg
		if strings.HasSuffix(arg, "/...") {
			recursive = true
			path = strings.TrimSuffix(arg, "/...")
		} else if arg == "..." {
			recursive = true
			path = "."
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if shouldKeep(path, cfg) {
				out = append(out, path)
			}
			continue
		}
		walkFn := func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				name := d.Name()
				if name == "vendor" || (strings.HasPrefix(name, ".") && p != path) {
					return fs.SkipDir
				}
				if !recursive && p != path {
					return fs.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".go") {
				return nil
			}
			if shouldKeep(p, cfg) {
				out = append(out, p)
			}
			return nil
		}
		if err := filepath.WalkDir(path, walkFn); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// shouldKeep returns true if path should be included after applying cfg filters.
func shouldKeep(path string, cfg config.Config) bool {
	if !cfg.FormatTestFiles && strings.HasSuffix(path, "_test.go") {
		return false
	}
	if cfg.SkipGenerated && isGenerated(path) {
		return false
	}
	return true
}

// isGenerated reports whether the file at path begins with the canonical Go
// "Code generated ... DO NOT EDIT." marker (within the first 5 lines).
func isGenerated(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for i := 0; i < 5 && scanner.Scan(); i++ {
		line := scanner.Text()
		if strings.Contains(line, "Code generated") && strings.Contains(line, "DO NOT EDIT") {
			return true
		}
	}
	return false
}
