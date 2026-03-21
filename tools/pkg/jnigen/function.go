package jnigen

// Function represents an individual JNI function.
type Function struct {
	Name      string   `yaml:"name"`
	Params    []string `yaml:"params,omitempty"`
	Returns   string   `yaml:"returns,omitempty"`
	Exception bool     `yaml:"exception,omitempty"`
}
