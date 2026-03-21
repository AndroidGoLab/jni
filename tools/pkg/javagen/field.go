package javagen

import "strings"

// Field describes a getter method on a data class.
// Supports both java_method (getter) and java_field (direct field access).
type Field struct {
	JavaMethod string `yaml:"java_method"`
	JavaField  string `yaml:"java_field"`
	Returns    string `yaml:"returns"`
	GoName     string `yaml:"go_name"`
	GoType     string `yaml:"go_type"`
}

// GetterName returns the Java getter method name for this field.
// If java_method is set, it is used directly. If java_field is set,
// it is converted to a getter name (e.g., "packageName" -> "getPackageName").
func (f Field) GetterName() string {
	if f.JavaMethod != "" {
		return f.JavaMethod
	}
	if f.JavaField != "" {
		if len(f.JavaField) == 0 {
			return ""
		}
		return "get" + strings.ToUpper(f.JavaField[:1]) + f.JavaField[1:]
	}
	return ""
}
