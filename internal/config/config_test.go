package config

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/yowainwright/go-lint-legibility/internal/analyzers"
)

func TestLoadStandaloneJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(
		t,
		dir,
		".go-lint-legibilityrc",
		`{"version":1,"rules":{"max-function-lines":["error",{"max":40}]}}`,
	)

	settings, source, err := Load("", dir)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if got := pointerValue(settings.MaxFunctionLines); got != 40 {
		t.Fatalf("max-function-lines = %d, want 40", got)
	}
	if source.Shared {
		t.Fatal("standalone config reported as shared")
	}
}

func TestLoadStandaloneYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(
		t,
		dir,
		".go-lint-legibilityrc.yaml",
		"version: 1\nrules:\n  max-function-params:\n    - error\n    - max: 3\n",
	)

	settings, _, err := Load("", dir)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if got := pointerValue(settings.MaxFunctionParams); got != 3 {
		t.Fatalf("max-function-params = %d, want 3", got)
	}
}

func TestLoadStandaloneTOML(t *testing.T) {
	dir := t.TempDir()
	writeFile(
		t,
		dir,
		".go-lint-legibilityrc.toml",
		"version = 1\n\n[rules]\nmax-function-lines = [\"error\", { max = 40 }]\n",
	)

	settings, _, err := Load("", dir)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if got := pointerValue(settings.MaxFunctionLines); got != 40 {
		t.Fatalf("max-function-lines = %d, want 40", got)
	}
}

func TestLoadSharedTOMLSection(t *testing.T) {
	dir := t.TempDir()
	writeFile(
		t,
		dir,
		".legibilityrc.toml",
		"[go-lint-legibility]\nversion = 1\n\n[go-lint-legibility.rules]\nLEG042 = \"off\"\n",
	)

	settings, source, err := Load("", dir)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if !source.Shared {
		t.Fatal("shared TOML config not reported as shared")
	}
	if !slices.Contains(settings.DisabledRules, "no-getter-prefix") {
		t.Fatalf("disabled rules = %v, want no-getter-prefix", settings.DisabledRules)
	}
}

func TestInvalidTOMLFailsClosed(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".go-lint-legibilityrc.toml", "version = = 1\n")

	_, _, err := Load("", dir)
	if !errorMentions(err, "parse") {
		t.Fatalf("error = %v, want a parse error", err)
	}
}

func TestLoadSharedSection(t *testing.T) {
	dir := t.TempDir()
	writeFile(
		t,
		dir,
		".legibilityrc",
		`{"go-lint-legibility":{"version":1,"rules":{"no-getter-prefix":"off"}},"other":{}}`,
	)

	settings, source, err := Load("", dir)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if !source.Shared {
		t.Fatal("shared config not reported as shared")
	}
	if !slices.Contains(settings.DisabledRules, "no-getter-prefix") {
		t.Fatalf("disabled rules = %v, want no-getter-prefix", settings.DisabledRules)
	}
}

func TestSharedFileWithoutSectionIsIgnored(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".legibilityrc", `{"linters":{"fs-lint":{"enabled":true}}}`)

	_, source, err := Load("", dir)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if source.Path != "" {
		t.Fatalf("source = %q, want no config found", source.Path)
	}
}

func TestLoadWithoutConfigReturnsDefaults(t *testing.T) {
	settings, source, err := Load("", t.TempDir())
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if source.Path != "" {
		t.Fatalf("source = %q, want no config", source.Path)
	}
	if !reflect.DeepEqual(settings, analyzers.Settings{}) {
		t.Fatalf("settings = %+v, want zero value", settings)
	}
}

func TestDiscoveryWalksUp(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, ".go-lint-legibilityrc", `{"version":1,"rules":{}}`)

	_, source, err := Load("", child)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if want := filepath.Join(root, ".go-lint-legibilityrc"); source.Path != want {
		t.Fatalf("source = %q, want %q", source.Path, want)
	}
}

const (
	discoveredMax10 = `{"version":1,"rules":{"max-function-lines":["error",{"max":10}]}}`
	explicitMax30   = `{"version":1,"rules":{"max-function-lines":["error",{"max":30}]}}`
	optInVerbNames  = `{"version":1,"rules":{"prefer-verb-function-names":"error"}}`
)

func TestExplicitConfigWins(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".go-lint-legibilityrc", discoveredMax10)
	custom := writeFile(t, dir, "custom.json", explicitMax30)

	settings, _, err := Load(custom, dir)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if got := pointerValue(settings.MaxFunctionLines); got != 30 {
		t.Fatalf("max-function-lines = %d, want 30", got)
	}
}

func TestAmbiguousConfigsFail(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".go-lint-legibilityrc", `{"version":1,"rules":{}}`)
	writeFile(t, dir, ".legibilityrc", `{"go-lint-legibility":{"version":1,"rules":{}}}`)

	_, _, err := Load("", dir)
	if !errorMentions(err, "multiple active configs") {
		t.Fatalf("error = %v, want a multiple active configs error", err)
	}
}

func TestOptInRuleIsAdditive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".go-lint-legibilityrc", optInVerbNames)

	settings, _, err := Load("", dir)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if len(settings.EnabledRules) != 0 {
		t.Fatalf("enabled-rules = %v, want none so defaults stay on", settings.EnabledRules)
	}
	if !ruleRuns(settings, "prefer-verb-function-names") {
		t.Fatal("opt-in rule did not run")
	}
	if !ruleRuns(settings, "max-function-lines") {
		t.Fatal("default rule stopped running")
	}
}

func TestOffDisablesRule(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".go-lint-legibilityrc", `{"version":1,"rules":{"LEG038":"off"}}`)

	settings, _, err := Load("", dir)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	if ruleRuns(settings, "max-function-lines") {
		t.Fatal("rule still runs after being turned off by code")
	}
}

type invalidConfigCase struct {
	name string
	body string
	want string
}

var invalidConfigCases = []invalidConfigCase{
	{"unknown rule", `{"version":1,"rules":{"no-such-rule":"error"}}`, "unknown rule"},
	{
		"unknown option",
		`{"version":1,"rules":{"max-function-lines":["error",{"limit":3}]}}`,
		"unknown option",
	},
	{"warn", `{"version":1,"rules":{"max-function-lines":"warn"}}`, "not supported yet"},
	{"bad severity", `{"version":1,"rules":{"max-function-lines":"loud"}}`, "unknown severity"},
	{"bad version", `{"version":2,"rules":{}}`, "unsupported version"},
	{"unknown top-level key", `{"version":1,"extra":true}`, "invalid config"},
	{
		"bad option type",
		`{"version":1,"rules":{"max-function-lines":["error",{"max":"ten"}]}}`,
		"invalid option value",
	},
	{"bad entry", `{"version":1,"rules":{"max-function-lines":7}}`, "expected"},
}

func TestInvalidConfigsFailClosed(t *testing.T) {
	for _, c := range invalidConfigCases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, ".go-lint-legibilityrc", c.body)

			_, _, err := Load("", dir)
			if !errorMentions(err, c.want) {
				t.Fatalf("error = %v, want it to contain %q", err, c.want)
			}
		})
	}
}

func TestOptionTablePointsAtRealRulesAndSettings(t *testing.T) {
	known := knownRuleNames()
	tags := settingsTags()
	for rule, options := range optionKeys() {
		if !known[rule] {
			t.Errorf("options.json names %q, which is not a rule", rule)
		}
		checkOptionKeys(t, rule, options, tags)
	}
}

func checkOptionKeys(t *testing.T, rule string, options map[string]string, tags map[string]bool) {
	t.Helper()
	for option, key := range options {
		if !tags[key] {
			t.Errorf(
				"rule %s option %s maps to %q, which is not a Settings field",
				rule,
				option,
				key,
			)
		}
	}
}

func errorMentions(err error, text string) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), text)
}

func writeFile(t *testing.T, dir string, name string, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func ruleRuns(settings analyzers.Settings, name string) bool {
	for _, rule := range analyzers.Rules() {
		if rule.Name == name {
			return settings.RuleEnabled(rule.Code, rule.Name, rule.DefaultEnabled)
		}
	}

	return false
}

func knownRuleNames() map[string]bool {
	known := map[string]bool{}
	for _, rule := range analyzers.Rules() {
		known[rule.Name] = true
	}

	return known
}

func settingsTags() map[string]bool {
	tags := map[string]bool{}
	settingsType := reflect.TypeOf(analyzers.Settings{})
	for index := range settingsType.NumField() {
		name, _, _ := strings.Cut(settingsType.Field(index).Tag.Get("json"), ",")
		tags[name] = true
	}

	return tags
}

func pointerValue(value *int) int {
	if value == nil {
		return -1
	}

	return *value
}
