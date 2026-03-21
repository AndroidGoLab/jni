package javagen

// Constant describes a named constant value.
type Constant struct {
	GoName       string `yaml:"go_name"`
	Value        string `yaml:"value"`
	GoType       string `yaml:"go_type"`
	GoUnderlying string `yaml:"go_underlying"`
}
