---
name: golangci-lint-legibility
description: Use, configure, diagnose, benchmark, and extend golangci-lint-legibility. Use when working with LEG diagnostics, legibility settings, Go readability findings, linter performance, or new and changed analyzer rules.
---

# Go legibility

Choose the matching branch:

- For a `LEG###` finding or consumer configuration, follow **Consumer loop**.
- For an analyzer bug or rule change, follow **Contributor loop**.
- For a speed claim, follow **Performance loop** before changing code.

## Consumer loop

1. Inspect the repository's existing entry points:

   ```sh
   rg -n "legibility|enabled-rules|disabled-rules" .golangci.y*ml .custom-gcl.y*ml Makefile 2>/dev/null
   ```

   Reuse its pinned `golangci-lint` version, custom binary path, and wrapper command.

2. Run the narrowest target that reproduces the finding. For a working-tree change, start with:

   ```sh
   ./bin/legibility-golangci-lint run --new-from-rev=HEAD ./...
   ```

   For a package-specific finding, run `./bin/legibility-golangci-lint run ./path/to/package/...`. Use the repository's configured command when it differs. Build the custom binary when it is missing or its plugin version changed.

3. Read each diagnostic as `code rule-name: message`. Locate the rule before editing:

   ```sh
   rg -n "LEG[0-9]{3}|rule-name" README.md internal/analyzers .golangci.y*ml
   ```

   In a consumer repository without the analyzer source, consult the upstream [rule catalog](https://github.com/yowainwright/golangci-lint-legibility#rules).

4. Fix the reported construct with the smallest semantics-preserving edit. Use named intermediate values, guard clauses, focused functions, or idiomatic Go when the selected rule calls for them.

5. Rerun the same package until the targeted diagnostics are gone. Run the repository's broader lint and test checks once.

### Configuration semantics

- Treat a non-empty `enabled-rules` list as an allowlist.
- Apply `disabled-rules` after the allowlist to subtract exceptions.
- Select rules by `LEG###`, rule name, or `all`.
- Keep thresholds in `.golangci.yml`; keep plugin and `golangci-lint` versions in `.custom-gcl.yml`.
- Start with defaults. Add opt-in rules or threshold overrides for an explicit repository policy.

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

Finish with the diagnostics or behavior changed, the narrow and broad checks executed, and benchmark evidence for every performance claim. State any intentionally skipped Docker or full-repository check.
