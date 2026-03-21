package javagen

// StaticField describes a Java static field to read.
type StaticField struct {
	JavaField string `yaml:"java_field"`
	Returns   string `yaml:"returns"`
	GoName    string `yaml:"go_name"`
	GoType    string `yaml:"go_type"`
}
