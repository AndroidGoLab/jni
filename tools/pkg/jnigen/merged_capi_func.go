package jnigen

// MergedCapiFunc represents a function in the capi layer.
type MergedCapiFunc struct {
	CName      string
	HelperName string
	Vtable     string
	Params     []CapiParam
	Returns    string
	IsVoid     bool
}
