package specgen

// SpecMethod is a method in the YAML spec.
type SpecMethod struct {
	JavaMethod string      `yaml:"java_method"`
	GoName     string      `yaml:"go_name"`
	Params     []SpecParam `yaml:"params,omitempty"`
	Returns    string      `yaml:"returns"`
	Error      bool        `yaml:"error"`
}
