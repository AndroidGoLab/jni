package javagen

// AbstractCallbackMethod describes a single abstract method in an abstract callback class.
type AbstractCallbackMethod struct {
	JavaMethod string   `yaml:"java_method"`
	Params     []string `yaml:"params"`
	Returns    string   `yaml:"returns"`
	GoField    string   `yaml:"go_field"`
}
