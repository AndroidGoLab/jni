package javagen

// ReturnKind classifies how a method's return value should be handled
// in the generated code.
type ReturnKind string

const (
	ReturnVoid      ReturnKind = "void"
	ReturnString    ReturnKind = "string"
	ReturnBool      ReturnKind = "bool"
	ReturnObject    ReturnKind = "object"
	ReturnPrimitive ReturnKind = "primitive"
)
