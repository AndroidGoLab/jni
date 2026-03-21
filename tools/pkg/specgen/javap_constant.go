package specgen

// JavapConstant is a public static final field.
type JavapConstant struct {
	Name     string // e.g. "ERROR_BAD_VALUE"
	JavaType string // e.g. "int", "java.lang.String"
	Value    string // e.g. "1", "gps" -- extracted from ConstantValue attribute
}
