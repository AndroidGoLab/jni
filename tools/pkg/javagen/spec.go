package javagen

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

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
		for j, m := range cls.AllMethods() {
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
