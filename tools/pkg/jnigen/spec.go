package jnigen

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Spec is the top-level structure of spec/jni.yaml.
type Spec struct {
	Primitives       []Primitive      `yaml:"primitives"`
	ReferenceTypes   []ReferenceType  `yaml:"reference_types"`
	OpaqueTypes      []OpaqueType     `yaml:"opaque_types"`
	FunctionFamilies []FunctionFamily `yaml:"function_families"`
	EnvFunctions     []Function       `yaml:"env_functions"`
	VMFunctions      []Function       `yaml:"vm_functions"`
	Constants        []Constant       `yaml:"constants"`
}

// Primitive represents a JNI primitive type.
type Primitive struct {
	CType  string `yaml:"c_type"`
	GoType string `yaml:"go_type"`
	Suffix string `yaml:"suffix"`
}

// ReferenceType represents a JNI reference type in the hierarchy.
type ReferenceType struct {
	CType  string  `yaml:"c_type"`
	Parent *string `yaml:"parent"`
}

// OpaqueType represents an opaque JNI type (method/field IDs).
type OpaqueType struct {
	CType  string `yaml:"c_type"`
	GoType string `yaml:"go_type,omitempty"`
}

// FunctionFamily represents a typed function family that expands across types.
type FunctionFamily struct {
	Pattern   string   `yaml:"pattern"`
	Vtable    string   `yaml:"vtable"`
	Params    []string `yaml:"params"`
	Returns   string   `yaml:"returns"`
	Expand    []string `yaml:"expand"`
	Exception bool     `yaml:"exception"`
}

// Function represents an individual JNI function.
type Function struct {
	Name      string   `yaml:"name"`
	Params    []string `yaml:"params,omitempty"`
	Returns   string   `yaml:"returns,omitempty"`
	Exception bool     `yaml:"exception,omitempty"`
}

// Constant represents a JNI constant.
type Constant struct {
	Name   string `yaml:"name"`
	Value  string `yaml:"value"`
	GoType string `yaml:"go_type"`
}

// LoadSpec reads and parses a spec YAML file.
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if err := validateSpec(&spec); err != nil {
		return nil, fmt.Errorf("validating %s: %w", path, err)
	}

	return &spec, nil
}

func validateSpec(spec *Spec) error {
	if len(spec.Primitives) == 0 {
		return fmt.Errorf("no primitives defined")
	}
	for i, p := range spec.Primitives {
		if p.CType == "" {
			return fmt.Errorf("primitive[%d]: missing c_type", i)
		}
		if p.GoType == "" {
			return fmt.Errorf("primitive[%d] (%s): missing go_type", i, p.CType)
		}
		if p.Suffix == "" {
			return fmt.Errorf("primitive[%d] (%s): missing suffix", i, p.CType)
		}
	}
	for i, rt := range spec.ReferenceTypes {
		if rt.CType == "" {
			return fmt.Errorf("reference_types[%d]: missing c_type", i)
		}
	}
	for i, f := range spec.FunctionFamilies {
		if f.Pattern == "" {
			return fmt.Errorf("function_families[%d]: missing pattern", i)
		}
		if f.Vtable == "" {
			return fmt.Errorf("function_families[%d] (%s): missing vtable", i, f.Pattern)
		}
		if len(f.Expand) == 0 {
			return fmt.Errorf("function_families[%d] (%s): missing expand list", i, f.Pattern)
		}
	}
	for i, f := range spec.EnvFunctions {
		if f.Name == "" {
			return fmt.Errorf("env_functions[%d]: missing name", i)
		}
	}
	for i, f := range spec.VMFunctions {
		if f.Name == "" {
			return fmt.Errorf("vm_functions[%d]: missing name", i)
		}
	}
	for i, c := range spec.Constants {
		if c.Name == "" {
			return fmt.Errorf("constants[%d]: missing name", i)
		}
	}
	return nil
}
