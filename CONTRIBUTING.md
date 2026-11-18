# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

Two modules: the repo root is the publishable core library, `hub/` is the
server + CLI. A `go.work` workspace sits at the root — per-module work
needs `GOWORK=off` because the workspace hides replace-directive drift.

    GOWORK=off go test ./... -race                    # core module
    golangci-lint run ./...                           # core (root config)
    cd hub && GOWORK=off go test ./... -race          # hub module
    cd hub && golangci-lint run -c ../.golangci.yml ./...

Run `go test` invocations sequentially (a shared GOCACHE produces
transient failures under concurrency). After editing any `.templ` file,
run `go tool templ generate ./internal/dashboard/` and commit the
generated `*_templ.go` diff together with the edit (generated files are
committed deliberately). The full gate: `nix flake check`.

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
