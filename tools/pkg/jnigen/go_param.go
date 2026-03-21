package jnigen

// GoParam is a parameter in the idiomatic Go API.
type GoParam struct {
	Name       string
	GoType     string
	CType      string // original C type from overlay (e.g., "jboolean*", "const char*")
	IsVariadic bool
}
