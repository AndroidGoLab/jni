package callbackgen

import "fmt"

// primitiveBoxing maps Java primitive types to their boxing expression format.
// The %s placeholder is replaced with the parameter name.
var primitiveBoxing = map[string]string{
	"int":     "Integer.valueOf(%s)",
	"long":    "Long.valueOf(%s)",
	"boolean": "Boolean.valueOf(%s)",
	"float":   "Float.valueOf(%s)",
	"double":  "Double.valueOf(%s)",
	"byte":    "Byte.valueOf(%s)",
	"char":    "Character.valueOf(%s)",
	"short":   "Short.valueOf(%s)",
}

// IsPrimitive returns true if the Java type is a primitive that needs boxing.
func IsPrimitive(javaType string) bool {
	_, ok := primitiveBoxing[javaType]
	return ok
}

// BoxExpression returns the Java expression that boxes a primitive value.
// For object types it returns the parameter name unmodified.
func BoxExpression(javaType, paramName string) string {
	if format, ok := primitiveBoxing[javaType]; ok {
		return fmt.Sprintf(format, paramName)
	}
	return paramName
}
