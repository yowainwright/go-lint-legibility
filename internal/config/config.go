package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"go.yaml.in/yaml/v3"

	"github.com/yowainwright/go-lint-legibility/internal/analyzers"
)

type Source struct {
	Path   string
	Shared bool
}

type candidate struct {
	name   string
	shared bool
}

type document struct {
	Version int                        `json:"version"`
	Rules   map[string]json.RawMessage `json:"rules"`
}

type entry struct {
	severity string
	options  map[string]any
}

type overlay struct {
	values     map[string]any
	selectors  map[string]string
	disabled   []string
	additional []string
}

const (
	sharedKey      = "go-lint-legibility"
	formatVersion  = 1
	maxConfigBytes = 1 << 20
)

//go:embed options.json
var optionKeysJSON string

func Load(explicit string, start string) (analyzers.Settings, Source, error) {
	source, found, err := locate(explicit, start)
	if err != nil {
		return analyzers.Settings{}, Source{}, err
	}
	if !found {
		return analyzers.Settings{}, Source{}, nil
	}

	settings, err := read(source)
	return settings, source, err
}

func locate(explicit string, start string) (Source, bool, error) {
	if explicit != "" {
		isShared := strings.HasPrefix(filepath.Base(explicit), ".legibilityrc")
		return Source{Path: explicit, Shared: isShared}, true, nil
	}

	return discover(start)
}

func discover(start string) (Source, bool, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return Source{}, false, err
	}

	return walkUp(dir)
}

func walkUp(dir string) (Source, bool, error) {
	for {
		source, found, err := inspect(dir)
		if err != nil {
			return Source{}, false, err
		}
		if found {
			return source, true, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return Source{}, false, nil
		}
		dir = parent
	}
}

func inspect(dir string) (Source, bool, error) {
	active, err := activeSources(dir)
	if err != nil {
		return Source{}, false, err
	}
	if len(active) > 1 {
		return Source{}, false, fmt.Errorf(
			"multiple active configs in %s: %s",
			dir,
			describe(active),
		)
	}
	if len(active) == 0 {
		return Source{}, false, nil
	}

	return active[0], true, nil
}

func activeSources(dir string) ([]Source, error) {
	var active []Source
	for _, option := range candidates() {
		source := Source{Path: filepath.Join(dir, option.name), Shared: option.shared}
		isActive, err := isActive(source)
		if err != nil {
			return nil, err
		}
		if isActive {
			active = append(active, source)
		}
	}

	return active, nil
}

func candidates() []candidate {
	return []candidate{
		{".go-lint-legibilityrc", false},
		{".go-lint-legibilityrc.json", false},
		{".go-lint-legibilityrc.yaml", false},
		{".go-lint-legibilityrc.yml", false},
		{".go-lint-legibilityrc.toml", false},
		{".legibilityrc", true},
		{".legibilityrc.json", true},
		{".legibilityrc.yaml", true},
		{".legibilityrc.yml", true},
		{".legibilityrc.toml", true},
	}
}

func isActive(source Source) (bool, error) {
	data, err := readConfigFile(source.Path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !source.Shared {
		return true, nil
	}

	_, found, err := section(source.Path, data)
	if err != nil {
		return false, fmt.Errorf("%s: %w", source.Path, err)
	}

	return found, nil
}

func describe(sources []Source) string {
	names := make([]string, 0, len(sources))
	for _, source := range sources {
		names = append(names, filepath.Base(source.Path))
	}

	return strings.Join(names, ", ")
}

func read(source Source) (analyzers.Settings, error) {
	data, err := readConfigFile(source.Path)
	if err != nil {
		return analyzers.Settings{}, fmt.Errorf("read %s: %w", source.Path, err)
	}

	body, err := body(source.Path, data, source.Shared)
	if err != nil {
		return analyzers.Settings{}, fmt.Errorf("%s: %w", source.Path, err)
	}

	settings, err := decode(body)
	if err != nil {
		return analyzers.Settings{}, fmt.Errorf("%s: %w", source.Path, err)
	}

	return settings, nil
}

func readConfigFile(path string) ([]byte, error) {
	file, err := openConfigFile(path)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(io.LimitReader(file, maxConfigBytes+1))
	closeErr := file.Close()
	readFailed := err != nil || closeErr != nil
	if readFailed {
		return nil, errors.Join(err, closeErr)
	}
	if len(data) > maxConfigBytes {
		return nil, fmt.Errorf("config %s exceeds %d bytes", path, maxConfigBytes)
	}

	return data, nil
}

func openConfigFile(path string) (*os.File, error) {
	if err := validateConfigPath(path); err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	if err := validateOpenedConfigFile(path, file); err != nil {
		return nil, errors.Join(err, file.Close())
	}

	return file, nil
}

func validateConfigPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	return validateConfigFile(path, info)
}

func validateOpenedConfigFile(path string, file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}

	return validateConfigFile(path, info)
}

func validateConfigFile(path string, info os.FileInfo) error {
	if !info.Mode().IsRegular() {
		return fmt.Errorf("config %s is not a regular file", path)
	}
	if info.Size() > maxConfigBytes {
		return fmt.Errorf("config %s exceeds %d bytes", path, maxConfigBytes)
	}

	return nil
}

func body(path string, data []byte, shared bool) ([]byte, error) {
	if !shared {
		return toJSON(path, data)
	}

	body, found, err := section(path, data)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("no %q section", sharedKey)
	}

	return body, nil
}

func section(path string, data []byte) ([]byte, bool, error) {
	shared, err := parseTable(path, data)
	if err != nil {
		return nil, false, err
	}

	value, found := shared[sharedKey]
	if !found {
		return nil, false, nil
	}

	body, err := json.Marshal(value)
	return body, true, err
}

func toJSON(path string, data []byte) ([]byte, error) {
	table, err := parseTable(path, data)
	if err != nil {
		return nil, err
	}

	return json.Marshal(table)
}

func parseTable(path string, data []byte) (map[string]any, error) {
	var table map[string]any
	if filepath.Ext(path) == ".toml" {
		return table, wrapParse(toml.Unmarshal(data, &table))
	}

	return table, wrapParse(yaml.Unmarshal(data, &table))
}

func wrapParse(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("parse: %w", err)
}

func decode(body []byte) (analyzers.Settings, error) {
	var doc document
	if err := strictDecode(body, &doc); err != nil {
		return analyzers.Settings{}, fmt.Errorf("invalid config: %w", err)
	}
	if doc.Version != formatVersion {
		return analyzers.Settings{}, fmt.Errorf(
			"unsupported version %d: expected %d",
			doc.Version,
			formatVersion,
		)
	}

	return build(doc.Rules)
}

func strictDecode(body []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func build(rules map[string]json.RawMessage) (analyzers.Settings, error) {
	merged := overlay{
		values:    map[string]any{},
		selectors: map[string]string{},
	}
	for _, selector := range slices.Sorted(maps.Keys(rules)) {
		if err := merged.apply(selector, rules[selector]); err != nil {
			return analyzers.Settings{}, fmt.Errorf("rule %s: %w", selector, err)
		}
	}

	return merged.settings()
}

func (o *overlay) apply(selector string, raw json.RawMessage) error {
	rule, found := lookup(selector)
	if !found {
		return errors.New("unknown rule")
	}
	if previous, found := o.selectors[rule.Name]; found {
		return fmt.Errorf("rule %s specified by both %q and %q", rule.Name, previous, selector)
	}
	o.selectors[rule.Name] = selector

	parsed, err := parseEntry(raw)
	if err != nil {
		return err
	}
	if err := o.severity(rule, parsed.severity); err != nil {
		return err
	}

	return o.options(rule, parsed.options)
}

func lookup(selector string) (analyzers.Rule, bool) {
	for _, rule := range analyzers.Rules() {
		isMatch := strings.EqualFold(selector, rule.Code) || strings.EqualFold(selector, rule.Name)
		if isMatch {
			return rule, true
		}
	}

	return analyzers.Rule{}, false
}

func parseEntry(raw json.RawMessage) (entry, error) {
	var severity string
	if err := json.Unmarshal(raw, &severity); err == nil {
		return entry{severity: severity}, nil
	}

	var tuple []json.RawMessage
	if err := json.Unmarshal(raw, &tuple); err != nil {
		return entry{}, errors.New(`expected "off", "error", or [severity, options]`)
	}

	return entryFromTuple(tuple)
}

func entryFromTuple(tuple []json.RawMessage) (entry, error) {
	hasWrongLength := len(tuple) == 0 || len(tuple) > 2
	if hasWrongLength {
		return entry{}, errors.New("expected [severity] or [severity, options]")
	}

	var parsed entry
	if err := json.Unmarshal(tuple[0], &parsed.severity); err != nil {
		return entry{}, errors.New("severity must be a string")
	}
	if len(tuple) == 1 {
		return parsed, nil
	}

	return withOptions(parsed, tuple[1])
}

func withOptions(parsed entry, raw json.RawMessage) (entry, error) {
	if err := json.Unmarshal(raw, &parsed.options); err != nil {
		return entry{}, errors.New("options must be an object")
	}

	return parsed, nil
}

func (o *overlay) severity(rule analyzers.Rule, severity string) error {
	switch severity {
	case "off":
		o.disabled = append(o.disabled, rule.Name)
	case "error":
		if !rule.DefaultEnabled {
			o.additional = append(o.additional, rule.Name)
		}
	case "warn":
		return errors.New(`severity "warn" is not supported yet; use "error" or "off"`)
	default:
		return fmt.Errorf("unknown severity %q: use \"error\" or \"off\"", severity)
	}

	return nil
}

func (o *overlay) options(rule analyzers.Rule, options map[string]any) error {
	keys := optionKeys()[rule.Name]
	for _, option := range slices.Sorted(maps.Keys(options)) {
		key, found := keys[option]
		if !found {
			return fmt.Errorf("unknown option %q", option)
		}
		o.values[key] = options[option]
	}

	return nil
}

func optionKeys() map[string]map[string]string {
	var keys map[string]map[string]string
	if err := json.Unmarshal([]byte(optionKeysJSON), &keys); err != nil {
		return map[string]map[string]string{}
	}

	return keys
}

func (o *overlay) settings() (analyzers.Settings, error) {
	o.values["disabled-rules"] = o.disabled
	o.values["additional-rules"] = o.additional

	body, err := json.Marshal(o.values)
	if err != nil {
		return analyzers.Settings{}, err
	}

	var settings analyzers.Settings
	if err := strictDecode(body, &settings); err != nil {
		return analyzers.Settings{}, fmt.Errorf("invalid option value: %w", err)
	}

	return settings, nil
}
