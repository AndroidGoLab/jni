package specgen

// SpecAbstractCallbackMethod is a method in an abstract callback class.
type SpecAbstractCallbackMethod struct {
	JavaMethod string   `yaml:"java_method"`
	Params     []string `yaml:"params"`
	Returns    string   `yaml:"returns"`
	GoField    string   `yaml:"go_field"`
}
