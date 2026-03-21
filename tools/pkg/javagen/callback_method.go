package javagen

// CallbackMethod describes a single method in a callback interface.
type CallbackMethod struct {
	JavaMethod string   `yaml:"java_method"`
	Params     []string `yaml:"params"`
	GoField    string   `yaml:"go_field"`
}
