package legibility

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/yowainwright/golangci-lint-legibility/internal/analyzers"
	"golang.org/x/tools/go/analysis"
)

const Name = "legibility"

type Plugin struct {
	settings analyzers.Settings
}

func init() {
	register.Plugin(Name, New)
}

func New(conf any) (register.LinterPlugin, error) {
	settings, err := decodeSettings(conf)
	if err != nil {
		return nil, err
	}

	return &Plugin{settings: settings}, nil
}

func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return analyzers.New(p.settings), nil
}

//nolint:legibility // LEG042: register.LinterPlugin requires this method name.
func (p *Plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}

func decodeSettings(conf any) (analyzers.Settings, error) {
	if conf == nil {
		return analyzers.Settings{}, nil
	}

	return register.DecodeSettings[analyzers.Settings](conf)
}
