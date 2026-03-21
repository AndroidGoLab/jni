package javagen

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Overlay provides per-package customization applied on top of a Spec.
type Overlay struct {
	GoNameOverrides map[string]string             `yaml:"go_name_overrides"`
	MethodGrouping  map[string][]string           `yaml:"method_grouping"`
	ExtraMethods    []Method                      `yaml:"extra_methods"`
	TypeOverrides   map[string]string             `yaml:"type_overrides"`
	ConversionFuncs map[string]ConversionOverride `yaml:"conversion_funcs"`
}

// LoadOverlay reads an overlay YAML file. If the file does not exist, an
// empty Overlay is returned without error.
func LoadOverlay(path string) (*Overlay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Overlay{}, nil
		}
		return nil, fmt.Errorf("read overlay %s: %w", path, err)
	}
	return ParseOverlay(data)
}

// ParseOverlay parses an overlay from YAML data.
func ParseOverlay(data []byte) (*Overlay, error) {
	var ov Overlay
	if err := yaml.Unmarshal(data, &ov); err != nil {
		return nil, fmt.Errorf("parse overlay: %w", err)
	}
	return &ov, nil
}
