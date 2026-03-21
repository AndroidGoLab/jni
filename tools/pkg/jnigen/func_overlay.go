package jnigen

// FuncOverlay holds per-function overlay data.
type FuncOverlay struct {
	GoName         string         `yaml:"go_name"`
	CheckException bool           `yaml:"check_exception"`
	Params         []ParamOverlay `yaml:"params,omitempty"`
	Returns        *ReturnOverlay `yaml:"returns,omitempty"`
	Skip           bool           `yaml:"skip,omitempty"`
}
