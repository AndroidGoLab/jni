package jnigen

// FunctionFamily represents a typed function family that expands across types.
type FunctionFamily struct {
	Pattern   string   `yaml:"pattern"`
	Vtable    string   `yaml:"vtable"`
	Params    []string `yaml:"params"`
	Returns   string   `yaml:"returns"`
	Expand    []string `yaml:"expand"`
	Exception bool     `yaml:"exception"`
}
