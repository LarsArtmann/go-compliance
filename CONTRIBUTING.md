# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

Two modules: the repo root is the publishable core library, `hub/` is the
server + CLI. A `go.work` workspace sits at the root; per-module work runs
with `GOWORK=off`, so a command sees what CI and a consumer of the tagged
core see, not the working tree.

    GOWORK=off go test -race -count=1 ./...             # core module
    cd hub && GOWORK=off go test -race -count=1 ./...   # hub module

Run `go test` invocations sequentially (a shared build cache produces
transient failures under concurrency). The quality gate is BuildFlow
(`buildflow --dry-run --verbose` lists what it would run); it owns linting
and formatting for both modules, so there is no separate lint command.
`nix flake check` builds and checks both modules sandboxed.

After editing a dashboard `.templ` file, format and generate from `hub/`
(never from the repo root, which rewrites the paths inside the generated
files), then rebuild the stylesheet from the repo root:

    cd hub && go tool templ fmt -fail ./internal/dashboard/ && go tool templ generate ./internal/dashboard/
    nix shell .#tailwindcss -c bash scripts/regen-ui.sh

Commit the generated `*_templ.go` diff together with the edit (generated
files are committed deliberately), and the three CSS artifacts
`scripts/regen-ui.sh` rewrites: CI regenerates them and fails on any
difference. Anything visual also goes through the real-browser gate,
`scripts/screenshot.sh` (themes, a narrow and a just-wide layout, contrast,
keyboard, filters), before it ships.

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
