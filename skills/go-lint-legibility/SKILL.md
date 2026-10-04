---
name: go-lint-legibility
description: Set up, configure, diagnose, benchmark, and extend go-lint-legibility. Use when installing the linter, working with LEG diagnostics or settings, reviewing Go readability findings, measuring performance, or changing analyzer rules.
---

# Go legibility

Choose the matching branch:

- For first-time installation or integration, follow **Setup loop**.
- For a `LEG###` finding or consumer configuration, follow **Consumer loop**.
- For an analyzer bug or rule change, follow **Contributor loop**.
- For a speed claim, follow **Performance loop** before changing code.

## Setup loop

1. Inspect the project's Go version, commands, and existing configuration before changing anything:

   ```sh
   go version
   rg --files --hidden -g 'go.mod' -g 'Makefile' -g '.golangci.y*ml' -g '.custom-gcl.y*ml' -g '.go-lint-legibilityrc*' -g '.legibilityrc*'
   ```

   Reuse the project's existing linter command and pinned versions. If the user asks how to set up the linter, explain the steps; if they ask you to set it up, complete and verify the chosen path. For a new setup with no stated preference or existing `golangci-lint` workflow, use the standalone Go command; it needs no custom build or config file.

2. For standalone use, install and run the module:

   ```sh
   go install github.com/yowainwright/go-lint-legibility/cmd/go-lint-legibility@latest
   ```

   If the command is not found, check whether `GOBIN` or `GOPATH/bin` is on `PATH`. `go vet` can run the analyzer with `go vet -vettool="$(which go-lint-legibility)" ./...`.

   Run the installed binary on the module:

   ```sh
   go-lint-legibility ./...
   ```

3. Leave configuration out when defaults are enough. If the user needs rule overrides, use a supported `.go-lint-legibilityrc*` file, or a shared `.legibilityrc*` file with settings under the `go-lint-legibility` key. Point to the README's [standalone configuration guide](https://github.com/yowainwright/go-lint-legibility#configure) for accepted formats and examples. Do not translate plugin settings directly into standalone config; their schemas differ.

4. For a project that already uses `golangci-lint`, or when the user explicitly chooses the plugin, follow the README's [plugin setup](https://github.com/yowainwright/go-lint-legibility#golangci-lint-plugin): pin the `golangci-lint` and plugin module versions in `.custom-gcl.yml`, build the custom binary, enable `legibility` in `.golangci.yml`, and invoke that binary. Keep the plugin and host on the same Go toolchain. Do not replace an existing lint workflow without the user's direction.

5. Run the selected linter command on the project. Exit code `0` means no diagnostics, `3` means diagnostics were found, and `2` indicates invalid configuration. Fix the config when it exits `2`; treat code `3` as lint findings, not a failed installation.

## Consumer loop

1. Inspect the repository's existing entry points:

   ```sh
   rg -n "legibility|enabled-rules|disabled-rules" .golangci.y*ml .custom-gcl.y*ml .go-lint-legibilityrc* .legibilityrc* Makefile 2>/dev/null
   ```

   Reuse its pinned tool version, binary path, and wrapper command. If the project uses the standalone command, keep using it.

2. Run the narrowest target that reproduces the finding. For standalone use:

   ```sh
   go-lint-legibility ./path/to/package/...
   ```

   For plugin use, run the project's custom `golangci-lint` binary on the package. Use the repository's configured command when it differs. Build the custom binary when it is missing or its plugin version changed.

3. Read each diagnostic as `code rule-name: message`. Locate the rule before editing:

   ```sh
   rg -n "LEG[0-9]{3}|rule-name" README.md internal/analyzers .golangci.y*ml
   ```

   In a consumer repository without the analyzer source, consult the upstream [rule catalog](https://github.com/yowainwright/go-lint-legibility#rules).

4. Fix the reported construct with the smallest semantics-preserving edit. Use named intermediate values, guard clauses, focused functions, or idiomatic Go when the selected rule calls for them.

5. Rerun the same package until the targeted diagnostics are gone. Run the repository's broader lint and test checks once.

### Plugin configuration semantics

- In `.golangci.yml`, treat a non-empty `enabled-rules` list as an allowlist.
- Apply `disabled-rules` after the allowlist to subtract exceptions.
- Select rules by `LEG###`, rule name, or `all`.
- Keep thresholds in `.golangci.yml`; keep plugin and `golangci-lint` versions in `.custom-gcl.yml`.
- Start with defaults. Add opt-in rules or threshold overrides for an explicit repository policy.

For standalone configuration, use the `rules` map described in the README. Rules not listed keep their defaults; `"error"` enables a rule and `"off"` disables it.

## Contributor loop

1. Locate the complete rule surface:

   ```sh
   rg -n 'LEG[0-9]{3}|rule-name' internal/analyzers README.md tests .golangci.yml
   ```

   Account for the registry entry, analyzer implementation, settings, unit fixtures, end-to-end fixtures, and public rule catalog.

2. Reproduce a bug with the smallest analyzer test or add a failing case for new behavior. Include nearby negative cases that must remain quiet.

3. Keep analyzer logic syntax-only. Reuse existing AST helpers and diagnostic formatting. Preserve generated-file, test-file, and Go-directive exceptions used by neighboring rules.

4. Run fast checks while iterating:

   ```sh
   gofmt -w <changed-go-files>
   go test ./internal/analyzers
   ```

5. Expand validation once the focused loop is green:

   ```sh
   go test ./...
   go vet ./...
   ```

   Run `make e2e` for plugin wiring, settings decoding, diagnostic integration, or fixture changes. Run `make lint` when the custom binary needs a self-check. Run `make check` once for release-level validation.

6. Verify every changed behavior is represented by tests and, when user-visible, by the README rule table, example, and settings table.

## Performance loop

1. Record the exact repository, package target, enabled rules, Go version, `golangci-lint` version, and cache state.
2. Measure the same command several times and compare warm runs separately from the first run.
3. Profile before combining or rewriting analyzers. Check repeated AST walks, parent-map construction, source rendering, regular-expression work, and per-file normalization first.
4. Make one focused optimization and retain identical diagnostics with analyzer and end-to-end tests.
5. Report before and after medians with the command and fixture size. Treat an unmeasured speedup as a hypothesis.

## Completion

For setup work, report the chosen integration, files or commands changed, and the validation command and result. For rule changes, report the diagnostics or behavior changed and checks executed. Include benchmark evidence for performance claims and state any intentionally skipped Docker or full-repository check.
