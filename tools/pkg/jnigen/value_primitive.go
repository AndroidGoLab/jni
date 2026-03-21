package jnigen

// ValuePrimitive holds data for a single typed Value constructor.
type ValuePrimitive struct {
	GoName  string // e.g., "Int"
	GoType  string // e.g., "int32"
	CGoType string // e.g., "capi.Jint" -- the capi alias for the C type
}
