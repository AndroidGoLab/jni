package specgen

// SpecClass is a class in the YAML spec.
type SpecClass struct {
	JavaClass     string       `yaml:"java_class"`
	GoType        string       `yaml:"go_type"`
	Obtain        string       `yaml:"obtain,omitempty"`
	ServiceName   string       `yaml:"service_name,omitempty"`
	Kind          string       `yaml:"kind,omitempty"`
	Close         bool         `yaml:"close,omitempty"`
	Methods       []SpecMethod `yaml:"methods,omitempty"`
	StaticMethods []SpecMethod `yaml:"static_methods,omitempty"`
	Fields        []SpecField  `yaml:"fields,omitempty"`
}
