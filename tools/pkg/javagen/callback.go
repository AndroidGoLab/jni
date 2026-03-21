package javagen

// Callback describes a Java callback interface to bridge to Go.
type Callback struct {
	JavaInterface string           `yaml:"java_interface"`
	GoType        string           `yaml:"go_type"`
	Methods       []CallbackMethod `yaml:"methods"`
}
