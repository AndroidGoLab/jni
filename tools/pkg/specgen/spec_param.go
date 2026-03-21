package specgen

// SpecParam is a method parameter in the YAML spec.
type SpecParam struct {
	JavaType string `yaml:"java_type"`
	GoName   string `yaml:"go_name"`
}
