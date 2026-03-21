package javagen

// MergedConstantGroup groups constants by their Go type.
type MergedConstantGroup struct {
	GoType   string
	BaseType string
	Values   []MergedConstant
}
