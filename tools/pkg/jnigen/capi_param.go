package jnigen

// CapiParam is a parameter in the capi layer.
type CapiParam struct {
	CType      string
	CName      string
	VtableCast string // if set, cast to this type in the vtable call
}
