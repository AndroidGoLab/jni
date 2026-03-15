package callbackgen

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Spec is the top-level structure of the callbacks YAML spec.
type Spec struct {
	Callbacks []CallbackEntry `yaml:"callbacks"`
}

// CallbackEntry describes a single abstract class or interface
// that needs a generated Java adapter.
type CallbackEntry struct {
	// Class is the fully-qualified Java class or interface name.
	Class string `yaml:"class"`

	// Adapter is the generated adapter class name (unqualified).
	Adapter string `yaml:"adapter"`

	// Interface indicates whether the base type is a Java interface
	// (uses "implements") rather than a class (uses "extends").
	Interface bool `yaml:"interface"`

	// Methods lists the abstract methods to override.
	Methods []MethodEntry `yaml:"methods"`
}

// MethodEntry describes a single abstract method to override.
type MethodEntry struct {
	Name   string       `yaml:"name"`
	Params []ParamEntry `yaml:"params"`
}

// ParamEntry describes a single method parameter.
type ParamEntry struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}

// LoadSpec reads and parses a callbacks YAML spec file.
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &spec, nil
}
