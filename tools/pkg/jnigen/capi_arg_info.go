package jnigen

// CapiArgInfo holds information needed to generate a single capi call argument.
type CapiArgInfo struct {
	Name       string // param name (for explicit params)
	GoType     string // Go type
	CType      string // original C type from overlay (e.g., "jboolean*", "const char*")
	IsVariadic bool   // whether this is a variadic param
	IsImplicit bool   // true if this is an implicit param
	Implicit   string // Go expression for implicit params (e.g., "len(buf)")
}
