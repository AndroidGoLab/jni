package jnigen

// MergedMethod represents a method in the idiomatic layer.
type MergedMethod struct {
	GoName         string
	Receiver       string
	Params         []GoParam     // Only non-implicit params (visible in Go signature)
	AllParams      []CapiArgInfo // All params in C order (for capi call generation)
	Returns        []GoReturn
	CapiCall       string
	CheckException bool
	Transforms     []string
	Defers         []string
}
