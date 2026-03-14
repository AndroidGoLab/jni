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

// FuncOverlay holds per-function overlay data.
type FuncOverlay struct {
	GoName         string         `yaml:"go_name"`
	CheckException bool           `yaml:"check_exception"`
	Params         []ParamOverlay `yaml:"params,omitempty"`
	Returns        *ReturnOverlay `yaml:"returns,omitempty"`
	Skip           bool           `yaml:"skip,omitempty"`
}

// ParamOverlay defines a Go-friendly parameter override.
type ParamOverlay struct {
	Name      string `yaml:"name"`
	CType     string `yaml:"c_type"`
	GoType    string `yaml:"go_type"`
	Transform string `yaml:"transform,omitempty"`
	Implicit  string `yaml:"implicit,omitempty"`
}

// ReturnOverlay defines a Go-friendly return type override.
type ReturnOverlay struct {
	GoType    string `yaml:"go_type"`
	Transform string `yaml:"transform,omitempty"`
}

// FamilyOverlay defines overlay for a typed function family.
type FamilyOverlay struct {
	GoPattern      string         `yaml:"go_pattern"`
	CheckException bool           `yaml:"check_exception"`
	Params         []ParamOverlay `yaml:"params"`
	// Returns can be a map (keyed by "primitive", "Object", "Void") or a string.
	// We use a custom unmarshaler.
	Returns interface{} `yaml:"returns,omitempty"`
}

// FamilyReturnMap returns the returns field as a map if it is one, or nil.
func (fo *FamilyOverlay) FamilyReturnMap() map[string]string {
	if fo.Returns == nil {
		return nil
	}
	switch v := fo.Returns.(type) {
	case map[string]interface{}:
		result := make(map[string]string, len(v))
		for k, val := range v {
			if val == nil {
				result[k] = ""
			} else {
				result[k] = fmt.Sprintf("%v", val)
			}
		}
		return result
	case string:
		// A single string means all types use this return type.
		return map[string]string{"_all": v}
	}
	return nil
}

// ParamTransform defines how to convert between Go and C types.
type ParamTransform struct {
	Description string `yaml:"description"`
	GoType      string `yaml:"go_type"`
	CType       string `yaml:"c_type"`
	ToC         string `yaml:"to_c,omitempty"`
	CArg        string `yaml:"c_arg,omitempty"`
	FromC       string `yaml:"from_c,omitempty"`
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
