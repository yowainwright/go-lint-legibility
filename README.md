# go-lint-legibility

[![CI](https://github.com/yowainwright/go-lint-legibility/actions/workflows/ci.yml/badge.svg)](https://github.com/yowainwright/go-lint-legibility/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/yowainwright/go-lint-legibility.svg)](https://pkg.go.dev/github.com/yowainwright/go-lint-legibility)
[![GitHub release](https://img.shields.io/github/v/release/yowainwright/go-lint-legibility)](https://github.com/yowainwright/go-lint-legibility/releases)
[![license](https://img.shields.io/github/license/yowainwright/go-lint-legibility)](LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/yowainwright/go-lint-legibility/badge)](https://scorecard.dev/viewer/?uri=github.com/yowainwright/go-lint-legibility)

Go lint rules for readable expressions, control flow, names, and comments.

Run it as a standalone command or as an optional `golangci-lint` module plugin.

## Quick start

```sh
go install github.com/yowainwright/go-lint-legibility/cmd/go-lint-legibility@latest
```

Make sure Go's install directory (`GOBIN`, or `GOPATH/bin` by default) is on `PATH`.

```sh
go-lint-legibility ./...
```

Release binaries for macOS and Linux (arm64 and amd64) are attached to each [GitHub release](https://github.com/yowainwright/go-lint-legibility/releases).

> [!NOTE]
> `v0.3.1` was the final release under the name `golangci-lint-legibility`. Starting with `v0.4.0`, the module path is `github.com/yowainwright/go-lint-legibility`.

To run it through `go vet` instead:

```sh
go vet -vettool="$(which go-lint-legibility)" ./...
```

Exit codes: `0` for no diagnostics, `3` when diagnostics are reported, and `2` for a configuration error.

## Configure

Configuration is optional; default-on rules use go-lint-legibility defaults. The linter searches the current directory and its parents, using the nearest config. Multiple supported config files in one directory or unknown settings are errors (exit `2`).

| File | Format |
| --- | --- |
| `.go-lint-legibilityrc` | JSON or YAML |
| `.go-lint-legibilityrc.json`, `.yaml`, `.yml` | JSON or YAML |
| `.go-lint-legibilityrc.toml` | TOML |
| `.legibilityrc`, `.legibilityrc.json`, `.yaml`, `.yml` | JSON or YAML, under the `go-lint-legibility` key |
| `.legibilityrc.toml` | TOML, under the `[go-lint-legibility]` table |

```json
{
  "version": 1,
  "rules": {
    "max-function-lines": ["error", { "max": 40 }],
    "no-computed-values": ["error", { "max": 1 }],
    "prefer-verb-function-names": "error",
    "LEG042": "off"
  }
}
```

The same settings can be added in a shared `.legibilityrc`:

```yaml
go-lint-legibility:
  version: 1
  rules:
    max-function-lines: [error, { max: 40 }]
```

> [!NOTE]
> A shared `.legibilityrc` without a `go-lint-legibility` key is ignored, so it can hold settings for other Legibility tools. Use `-config <file>` to select a file explicitly; a file named `.legibilityrc*` is read from its `go-lint-legibility` section.

`version` must be `1`. Each entry in `rules` is keyed by rule name or `LEG###` code and is either a severity or `[severity, options]`.

| Severity | Meaning |
| --- | --- |
| `"error"` | Run the rule. This also turns on an opt-in rule. |
| `"off"` | Do not run the rule. |

`"warn"` is reserved and not supported yet; it is rejected so a rule is never silently downgraded. Rules you do not list keep their defaults.

See the [settings reference](#settings) for all rule options and defaults.

### Suppress a diagnostic

go-lint-legibility honors the `golangci-lint` comment syntax, so the same comments work with either way of running it:

```go
if ready && enabled { //nolint:LEG002 // reason
}

//nolint:legibility // every rule, with a reason
func GetLoadMode() string {
```

A directive suppresses diagnostics on its own line, or on the line directly below it when it sits alone on a line. A blank line between the comment and the code ends the suppression. Selectors are `legibility` (every rule), a `LEG###` code, or a rule name, separated by commas. A bare `//nolint` suppresses every linter, as it does in `golangci-lint`. Unlike `golangci-lint`, a directive does not cover a whole block. For a wider exception, turn the rule off in the configuration file.

## golangci-lint plugin

Use this path if your project already runs `golangci-lint`. It needs a custom `golangci-lint` build, which is larger than the standalone binary. Plugin settings go in `.golangci.yml`, not the standalone config files above.

Create `.custom-gcl.yml` in the project that wants to use the linter:

```yaml
version: v2.14.0
name: legibility-golangci-lint
destination: ./bin
plugins:
  - module: github.com/yowainwright/go-lint-legibility
    import: github.com/yowainwright/go-lint-legibility/plugin
    version: v0.4.0
```

Install `golangci-lint`, then build the custom binary:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
golangci-lint custom
```

### Configure the plugin

Add `legibility` to `.golangci.yml`:

```yaml
version: "2"

linters:
  default: standard
  enable:
    - legibility
  settings:
    custom:
      legibility:
        type: module
        description: Syntax-only Go legibility rules.
        original-url: github.com/yowainwright/go-lint-legibility
        settings:
          max-expression-operators: 4
          max-if-operators: 0
          max-control-flow-depth: 3
          max-array-chain-depth: 2
          max-computed-value-operators: 1
          min-object-lookup-chain-length: 3
          max-selector-chain-depth: 3
          min-switch-chain-length: 3
          max-if-init-operators: 0
          max-composite-literal-arg-depth: 1
          max-function-lines: 20
          max-function-params: 5
          max-naked-return-lines: 5
          # This example enables only these two opt-in rules.
          enabled-rules:
            - prefer-verb-function-names
            - prefer-boolean-prefixes
          disabled-rules:
            - prefer-guard-clauses
```

Run it with the custom binary:

```sh
./bin/legibility-golangci-lint run ./...
```

See the `golangci-lint` [module plugin docs](https://golangci-lint.run/docs/plugins/module-plugins/) for details. The plugin and host binary must use the same Go toolchain version and build environment.

## Agent skill

The [Go Legibility skill](skills/go-lint-legibility/SKILL.md) can guide an agent through standalone or plugin setup, configuration, `LEG###` findings, analyzer changes, and performance measurements. Invoke it with `$go-lint-legibility`.

The standalone command is the simplest setup. Use the plugin section above when the project already relies on `golangci-lint`.

## Known limits

- `"warn"` severity is rejected; every diagnostic is an error.
- `//nolint` covers its own line, or the line below when alone. It has no block scope.
- In `go vet` mode, config is discovered from the working directory upward. Use an absolute `-config` path; a relative one is untested.
- The `golangci-lint` plugin reads settings from `.golangci.yml`, not the config files above.

## Rules

Removed lines are don'ts. Added lines are dos.

| Code | Rule | Summary |
| --- | --- | --- |
| [`LEG001`](#leg001-max-expression-operators) | `max-expression-operators` | Limit operators inside a single expression. |
| [`LEG002`](#leg002-hoist-if-operators) | `hoist-if-operators` | Prefer named booleans before operator-heavy conditions. |
| [`LEG003`](#leg003-max-control-flow-depth) | `max-control-flow-depth` | Limit nested control-flow depth. |
| [`LEG005`](#leg005-no-quadratic-patterns) | `no-quadratic-patterns` | Flag likely quadratic nested loops. |
| [`LEG006`](#leg006-no-redundant-boolean-logic) | `no-redundant-boolean-logic` | Avoid redundant boolean comparisons. |
| [`LEG007`](#leg007-prefer-positive-condition-names) | `prefer-positive-condition-names` | Prefer positive condition names. |
| [`LEG008`](#leg008-no-trivial-wrapper-functions) | `no-trivial-wrapper-functions` | Avoid functions that only forward parameters to another call. |
| [`LEG009`](#leg009-prefer-early-return) | `prefer-early-return` | Avoid else branches after a branch already exits. |
| [`LEG010`](#leg010-prefer-guard-clauses) | `prefer-guard-clauses` | Prefer guard clauses over wrapping the main path in one large if block. |
| [`LEG011`](#leg011-max-array-chain-depth) | `max-array-chain-depth` | Limit consecutive collection-style method chains. |
| [`LEG012`](#leg012-no-computed-values) | `no-computed-values` | Prefer named values before returning computed expressions. |
| [`LEG024`](#leg024-prefer-object-lookup) | `prefer-object-lookup` | Prefer set or map lookups over long equality-or chains. |
| [`LEG025`](#leg025-require-filename-matches-dirname) | `require-filename-matches-dirname` | Opt-in. Require files in named subdirectories to match the directory name. |
| [`LEG026`](#leg026-no-mixed-filename-casing) | `no-mixed-filename-casing` | Avoid filenames that mix casing conventions. |
| [`LEG031`](#leg031-no-deep-selector-chain) | `no-deep-selector-chain` | Avoid deep selector or index chains without named intermediate values. |
| [`LEG034`](#leg034-prefer-switch-over-long-if-chain) | `prefer-switch-over-long-if-chain` | Prefer switch over long if chains that compare the same value. |
| [`LEG035`](#leg035-no-bool-literal-args) | `no-bool-literal-args` | Avoid boolean literals as call arguments. |
| [`LEG036`](#leg036-no-complex-if-init) | `no-complex-if-init` | Avoid combining an if initializer with an operator-heavy condition. |
| [`LEG037`](#leg037-no-deep-composite-literal-arg) | `no-deep-composite-literal-arg` | Avoid deeply nested composite literals as call arguments. |
| [`LEG038`](#leg038-max-function-lines) | `max-function-lines` | Limit functions to a focused line budget. |
| [`LEG039`](#leg039-no-unmatched-comments) | `no-unmatched-comments` | Policy opt-in. Reject comments without an allowed matcher or boundary identifier. |
| [`LEG040`](#leg040-no-automated-comment-attribution) | `no-automated-comment-attribution` | Reject explicit automated source signatures in comments. |
| [`LEG041`](#leg041-prefer-line-comments) | `prefer-line-comments` | Prefer `//` comments for ordinary multiline prose. |
| [`LEG042`](#leg042-no-getter-prefix) | `no-getter-prefix` | Avoid the `Get` prefix on accessor names. |
| [`LEG043`](#leg043-no-underscore-names) | `no-underscore-names` | Prefer `MixedCaps` over underscores in declared names. |
| [`LEG044`](#leg044-prefer-initialism-casing) | `prefer-initialism-casing` | Keep initialisms such as `ID` and `URL` fully capitalized. |
| [`LEG045`](#leg045-no-package-name-stutter) | `no-package-name-stutter` | Avoid repeating the package name in exported names. |
| [`LEG046`](#leg046-no-generic-package-names) | `no-generic-package-names` | Avoid catch-all package names such as `util` or `common`. |
| [`LEG047`](#leg047-prefer-range-loop) | `prefer-range-loop` | Prefer a range clause over an index counter loop. |
| [`LEG048`](#leg048-no-redundant-break) | `no-redundant-break` | Remove trailing `break` statements from case clauses. |
| [`LEG049`](#leg049-no-naked-returns) | `no-naked-returns` | Avoid naked returns outside short functions. |
| [`LEG050`](#leg050-prefer-lowercase-error-strings) | `prefer-lowercase-error-strings` | Prefer lowercase error strings without trailing punctuation. |
| [`LEG051`](#leg051-max-function-params) | `max-function-params` | Limit the number of parameters in a signature. |
| [`LEG052`](#leg052-prefer-verb-function-names) | `prefer-verb-function-names` | Opt-in. Prefer action verbs at the start of function names. |
| [`LEG053`](#leg053-prefer-boolean-prefixes) | `prefer-boolean-prefixes` | Opt-in. Prefer predicate or state prefixes for boolean names. |

<a id="leg001-max-expression-operators"></a>

### `LEG001 max-expression-operators`

#### do / don't

```diff
- return user.Active && user.Score > 10 && (user.Role == "admin" || user.Role == "owner")
+ isAdmin := user.Role == "admin"
+ isOwner := user.Role == "owner"
+ hasPrivilegedRole := isAdmin || isOwner
+ return user.Active && user.Score > 10 && hasPrivilegedRole
```

<a id="leg002-hoist-if-operators"></a>

### `LEG002 hoist-if-operators`

#### do / don't

```diff
- if user != nil && user.Active && !user.Locked {
- 	sendInvite(user)
- }
+ canInviteUser := user != nil && user.Active && !user.Locked
+ if canInviteUser {
+ 	sendInvite(user)
+ }
```

<a id="leg003-max-control-flow-depth"></a>

### `LEG003 max-control-flow-depth`

#### do / don't

```diff
- if user != nil {
- 	for _, invite := range invites {
- 		if invite.Pending {
- 			if invite.Retries < 3 {
- 				sendInvite(invite)
- 			}
- 		}
- 	}
- }
+ if user == nil {
+ 	return
+ }
+ for _, invite := range pendingInvites(invites) {
+ 	retryInvite(invite)
+ }
```

<a id="leg005-no-quadratic-patterns"></a>

### `LEG005 no-quadratic-patterns`

#### do / don't

```diff
- for _, user := range users {
- 	for _, owner := range owners {
- 		if user.ID == owner.UserID {
- 			assignOwner(user, owner)
- 		}
- 	}
- }
+ ownersByUserID := mapOwnersByUserID(owners)
+ for _, user := range users {
+ 	owner := ownersByUserID[user.ID]
+ 	assignOwner(user, owner)
+ }
```

<a id="leg006-no-redundant-boolean-logic"></a>

### `LEG006 no-redundant-boolean-logic`

#### do / don't

```diff
- return enabled == true
+ return enabled
```

<a id="leg007-prefer-positive-condition-names"></a>

### `LEG007 prefer-positive-condition-names`

#### do / don't

```diff
- isNotReady := status != StatusReady
- if isNotReady {
+ isReady := status == StatusReady
+ if !isReady {
     return nil
 }
```

<a id="leg008-no-trivial-wrapper-functions"></a>

### `LEG008 no-trivial-wrapper-functions`

#### do / don't

```diff
- func cleanName(name string) string {
- 	return strings.TrimSpace(name)
- }
-
- displayName := cleanName(input)
+ displayName := strings.TrimSpace(input)
```

<a id="leg009-prefer-early-return"></a>

### `LEG009 prefer-early-return`

#### do / don't

```diff
 if err != nil {
     return err
- } else {
- 	saveUser(user)
 }
+ saveUser(user)
```

<a id="leg010-prefer-guard-clauses"></a>

### `LEG010 prefer-guard-clauses`

#### do / don't

```diff
- if user.Active {
- 	validateUser(user)
- 	saveUser(user)
- }
+ if !user.Active {
+ 	return
+ }
+ validateUser(user)
+ saveUser(user)
```

<a id="leg011-max-array-chain-depth"></a>

### `LEG011 max-array-chain-depth`

#### do / don't

```diff
- hasLargeActiveOrder := orders.Filter(activeOrder).Map(orderTotal).Some(overLimit)
+ activeOrders := orders.Filter(activeOrder)
+ orderTotals := activeOrders.Map(orderTotal)
+ hasLargeActiveOrder := orderTotals.Some(overLimit)
```

<a id="leg012-no-computed-values"></a>

### `LEG012 no-computed-values`

#### do / don't

```diff
- return subtotal + tax + shipping
+ taxedSubtotal := subtotal + tax
+ return taxedSubtotal + shipping
```

<a id="leg024-prefer-object-lookup"></a>

### `LEG024 prefer-object-lookup`

#### do / don't

```diff
- return role == "admin" || role == "owner" || role == "staff"
+ allowedRoles := map[string]bool{
+ 	"admin": true,
+ 	"owner": true,
+ 	"staff": true,
+ }
+ return allowedRoles[role]
```

<a id="leg025-require-filename-matches-dirname"></a>

### `LEG025 require-filename-matches-dirname`

Opt-in. Enable it with `enabled-rules` when the project uses directory-mirrored filenames.

#### do / don't

```diff
- internal/orders/service.go
+ internal/orders/orders_service.go
```

<a id="leg026-no-mixed-filename-casing"></a>

### `LEG026 no-mixed-filename-casing`

#### do / don't

```diff
- internal/API_client.go
+ internal/api_client.go
```

<a id="leg031-no-deep-selector-chain"></a>

### `LEG031 no-deep-selector-chain`

#### do / don't

```diff
- return config.User.Profile.Settings.Email.Enabled
+ userProfile := config.User.Profile
+ emailSettings := userProfile.Settings.Email
+ return emailSettings.Enabled
```

<a id="leg034-prefer-switch-over-long-if-chain"></a>

### `LEG034 prefer-switch-over-long-if-chain`

#### do / don't

```diff
- if status == "new" {
- 	return 1
- } else if status == "active" {
- 	return 2
- } else if status == "closed" {
- 	return 3
- }
+ switch status {
+ case "new":
+ 	return 1
+ case "active":
+ 	return 2
+ case "closed":
+ 	return 3
+ }
```

<a id="leg035-no-bool-literal-args"></a>

### `LEG035 no-bool-literal-args`

#### do / don't

```diff
- createUser(user, true, false)
+ shouldSendInvite := true
+ shouldRequirePasswordReset := false
+ createUser(user, shouldSendInvite, shouldRequirePasswordReset)
```

<a id="leg036-no-complex-if-init"></a>

### `LEG036 no-complex-if-init`

#### do / don't

```diff
- if user, ok := users[id]; ok && user.Active {
- 	save(user)
- }
+ user, ok := users[id]
+ canSaveUser := ok && user.Active
+ if canSaveUser {
+ 	save(user)
+ }
```

<a id="leg037-no-deep-composite-literal-arg"></a>

### `LEG037 no-deep-composite-literal-arg`

#### do / don't

```diff
- save(Config{HTTP: HTTPConfig{Timeout: 10}})
+ httpConfig := HTTPConfig{Timeout: 10}
+ save(Config{HTTP: httpConfig})
```

<a id="leg038-max-function-lines"></a>

### `LEG038 max-function-lines`

With `max-function-lines: 6`:

#### do / don't

```diff
- func syncUser(user User) error {
- 	validateUser(user)
- 	normalizeUser(&user)
- 	saveUser(user)
- 	sendWelcomeEmail(user)
- 	writeAuditLog(user)
- 	refreshSearchIndex(user)
- 	return nil
- }
+ func syncUser(user User) error {
+ 	if err := prepareUser(&user); err != nil {
+ 		return err
+ 	}
+ 	return persistUser(user)
+ }
```

<a id="leg039-no-unmatched-comments"></a>

### `LEG039 no-unmatched-comments`

Opt-in. When a repository configures an allowed comment pattern, ordinary comments must match it. With a ticket matcher enabled:

#### do / don't

```diff
- // Retry requests in provider order.
+ // ENG-482: Retry requests in provider order.
```

<a id="leg040-no-automated-comment-attribution"></a>

### `LEG040 no-automated-comment-attribution`

Keep useful context, but remove comments that attribute code or prose to an automated tool.

#### do / don't

```diff
- // Generated by Codex.
+ // The provider requires retries to remain ordered.
```

<a id="leg041-prefer-line-comments"></a>

### `LEG041 prefer-line-comments`

Use Go line comments for ordinary multiline prose.

#### do / don't

```diff
- /*
- The provider requires retries to remain ordered.
- */
+ // The provider requires retries to remain ordered.
```

<a id="leg042-no-getter-prefix"></a>

### `LEG042 no-getter-prefix`

[Effective Go, Getters](https://go.dev/doc/effective_go#Getters). Reported for zero-argument functions and methods that return a value. Generated files are skipped so protocol buffer accessors stay quiet.

#### do / don't

```diff
- func (u User) GetOwner() string {
+ func (u User) Owner() string {
 	return u.owner
 }
```

<a id="leg043-no-underscore-names"></a>

### `LEG043 no-underscore-names`

[Effective Go, MixedCaps](https://go.dev/doc/effective_go#mixed-caps). Test files and generated files are skipped.

#### do / don't

```diff
- const max_retries = 3
+ const maxRetries = 3
```

<a id="leg044-prefer-initialism-casing"></a>

### `LEG044 prefer-initialism-casing`

[Google Go style, initialisms](https://google.github.io/styleguide/go/decisions#initialisms). Generated files are skipped.

#### do / don't

```diff
- type Request struct {
- 	Url    string
- 	userId string
- }
+ type Request struct {
+ 	URL    string
+ 	userID string
+ }
```

<a id="leg045-no-package-name-stutter"></a>

### `LEG045 no-package-name-stutter`

[Effective Go, package names](https://go.dev/doc/effective_go#package-names) and [Google Go style, avoid repetition](https://google.github.io/styleguide/go/best-practices#avoid-repetition). Reported when an exported name starts with the package name.

#### do / don't

```diff
- func OrdersCreate() error
+ func Create() error
```

<a id="leg046-no-generic-package-names"></a>

### `LEG046 no-generic-package-names`

[Google Go style, util packages](https://google.github.io/styleguide/go/best-practices#util-packages).

#### do / don't

```diff
- package util
+ package retry
```

<a id="leg047-prefer-range-loop"></a>

### `LEG047 prefer-range-loop`

[Effective Go, For](https://go.dev/doc/effective_go#for).
To preserve string and growing-slice semantics, the rule reports only canonical
loops over explicitly typed array or slice parameters whose binding remains
stable in the loop body.

#### do / don't

```diff
- for i := 0; i < len(scores); i++ {
- 	sum += scores[i]
- }
+ for _, score := range scores {
+ 	sum += score
+ }
```

<a id="leg048-no-redundant-break"></a>

### `LEG048 no-redundant-break`

[Effective Go, Switch](https://go.dev/doc/effective_go#switch). Go case clauses do not fall through, so a trailing `break` adds nothing. Labeled breaks are left alone.

#### do / don't

```diff
 default:
 	notify(status)
- 	break
 }
```

<a id="leg049-no-naked-returns"></a>

### `LEG049 no-naked-returns`

[Effective Go, named result parameters](https://go.dev/doc/effective_go#named-results). Reported for named results in functions longer than `max-naked-return-lines`.

#### do / don't

```diff
- 	err = validate(data)
- 	return
+ 	return data, validate(data)
 }
```

<a id="leg050-prefer-lowercase-error-strings"></a>

### `LEG050 prefer-lowercase-error-strings`

[Effective Go, Errors](https://go.dev/doc/effective_go#errors) and
[Google Go style, error strings](https://google.github.io/styleguide/go/decisions#error-strings).
Covers `errors.New` and `fmt.Errorf`, recognizes import aliases, and ignores
shadowed or unrelated selectors. Initialisms and multiword identifiers such as
`HTTP` or `TLSConfig` are left alone.

#### do / don't

```diff
- return errors.New("Invalid user")
- return fmt.Errorf("cannot read %s.", name)
+ return errors.New("invalid user")
+ return fmt.Errorf("cannot read %s", name)
```

<a id="leg051-max-function-params"></a>

### `LEG051 max-function-params`

[Google Go style, function argument lists](https://google.github.io/styleguide/go/best-practices#function-argument-lists).
Applies to declarations, literals, named function types, and interface methods.

#### do / don't

```diff
- func send(host string, port int, user string, token string, retries int, verbose bool)
+ func send(target Target, options SendOptions)
```

<a id="leg052-prefer-verb-function-names"></a>

### `LEG052 prefer-verb-function-names`

Opt-in. Flags obvious payload-first function names such as `dataFn` and
`valueFunction`. Constructors and canonical methods such as `New`, `String`,
`Error`, `Len`, `Close`, `Read`, and `Write` are exempt.

#### do / don't

```diff
- func dataFn() {}
+ func getData() {}
```

<a id="leg053-prefer-boolean-prefixes"></a>

### `LEG053 prefer-boolean-prefixes`

Opt-in. Applies to syntactically declared `bool` fields, parameters, and
results. Prefer `is`, `has`, `can`, `should`, or `will`; idiomatic states such
as `ok`, `done`, `ready`, `valid`, `enabled`, `disabled`, and `found` are
exempt. Type aliases are not inferred by this syntax-only rule.

#### do / don't

```diff
- active bool
+ isActive bool
```

The naming guidance is informed by [Effective Go](https://go.dev/doc/effective_go#names),
[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), and
[Revive's `var-naming` rule](https://github.com/mgechev/revive/blob/master/RULES_DESCRIPTIONS.md#var-naming).
This project implements the additional function and boolean heuristics as
native `go/analysis` rules; no Revive source is copied.

## Recipes

The examples use locally built binaries from `./bin/`.

### Comment rules

No ownership marker is required. Start with the base configuration from [Configure](#configure), then add only the policy your repository uses.

#### Allow ordinary comments

`LEG039` is opt-in. Leave `comment-matchers`, `comment-prefix-identifiers`, and `comment-suffix-identifiers` unset to allow ordinary comments.

The [comment E2Es](tests/e2e/comments_test.go) run the [base config](tests/e2e/testdata/comments/default/.golangci.yml) against an [ordinary comment](tests/e2e/testdata/comments/default/main.go).

#### Require ticket references

Add one matcher to the `legibility` settings:

```yaml
linters:
  settings:
    custom:
      legibility:
        settings:
          comment-matchers:
            - '\b(ENG|OPS)-\d+\b'
```

The [comment E2Es](tests/e2e/comments_test.go) run this [exact config](tests/e2e/testdata/comments/ticket/.golangci.yml) against [rejected](tests/e2e/testdata/comments/ticket/unmatched/main.go) and [accepted](tests/e2e/testdata/comments/ticket/matched/main.go) comments.

#### Reject automated attribution

`LEG040` is enabled by default and recognizes common automated-tool names. To use a project-specific list, replace the built-in identifiers:

```yaml
linters:
  settings:
    custom:
      legibility:
        settings:
          automated-comment-identifiers:
            - internal-bot
```

The [comment E2Es](tests/e2e/comments_test.go) use the [base config](tests/e2e/testdata/comments/default/.golangci.yml) with a [rejected signature](tests/e2e/testdata/comments/attribution/signature/main.go) and [accepted context](tests/e2e/testdata/comments/attribution/context/main.go).

#### Prefer line comments

`LEG041` is enabled by default. No additional configuration is required.

The [comment E2Es](tests/e2e/comments_test.go) use the [base config](tests/e2e/testdata/comments/default/.golangci.yml) with [rejected block prose](tests/e2e/testdata/comments/line-comments/block/main.go) and an [accepted line comment](tests/e2e/testdata/comments/line-comments/line/main.go).

#### Preserve Go metadata

Go directives, build constraints, generated-code markers, and cgo preambles are exempt.

The [comment E2Es](tests/e2e/comments_test.go) enable only `LEG039` with the [strict config](tests/e2e/testdata/comments/metadata/.golangci.yml), then accept a [Go directive](tests/e2e/testdata/comments/metadata/directive/main.go) and [cgo preamble](tests/e2e/testdata/comments/metadata/cgo/main.go). Comment rules do not autofix source.

### Human workflow

Use advisory mode while editing:

```sh
./bin/legibility-golangci-lint run --new --issues-exit-code=0 ./...
```

Check all staged, unstaged, and untracked changes before committing:

```sh
./bin/legibility-golangci-lint run --new-from-rev=HEAD ./...
```

Write normal Go comments. No role label is needed.

### Agent workflow

Use the [companion Agent Skill](skills/go-lint-legibility) for the complete consumer, contributor, and performance workflow.

```sh
./bin/legibility-golangci-lint run --new-from-rev=HEAD ./...
```

Agents should prefer clearer code over new comments and should not suppress diagnostics.

### CI

Lint the whole repository in required checks:

```sh
./bin/legibility-golangci-lint run ./...
```

For repositories with an existing baseline, check out full history:

```yaml
with:
  fetch-depth: 0
```

```sh
./bin/legibility-golangci-lint run --new-from-merge-base=origin/main ./...
```

Inventory the complete baseline without output caps:

```sh
./bin/legibility-golangci-lint run \
  --issues-exit-code=0 \
  --max-issues-per-linter=0 \
  --max-same-issues=0 \
  ./...
```

## Troubleshooting

**Build step fails or binary crashes at runtime** — the plugin and `golangci-lint` must be built with the same Go toolchain. Run `go version` and confirm your toolchain matches the version in `go.mod`.

## Settings

| Setting | Default | Description |
| --- | ---: | --- |
| `enabled-rules` | all | Only run matching rule codes or names. |
| `disabled-rules` | none | Skip matching rule codes or names. |
| `max-expression-operators` | 4 | Maximum readability operators in a single expression. |
| `max-if-operators` | 0 | Maximum boolean operators in an `if` condition. |
| `max-control-flow-depth` | 3 | Maximum nested control-flow depth. |
| `max-array-chain-depth` | 2 | Maximum consecutive collection-style method calls. |
| `max-computed-value-operators` | 1 | Maximum operators in returned or composite literal values. |
| `min-object-lookup-chain-length` | 3 | Minimum equality-or chain length before suggesting a set, map, or switch. |
| `min-dirname-match-depth` | 3 | Minimum directory depth for the opt-in filename/dirname rule. |
| `max-selector-chain-depth` | 3 | Maximum selector or index chain depth. |
| `min-switch-chain-length` | 3 | Minimum repeated comparison chain length before suggesting `switch`. |
| `max-if-init-operators` | 0 | Maximum boolean operators when an `if` also has an initializer. |
| `max-composite-literal-arg-depth` | 1 | Maximum nested composite literal depth in call arguments. |
| `max-function-lines` | 20 | Maximum source lines in a function declaration or literal; nested literals are measured independently. |
| `max-function-params` | 5 | Maximum parameters in a function signature. |
| `max-naked-return-lines` | 5 | Longest function that may still use a naked return. |
| `comment-matchers` | none | Case-insensitive Go regular expressions allowed anywhere in a comment group. |
| `comment-prefix-identifiers` | none | Literal identifiers allowed at the start of a normalized comment group. |
| `comment-suffix-identifiers` | none | Literal identifiers allowed at the end of a normalized comment group. |
| `automated-comment-identifiers` | built in | Names treated as automated sources in explicit source signatures. |
| `negative-condition-name-pattern` | built in | Regular expression for negative boolean names. |

Rule selectors accept rule codes such as `LEG009`, rule names such as `prefer-early-return`, or `all`. `require-filename-matches-dirname` is opt-in because ordinary Go packages often contain files that should not mirror the directory name. `no-unmatched-comments` activates when an allow path is configured or when the rule is explicitly selected.

## License

[MIT License](LICENSE)
