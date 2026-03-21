package jnigen

// MergedConstant is a constant with Go name and type.
type MergedConstant struct {
	CName  string
	GoName string
	Value  string
	GoType string
}
