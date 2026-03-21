package specgen

// SpecCallbackMethod is a callback method in the YAML spec.
type SpecCallbackMethod struct {
	JavaMethod string   `yaml:"java_method"`
	Params     []string `yaml:"params"`
	GoField    string   `yaml:"go_field"`
}
