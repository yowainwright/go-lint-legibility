# golangci-lint-legibility

<!-- package badges derived from GitHub workflows, go.mod, LICENSE, and release tags -->

[![CI](https://github.com/yowainwright/golangci-lint-legibility/actions/workflows/ci.yml/badge.svg)](https://github.com/yowainwright/golangci-lint-legibility/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/yowainwright/golangci-lint-legibility.svg)](https://pkg.go.dev/github.com/yowainwright/golangci-lint-legibility)
[![GitHub release](https://img.shields.io/github/v/release/yowainwright/golangci-lint-legibility)](https://github.com/yowainwright/golangci-lint-legibility/releases)
[![license](https://img.shields.io/github/license/yowainwright/golangci-lint-legibility)](LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/yowainwright/golangci-lint-legibility/badge)](https://scorecard.dev/viewer/?uri=github.com/yowainwright/golangci-lint-legibility)

Syntax-only readability and comment-policy rules for Go. The rules favor named values, shallow control flow, predictable filenames, and comments that match repository policy.

## Install

This project is a `golangci-lint` module plugin. Install it through Homebrew or build a custom `golangci-lint` binary.

### Custom build

<!-- consumer custom-gcl config derived from go.mod module path and plugin package path -->

Create `.custom-gcl.yml` in the project that wants to use the linter:

```yaml
version: v2.12.2
name: legibility-golangci-lint
destination: ./bin
plugins:
  - module: github.com/yowainwright/golangci-lint-legibility
    import: github.com/yowainwright/golangci-lint-legibility/plugin
    version: v0.3.0
```

Install `golangci-lint`, then build the custom binary:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
golangci-lint custom
```

### Homebrew

<!-- Homebrew tap commands derived from Formula/golangci-lint-legibility.rb -->

```sh
brew tap yowainwright/golangci-lint-legibility https://github.com/yowainwright/golangci-lint-legibility
brew install golangci-lint-legibility
```

### Agent skill

<!-- agent workflow derived from skills/golangci-lint-legibility -->

Install the [companion Agent Skill](skills/golangci-lint-legibility) from this repository to configure the linter, resolve `LEG###` diagnostics, change analyzer rules, or measure performance. Invoke it as `$golangci-lint-legibility`.

## Rules

<!-- rules derived from internal/analyzers/analyzers.go and analyzer constructors -->

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

<!-- do/don't examples for non-comment rules registered in internal/analyzers/analyzers.go -->

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

<!-- comment rule examples derived from tests/e2e/comments_test.go and tests/e2e/testdata/comments -->

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

<!-- idiom rule examples derived from https://go.dev/doc/effective_go and https://google.github.io/styleguide/go/best-practices.html -->

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

## Configure

<!-- golangci-lint configuration derived from .golangci.yml and internal/analyzers/settings.go -->

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
        original-url: github.com/yowainwright/golangci-lint-legibility
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
          # enabled-rules is an exclusive replacement list; this example
          # enables only the two opt-in naming rules.
          enabled-rules:
            - prefer-verb-function-names
            - prefer-boolean-prefixes
          disabled-rules:
            - prefer-guard-clauses
```

Run:

```sh
./bin/legibility-golangci-lint run ./...
./bin/legibility-golangci-lint fmt ./...
```

See the `golangci-lint` [module plugin docs](https://golangci-lint.run/docs/plugins/module-plugins/) for the custom binary workflow. Module plugins require a build step because the plugin and host binary must share the same Go toolchain version and build environment.

## Recipes

<!-- workflow recipes derived from consumer binary names, golangci-lint run flags, CI checkout behavior, comment policy settings, and skills/golangci-lint-legibility -->

The examples use the locally built binary. Homebrew users can omit `./bin/`.

### Comment rules

<!-- comment recipes derived from tests/e2e/comments_test.go and tests/e2e/testdata/comments -->

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

Use the [companion Agent Skill](skills/golangci-lint-legibility) for the complete consumer, contributor, and performance workflow.

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

<!-- settings derived from internal/analyzers/settings.go -->

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

## Develop

<!-- development commands derived from Makefile, internal/analyzers/performance_test.go, and tests/e2e/Dockerfile -->

```sh
make tidy-check
make test
make vet
make e2e
make lint
go test ./internal/analyzers -run '^$' -bench Benchmark -benchmem -count=10
```

`make e2e` requires Docker. It uses the [E2E image](tests/e2e/Dockerfile) to build the custom linter and run it against the fixture projects. `make lint` builds the custom binary locally and runs it against this repository.
