# Contributing

## Development

<!-- local development commands derived from go.mod, .custom-gcl.yml, .golangci.yml, Makefile, and internal/analyzers/performance_test.go -->

Install Go and `golangci-lint` v2, then run:

```sh
go mod download
make tidy-check
make test
make vet
make e2e
make lint
go test ./internal/analyzers -run '^$' -bench Benchmark -benchmem -count=10
```

`make e2e` requires Docker. It uses the [E2E image](../tests/e2e/Dockerfile) to build the custom linter and run it against the fixture projects. `make lint` builds `./bin/legibility-golangci-lint` from `.custom-gcl.yml` and checks this repository.

## Code Style

Keep analyzer logic syntax-only unless a rule explicitly needs type information. Prefer small helpers, early returns, and named intermediate values for complex conditions.

## Pull Requests

Open focused pull requests with tests for new or changed rules. Include a short before-and-after example when changing diagnostics or settings.
