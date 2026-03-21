package javagen

// IntentExtra describes a value extracted from a sticky broadcast intent.
type IntentExtra struct {
	JavaExtra string `yaml:"java_extra"`
	JavaType  string `yaml:"java_type"`
	GoName    string `yaml:"go_name"`
	GoType    string `yaml:"go_type"`
}
