package jnigen

// Spec is the top-level structure of spec/jni.yaml.
type Spec struct {
	Primitives       []Primitive      `yaml:"primitives"`
	ReferenceTypes   []ReferenceType  `yaml:"reference_types"`
	OpaqueTypes      []OpaqueType     `yaml:"opaque_types"`
	FunctionFamilies []FunctionFamily `yaml:"function_families"`
	EnvFunctions     []Function       `yaml:"env_functions"`
	VMFunctions      []Function       `yaml:"vm_functions"`
	Constants        []Constant       `yaml:"constants"`
}
