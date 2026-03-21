package specgen

// SpecField is a data class field in the YAML spec.
type SpecField struct {
	JavaMethod string `yaml:"java_method"`
	Returns    string `yaml:"returns"`
	GoName     string `yaml:"go_name"`
	GoType     string `yaml:"go_type"`
}
