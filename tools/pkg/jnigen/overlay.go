package jnigen

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Overlay is the top-level structure of spec/overlays/jni.yaml.
type Overlay struct {
	TypeRenames     map[string]string         `yaml:"type_renames"`
	Receivers       map[string]string         `yaml:"receivers"`
	Functions       map[string]FuncOverlay    `yaml:"functions"`
	FamilyOverlays  map[string]FamilyOverlay  `yaml:"family_overlays"`
	ParamTransforms map[string]ParamTransform `yaml:"param_transforms"`
}

// LoadOverlay reads and parses an overlay YAML file.
func LoadOverlay(path string) (*Overlay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var overlay Overlay
	if err := yaml.Unmarshal(data, &overlay); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if overlay.TypeRenames == nil {
		overlay.TypeRenames = make(map[string]string)
	}
	if overlay.Receivers == nil {
		overlay.Receivers = make(map[string]string)
	}
	if overlay.Functions == nil {
		overlay.Functions = make(map[string]FuncOverlay)
	}
	if overlay.FamilyOverlays == nil {
		overlay.FamilyOverlays = make(map[string]FamilyOverlay)
	}
	if overlay.ParamTransforms == nil {
		overlay.ParamTransforms = make(map[string]ParamTransform)
	}

	return &overlay, nil
}
