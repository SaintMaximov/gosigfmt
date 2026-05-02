package runner

import (
	"fmt"
	"io"

	"github.com/SaintMaximov/gosigfmt/internal/format"
)

// processStdin handles the stdin input path. Mode==ModeWrite is rejected
// since there's no destination file.
func processStdin(opts Options) (int, error) {
	if opts.Mode == ModeWrite {
		return 1, fmt.Errorf("cannot use -w with stdin input")
	}
	src, err := io.ReadAll(opts.Stdin)
	if err != nil {
		return 1, err
	}
	cfg := opts.Cfg
	out, err := format.Format(src, cfg)
	if err != nil {
		return 1, err
	}
	switch opts.Mode {
	case ModePrint:
		_, _ = opts.Stdout.Write(out)
	case ModeList:
		if string(out) != string(src) {
			_, _ = fmt.Fprintln(opts.Stdout, "<stdin>")
			return 2, nil
		}
	case ModeDiff:
		_, _ = opts.Stdout.Write([]byte(Diff("<stdin>", src, out)))
	}
	return 0, nil
}
