package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

const ConfigFileName = ".gosigfmt.yaml"

// FindConfig walks up the directory tree from startDir, looking for a
// .gosigfmt.yaml file. If found, the file is parsed and the (Config, path) pair
// is returned. If not found, Defaults() and "" are returned with a nil error.
// Filesystem errors (other than not-exists) are returned as-is.
func FindConfig(startDir string) (Config, string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return Config{}, "", err
	}
	dir := abs
	for {
		candidate := filepath.Join(dir, ConfigFileName)
		data, err := os.ReadFile(candidate)
		if err == nil {
			cfg, perr := ParseYAML(data)
			if perr != nil {
				return Config{}, candidate, perr
			}
			return cfg, candidate, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return Config{}, "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// reached filesystem root
			return Defaults(), "", nil
		}
		dir = parent
	}
}
