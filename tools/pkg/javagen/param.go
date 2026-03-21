package javagen

// Param describes a method parameter.
type Param struct {
	JavaType string `yaml:"java_type"`
	GoName   string `yaml:"go_name"`
	GoType   string `yaml:"go_type"`
}
