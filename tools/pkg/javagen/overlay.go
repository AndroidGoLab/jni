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

// ConversionOverride specifies custom type conversion expressions.
type ConversionOverride struct {
	ToGo   string `yaml:"to_go"`
	ToJava string `yaml:"to_java"`
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

	var ov Overlay
	if err := yaml.Unmarshal(data, &ov); err != nil {
		return nil, fmt.Errorf("parse overlay %s: %w", path, err)
	}

	return &ov, nil
}
