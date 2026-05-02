# gosigfmt

[![CI](https://github.com/SaintMaximov/gosigfmt/actions/workflows/ci.yml/badge.svg)](https://github.com/SaintMaximov/gosigfmt/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/SaintMaximov/gosigfmt.svg)](https://pkg.go.dev/github.com/SaintMaximov/gosigfmt)
[![Go Report Card](https://goreportcard.com/badge/github.com/SaintMaximov/gosigfmt)](https://goreportcard.com/report/github.com/SaintMaximov/gosigfmt)
[![Release](https://img.shields.io/github/v/release/SaintMaximov/gosigfmt)](https://github.com/SaintMaximov/gosigfmt/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Format Go function and method signatures: signatures longer than a configurable
line length are split with each parameter on its own line; multi-line signatures
that fit on one line are collapsed back. Configuration via `.gosigfmt.yaml`,
similar to `golangci-lint`.

## Why

`gofmt` does not split overlong signatures across multiple lines. Long parameter
lists either stay on one wide line or get manually wrapped in inconsistent ways.
`gosigfmt` enforces a single canonical multi-line layout for signatures over
the limit, and normalizes hybrid forms back to either single-line or
each-param-per-line — nothing in between.

## Install

### Go install

```bash
go install github.com/SaintMaximov/gosigfmt/cmd/gosigfmt@latest
```

### Pre-built binaries

Download from [Releases](https://github.com/SaintMaximov/gosigfmt/releases).
Linux, macOS, and Windows on amd64/arm64.

### Docker

```bash
docker run --rm -v "$PWD:/work" -w /work \
  ghcr.io/saintmaximov/gosigfmt:latest -l ./...
```

## Quick start

```bash
gosigfmt file.go         # print formatted source to stdout
gosigfmt -w ./...        # rewrite files in place, recursively
gosigfmt -l ./...        # CI mode: list files needing formatting (exit 2 if any)
gosigfmt -d ./...        # show unified diff
cat file.go | gosigfmt   # stdin
```

## Example

Before:

```go
func update(ctx context.Context, id int64, name string, age int, address string, country string) error {
    // ...
}
```

After (`line_length: 100`):

```go
func update(
    ctx context.Context,
    id int64,
    name string,
    age int,
    address string,
    country string,
) error {
    // ...
}
```

## Configuration

Place a `.gosigfmt.yaml` at any directory above the files you format. The tool
walks up the directory tree from each file to find the nearest config (similar
to `.editorconfig` and `.golangci.yml`). All keys are optional; missing keys
take their default values.

```yaml
line_length: 100              # max line length
collapse_short: true          # collapse multi-line signatures back when they fit

# split_results: how to handle return values when a signature must be split
#   auto    — split params first; split results only if line still doesn't fit
#   always  — always split results when params are split
#   never   — only split params, never touch results
split_results: auto

expand_grouped_params: false  # `a, b, c int` stays grouped on split (false)
                              # or expands to one line per name (true)

targets:
  functions: true             # top-level funcs and methods
  interfaces: true            # interface method declarations
  generics: true              # type parameters: func F[T any](...)
  func_literals: false        # closures inside expressions

format_test_files: true       # include *_test.go
skip_generated: true          # skip files marked "Code generated ... DO NOT EDIT."
warn_on_skip: true            # print warnings to stderr on skip
```

## CLI reference

| Flag | Behavior |
|------|----------|
| `-w` | Write changes back to files in place |
| `-l` | List files needing formatting (CI mode); exits 2 when any are listed |
| `-d` | Print unified diff between current and desired format |
| `--config <path>` | Use this config file; disables discovery |
| `--no-config` | Ignore `.gosigfmt.yaml` files; use built-in defaults |
| `--parallel N` | Worker count (default: GOMAXPROCS) |
| `--version` | Print version |
| `-h`, `--help` | Help |

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success — nothing to format, or formatted successfully |
| 1 | Usage error or unrecoverable read/write/config error |
| 2 | `-l` mode: at least one file needs formatting |
| 3 | Internal formatter error (please file a bug) |

## CI integration

```bash
if [ -n "$(gosigfmt -l ./...)" ]; then
  echo "Files need gosigfmt:"
  gosigfmt -l ./...
  exit 1
fi
```

GitHub Actions step:

```yaml
- name: gosigfmt
  run: |
    go install github.com/SaintMaximov/gosigfmt/cmd/gosigfmt@latest
    gosigfmt -l ./...
```

## Editor integration

`gosigfmt` is a standard `gofmt`-style CLI; configure your editor's "format on
save" with the binary path and `-w` flag. There is no first-party plugin.

- **Vim**: `autocmd BufWritePost *.go !gosigfmt -w %`
- **VS Code**: install [Run on Save](https://marketplace.visualstudio.com/items?itemName=emeraldwalk.RunOnSave) and configure to invoke `gosigfmt -w ${file}`.
- **GoLand**: File Watcher with program `gosigfmt`, arguments `-w $FilePath$`.

## Comparison

| Tool | Splits long signatures | Collapses short multi-line | Configurable line length |
|------|------------------------|----------------------------|--------------------------|
| `gofmt` | no | no | no |
| `goimports` | no | no | no |
| `golines` | yes (general line wrapping) | no | yes |
| `gosigfmt` | yes (signatures only) | yes | yes |

`gosigfmt` is complementary to `gofmt` and `goimports`. Run them together; the
output of one is valid input for the other.

## Limitations

- Does not reformat call sites, struct literals, or non-signature constructs.
- Does not touch `func` types in struct fields or variable declarations.
- No first-party editor plugin; CLI only.

## Development

The project ships a `Makefile` for common tasks:

```bash
make lint        # format + run golangci-lint with --fix (via Docker)
make test        # go test -race ./...
make build       # build the CLI to bin/gosigfmt
make dogfood     # build and run gosigfmt -l ./...
make all         # lint + test + build + dogfood
```

`make lint` runs `golangci-lint` inside its official Docker image, so no local install is required (only `docker` or `podman` via `CONTAINER_ENGINE=podman`).

## License

MIT — see [LICENSE](LICENSE).
