package analyzers

import "strings"

const (
	defaultMaxExpressionOperators      = 4
	defaultMaxIfOperators              = 0
	defaultMaxControlFlowDepth         = 3
	defaultMaxArrayChainDepth          = 2
	defaultMaxComputedValueOperators   = 1
	defaultMinObjectLookupChainLength  = 3
	defaultMinDirnameMatchDepth        = 3
	defaultMaxSelectorChainDepth       = 3
	defaultMinSwitchChainLength        = 3
	defaultMaxIfInitOperators          = 0
	defaultMaxCompositeLiteralArgDepth = 1
	defaultMaxFunctionLines            = 20
	defaultMaxFunctionParams           = 5
	defaultMaxNakedReturnLines         = 5
	defaultNegativeConditionNamePrefix = "LEG"
)

type Settings struct {
	EnabledRules                 []string `json:"enabled-rules"`
	DisabledRules                []string `json:"disabled-rules"`
	CommentMatchers              []string `json:"comment-matchers"`
	CommentPrefixIdentifiers     []string `json:"comment-prefix-identifiers"`
	CommentSuffixIdentifiers     []string `json:"comment-suffix-identifiers"`
	AutomatedCommentIdentifiers  []string `json:"automated-comment-identifiers"`
	MaxExpressionOperators       *int     `json:"max-expression-operators"`
	MaxIfOperators               *int     `json:"max-if-operators"`
	MaxControlFlowDepth          *int     `json:"max-control-flow-depth"`
	MaxArrayChainDepth           *int     `json:"max-array-chain-depth"`
	MaxComputedValueOperators    *int     `json:"max-computed-value-operators"`
	MinObjectLookupChainLength   *int     `json:"min-object-lookup-chain-length"`
	MinDirnameMatchDepth         *int     `json:"min-dirname-match-depth"`
	MaxSelectorChainDepth        *int     `json:"max-selector-chain-depth"`
	MinSwitchChainLength         *int     `json:"min-switch-chain-length"`
	MaxIfInitOperators           *int     `json:"max-if-init-operators"`
	MaxCompositeLiteralArgDepth  *int     `json:"max-composite-literal-arg-depth"`
	MaxFunctionLines             *int     `json:"max-function-lines"`
	MaxFunctionParams            *int     `json:"max-function-params"`
	MaxNakedReturnLines          *int     `json:"max-naked-return-lines"`
	NegativeConditionNamePattern string   `json:"negative-condition-name-pattern"`
}

func (s Settings) automatedCommentIdentifiers() []string {
	if s.AutomatedCommentIdentifiers != nil {
		return s.AutomatedCommentIdentifiers
	}

	return defaultAutomatedCommentIdentifiers
}

func (s Settings) hasCommentPolicy() bool {
	hasMatchers := len(s.CommentMatchers) > 0
	if hasMatchers {
		return true
	}

	hasPrefixes := len(s.CommentPrefixIdentifiers) > 0
	if hasPrefixes {
		return true
	}

	return len(s.CommentSuffixIdentifiers) > 0
}

func (s Settings) RuleEnabled(code string, name string, defaultEnabled bool) bool {
	selected := defaultEnabled
	if len(s.EnabledRules) > 0 {
		selected = selectorMatchesAny(code, name, s.EnabledRules)
	}

	if !selected {
		return false
	}

	return !selectorMatchesAny(code, name, s.DisabledRules)
}

func (s Settings) maxExpressionOperators() int {
	return intSetting(s.MaxExpressionOperators, defaultMaxExpressionOperators)
}

func (s Settings) maxIfOperators() int {
	return intSetting(s.MaxIfOperators, defaultMaxIfOperators)
}

func (s Settings) maxControlFlowDepth() int {
	return intSetting(s.MaxControlFlowDepth, defaultMaxControlFlowDepth)
}

func (s Settings) maxArrayChainDepth() int {
	return intSetting(s.MaxArrayChainDepth, defaultMaxArrayChainDepth)
}

func (s Settings) maxComputedValueOperators() int {
	return intSetting(s.MaxComputedValueOperators, defaultMaxComputedValueOperators)
}

func (s Settings) minObjectLookupChainLength() int {
	return intSetting(s.MinObjectLookupChainLength, defaultMinObjectLookupChainLength)
}

func (s Settings) minDirnameMatchDepth() int {
	return intSetting(s.MinDirnameMatchDepth, defaultMinDirnameMatchDepth)
}

func (s Settings) maxSelectorChainDepth() int {
	return intSetting(s.MaxSelectorChainDepth, defaultMaxSelectorChainDepth)
}

func (s Settings) minSwitchChainLength() int {
	return intSetting(s.MinSwitchChainLength, defaultMinSwitchChainLength)
}

func (s Settings) maxIfInitOperators() int {
	return intSetting(s.MaxIfInitOperators, defaultMaxIfInitOperators)
}

func (s Settings) maxCompositeLiteralArgDepth() int {
	return intSetting(s.MaxCompositeLiteralArgDepth, defaultMaxCompositeLiteralArgDepth)
}

func (s Settings) maxFunctionLines() int {
	return intSetting(s.MaxFunctionLines, defaultMaxFunctionLines)
}

func (s Settings) maxFunctionParams() int {
	return intSetting(s.MaxFunctionParams, defaultMaxFunctionParams)
}

func (s Settings) maxNakedReturnLines() int {
	return intSetting(s.MaxNakedReturnLines, defaultMaxNakedReturnLines)
}

func intSetting(value *int, fallback int) int {
	if value == nil {
		return fallback
	}

	return *value
}

func selectorMatchesAny(code string, name string, selectors []string) bool {
	for _, selector := range selectors {
		if selectorMatches(code, name, selector) {
			return true
		}
	}

	return false
}

func selectorMatches(code string, name string, selector string) bool {
	normalized := strings.TrimSpace(selector)
	if normalized == "" {
		return false
	}

	if specialSelectorMatches(normalized) {
		return true
	}

	if strings.EqualFold(normalized, code) {
		return true
	}

	return strings.EqualFold(normalized, name)
}

func specialSelectorMatches(selector string) bool {
	if strings.EqualFold(selector, "all") {
		return true
	}

	return strings.EqualFold(selector, defaultNegativeConditionNamePrefix)
}
