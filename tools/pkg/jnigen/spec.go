package jnigen

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

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
