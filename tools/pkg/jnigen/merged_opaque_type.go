package jnigen

// MergedOpaqueType is an opaque type with Go name.
type MergedOpaqueType struct {
	CType    string
	GoType   string
	CapiType string // CamelCase of CType, for use in capi layer
}
