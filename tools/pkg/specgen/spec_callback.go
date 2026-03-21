package specgen

// SpecCallback is a callback interface in the YAML spec.
type SpecCallback struct {
	JavaInterface string               `yaml:"java_interface"`
	GoType        string               `yaml:"go_type"`
	Methods       []SpecCallbackMethod `yaml:"methods"`
}
