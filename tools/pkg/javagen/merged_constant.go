package javagen

// MergedConstant is a resolved constant value.
type MergedConstant struct {
	GoName string
	Value  string
	// IsVar is true when the value cannot be a Go compile-time constant
	// (e.g. NaN, Infinity) and must be declared as a var instead.
	IsVar bool
}
