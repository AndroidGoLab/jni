package jnigen

// GoReturn is a return value in the idiomatic Go API.
type GoReturn struct {
	GoType    string
	IsError   bool
	Transform string
}
