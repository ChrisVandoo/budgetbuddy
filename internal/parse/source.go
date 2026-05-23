package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

func LoadSources(path string) (*types.SourcesYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.SourcesYAML{Sources: make(map[string]types.SourceConfig)}, nil
		}
		return nil, fmt.Errorf("read sources file: %w", err)
	}

	var sources types.SourcesYAML
	if err := yaml.Unmarshal(data, &sources); err != nil {
		return nil, fmt.Errorf("unmarshal sources: %w", err)
	}

	if sources.Sources == nil {
		sources.Sources = make(map[string]types.SourceConfig)
	}

	return &sources, nil
}

func SaveSources(path string, sources *types.SourcesYAML) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create sources dir: %w", err)
	}

	data, err := yaml.Marshal(sources)
	if err != nil {
		return fmt.Errorf("marshal sources: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write sources file: %w", err)
	}

	return nil
}

func DetectSource(headers []string, sources *types.SourcesYAML) (string, *types.SourceConfig, bool) {
	headerStr := strings.Join(headers, ",")

	for key, config := range sources.Sources {
		if normalizeKey(key) == normalizeKey(headerStr) {
			return key, &config, true
		}
	}

	return "", nil, false
}

func normalizeKey(key string) string {
	return strings.TrimSpace(strings.ToLower(key))
}

func NormalizeAmount(csvValue string, mapping types.AmountMapping) (int64, error) {
	if mapping.SingleColumn {
		val, err := ParseCents(csvValue)
		if err != nil {
			return 0, err
		}
		if mapping.IsPositiveMoneyIn {
			return val, nil
		}
		return -val, nil
	}

	return 0, fmt.Errorf("dual column amount not supported directly")
}

func ParseCents(val string) (int64, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return 0, nil
	}

	var negative bool
	if val[0] == '-' {
		negative = true
		val = val[1:]
	} else if val[0] == '+' {
		val = val[1:]
	}

	val = strings.ReplaceAll(val, ",", "")
	val = strings.ReplaceAll(val, "$", "")
	val = strings.TrimSpace(val)

	parts := strings.Split(val, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid number: %s", val)
	}

	dollars := parts[0]
	if dollars == "" {
		dollars = "0"
	}
	for _, c := range dollars {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid number: %s", val)
		}
	}

	var cents int64
	for _, c := range dollars {
		cents = cents*10 + int64(c-'0')
	}
	cents *= 100

	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) > 2 {
			frac = frac[:2]
		}
		for _, c := range frac {
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("invalid number: %s", val)
			}
		}
		var fracCents int64
		for _, c := range frac {
			fracCents = fracCents*10 + int64(c-'0')
		}
		for i := len(frac); i < 2; i++ {
			fracCents *= 10
		}
		cents += fracCents
	}

	if negative {
		cents = -cents
	}
	return cents, nil
}
