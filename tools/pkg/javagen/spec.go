package javagen

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Spec represents a per-package Java API specification loaded from YAML.
type Spec struct {
	Package   string     `yaml:"package"`
	GoImport  string     `yaml:"go_import"`
	Classes   []Class    `yaml:"classes"`
	Callbacks []Callback `yaml:"callbacks"`
	Constants []Constant `yaml:"constants"`
}

// Class describes a Java class to wrap.
type Class struct {
	JavaClass   string   `yaml:"java_class"`
	GoType      string   `yaml:"go_type"`
	Obtain      string   `yaml:"obtain"`
	ServiceName string   `yaml:"service_name"`
	Kind        string   `yaml:"kind"`
	Close       bool     `yaml:"close"`
	Methods     []Method `yaml:"methods"`
	Fields      []Field  `yaml:"fields"`
}

// Method describes a Java method to generate a Go wrapper for.
type Method struct {
	JavaMethod string  `yaml:"java_method"`
	GoName     string  `yaml:"go_name"`
	Static     bool    `yaml:"static"`
	Params     []Param `yaml:"params"`
	Returns    string  `yaml:"returns"`
	Error      bool    `yaml:"error"`
}

// Param describes a method parameter.
type Param struct {
	JavaType string `yaml:"java_type"`
	GoName   string `yaml:"go_name"`
	GoType   string `yaml:"go_type"`
}

// Field describes a getter method on a data class.
// Supports both java_method (getter) and java_field (direct field access).
type Field struct {
	JavaMethod string `yaml:"java_method"`
	JavaField  string `yaml:"java_field"`
	Returns    string `yaml:"returns"`
	GoName     string `yaml:"go_name"`
	GoType     string `yaml:"go_type"`
}

// GetterName returns the Java getter method name for this field.
// If java_method is set, it is used directly. If java_field is set,
// it is converted to a getter name (e.g., "packageName" → "getPackageName").
func (f Field) GetterName() string {
	if f.JavaMethod != "" {
		return f.JavaMethod
	}
	if f.JavaField != "" {
		if len(f.JavaField) == 0 {
			return ""
		}
		return "get" + strings.ToUpper(f.JavaField[:1]) + f.JavaField[1:]
	}
	return ""
}

// Callback describes a Java callback interface to bridge to Go.
type Callback struct {
	JavaInterface string           `yaml:"java_interface"`
	GoType        string           `yaml:"go_type"`
	Methods       []CallbackMethod `yaml:"methods"`
}

// CallbackMethod describes a single method in a callback interface.
type CallbackMethod struct {
	JavaMethod string   `yaml:"java_method"`
	Params     []string `yaml:"params"`
	GoField    string   `yaml:"go_field"`
}

// Constant describes a named constant value.
type Constant struct {
	GoName       string `yaml:"go_name"`
	Value        string `yaml:"value"`
	GoType       string `yaml:"go_type"`
	GoUnderlying string `yaml:"go_underlying"`
}

var validPkgName = regexp.MustCompile(`^[a-z][a-z0-9_]*(/[a-z][a-z0-9_]*)*$`)

// LoadSpec reads and validates a Java API spec from a YAML file.
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	if err := validateSpec(&spec); err != nil {
		return nil, fmt.Errorf("validate %s: %w", path, err)
	}

	return &spec, nil
}

func validateSpec(spec *Spec) error {
	if !validPkgName.MatchString(spec.Package) {
		return fmt.Errorf("invalid package name: %q", spec.Package)
	}
	if spec.GoImport == "" {
		return fmt.Errorf("go_import is required")
	}
	for i, cls := range spec.Classes {
		if cls.JavaClass == "" {
			return fmt.Errorf("classes[%d]: java_class is required", i)
		}
		if cls.GoType == "" {
			return fmt.Errorf("classes[%d] (%s): go_type is required", i, cls.JavaClass)
		}
		if cls.Obtain == "system_service" && cls.ServiceName == "" {
			return fmt.Errorf("classes[%d] (%s): service_name is required when obtain is system_service", i, cls.JavaClass)
		}
		for j, m := range cls.Methods {
			if m.JavaMethod == "" {
				return fmt.Errorf("classes[%d].methods[%d]: java_method is required", i, j)
			}
			if m.GoName == "" {
				return fmt.Errorf("classes[%d].methods[%d] (%s): go_name is required", i, j, m.JavaMethod)
			}
			for k, p := range m.Params {
				if p.JavaType == "" {
					return fmt.Errorf("classes[%d].methods[%d].params[%d]: java_type is required", i, j, k)
				}
				if p.GoName == "" {
					return fmt.Errorf("classes[%d].methods[%d].params[%d]: go_name is required", i, j, k)
				}
			}
		}
	}
	return nil
}
