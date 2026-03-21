package javagen

// Class describes a Java class to wrap.
type Class struct {
	JavaClass     string        `yaml:"java_class"`
	GoType        string        `yaml:"go_type"`
	Obtain        string        `yaml:"obtain"`
	ServiceName   string        `yaml:"service_name"`
	Kind          string        `yaml:"kind"`
	Close         bool          `yaml:"close"`
	Methods       []Method      `yaml:"methods"`
	StaticMethods []Method      `yaml:"static_methods"`
	Fields        []Field       `yaml:"fields"`
	StaticFields  []StaticField `yaml:"static_fields"`
	IntentExtras  []IntentExtra `yaml:"intent_extras"`
}

// AllMethods returns both instance and static methods.
// Static methods have their Static field set to true.
func (c Class) AllMethods() []Method {
	all := make([]Method, 0, len(c.Methods)+len(c.StaticMethods))
	all = append(all, c.Methods...)
	for _, m := range c.StaticMethods {
		m.Static = true
		all = append(all, m)
	}
	return all
}
