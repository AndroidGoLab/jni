package javagen

// MergedField is a resolved field getter on a data class.
type MergedField struct {
	JavaMethod string // getter method name (for java_method fields)
	JavaName   string // raw field name (for java_field direct access)
	GoName     string
	GoType     string
	JNISig     string
	CallSuffix string
	IsField    bool // true = use GetFieldID/Get*Field, false = use GetMethodID/Call*Method
}
