package specgen

// SpecConstant is a constant in the YAML spec.
type SpecConstant struct {
	GoName string `yaml:"go_name"`
	Value  string `yaml:"value"`
	GoType string `yaml:"go_type"`
}
