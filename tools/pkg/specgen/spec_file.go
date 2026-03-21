package specgen

// SpecFile is the YAML output structure.
type SpecFile struct {
	Package   string         `yaml:"package"`
	GoImport  string         `yaml:"go_import"`
	Classes   []SpecClass    `yaml:"classes"`
	Callbacks []SpecCallback `yaml:"callbacks,omitempty"`
	Constants []SpecConstant `yaml:"constants,omitempty"`
}
