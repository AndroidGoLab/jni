package javagen

import (
	"fmt"
	"strings"
)

// TypeConv holds the Go and JNI type information for a Java type.
type TypeConv struct {
	GoType     string
	JNISig     string
	CallSuffix string
	IsObject   bool
}

var primitiveTypeMap = map[string]TypeConv{
	"boolean": {GoType: "bool", JNISig: "Z", CallSuffix: "Boolean"},
	"byte":    {GoType: "int8", JNISig: "B", CallSuffix: "Byte"},
	"char":    {GoType: "uint16", JNISig: "C", CallSuffix: "Char"},
	"short":   {GoType: "int16", JNISig: "S", CallSuffix: "Short"},
	"int":     {GoType: "int32", JNISig: "I", CallSuffix: "Int"},
	"long":    {GoType: "int64", JNISig: "J", CallSuffix: "Long"},
	"float":   {GoType: "float32", JNISig: "F", CallSuffix: "Float"},
	"double":  {GoType: "float64", JNISig: "D", CallSuffix: "Double"},
	"void":    {GoType: "", JNISig: "V", CallSuffix: "Void"},
}

// javaLangShortNames maps short class names to their fully-qualified java.lang equivalents.
var javaLangShortNames = map[string]string{
	"String":  "java.lang.String",
	"Object":  "java.lang.Object",
	"Integer": "java.lang.Integer",
	"Long":    "java.lang.Long",
	"Boolean": "java.lang.Boolean",
	"Float":   "java.lang.Float",
	"Double":  "java.lang.Double",
	"Byte":    "java.lang.Byte",
	"Short":   "java.lang.Short",
	"Class":   "java.lang.Class",
}

// ResolveType maps a Java type string to its TypeConv.
// Generic type parameters are stripped because JNI uses erased types.
func ResolveType(javaType string) TypeConv {
	// Handle pre-formatted JNI array syntax like "[Landroid.foo.Bar;"
	if strings.HasPrefix(javaType, "[") && !strings.HasSuffix(javaType, "[]") {
		return TypeConv{
			GoType:     "*jni.Object",
			JNISig:     normalizeJNISig(javaType),
			CallSuffix: "Object",
			IsObject:   true,
		}
	}

	// Check array types first.
	// JNI arrays are JNI objects, so the Go type is always *jni.Object.
	// Callers can convert between Go slices and JNI arrays as needed.
	if strings.HasSuffix(javaType, "[]") {
		elemType := strings.TrimSuffix(javaType, "[]")
		elemConv := ResolveType(elemType)
		return TypeConv{
			GoType:     "*jni.Object",
			JNISig:     "[" + elemConv.JNISig,
			CallSuffix: "Object",
			IsObject:   true,
		}
	}

	// Strip generic type parameters (JNI uses erased types).
	javaType = stripGenerics(javaType)

	// Check primitives.
	if tc, ok := primitiveTypeMap[javaType]; ok {
		return tc
	}

	// Check String specifically.
	if javaType == "String" || javaType == "java.lang.String" {
		return TypeConv{
			GoType:     "string",
			JNISig:     "Ljava/lang/String;",
			CallSuffix: "Object",
			IsObject:   true,
		}
	}

	// Expand short java.lang names.
	if full, ok := javaLangShortNames[javaType]; ok {
		javaType = full
	}

	// Fully-qualified object type.
	if strings.Contains(javaType, ".") {
		return TypeConv{
			GoType:     "*jni.Object",
			JNISig:     "L" + JavaClassToSlash(javaType) + ";",
			CallSuffix: "Object",
			IsObject:   true,
		}
	}

	// Unknown short name: treat as java.lang class.
	return TypeConv{
		GoType:     "*jni.Object",
		JNISig:     "Ljava/lang/" + javaType + ";",
		CallSuffix: "Object",
		IsObject:   true,
	}
}

// JNISignature computes the JNI method signature from parameter types and return type.
func JNISignature(params []Param, returnType string) string {
	var sb strings.Builder
	sb.WriteByte('(')
	for _, p := range params {
		sb.WriteString(JNITypeSignature(p.JavaType))
	}
	sb.WriteByte(')')
	sb.WriteString(JNITypeSignature(returnType))
	return sb.String()
}

// JavaClassToSlash converts a dotted Java class name to slash-separated JNI format.
// Handles inner classes: "android.app.AlarmManager.AlarmClockInfo" becomes
// "android/app/AlarmManager$AlarmClockInfo".
func JavaClassToSlash(className string) string {
	// If the name already contains $, it's already using JNI inner class notation.
	if strings.Contains(className, "$") {
		return strings.ReplaceAll(className, ".", "/")
	}

	parts := strings.Split(className, ".")
	// Find the boundary: package parts are lowercase, class parts start with uppercase.
	// The first uppercase part is the outer class; subsequent uppercase parts are inner classes.
	classStart := -1
	for i, p := range parts {
		if len(p) > 0 && p[0] >= 'A' && p[0] <= 'Z' {
			classStart = i
			break
		}
	}
	if classStart < 0 {
		// No uppercase parts found; treat as all-package.
		return strings.ReplaceAll(className, ".", "/")
	}

	// Package parts use /, outer class uses /, inner classes use $.
	var sb strings.Builder
	for i, p := range parts {
		if i > 0 {
			if i > classStart {
				sb.WriteByte('$')
			} else {
				sb.WriteByte('/')
			}
		}
		sb.WriteString(p)
	}
	return sb.String()
}

// JNITypeSignature converts a single Java type to its JNI type signature.
// Generic type parameters are stripped because JNI uses erased types.
func JNITypeSignature(javaType string) string {
	// Handle pre-formatted JNI array syntax like "[Landroid.foo.Bar;"
	if strings.HasPrefix(javaType, "[") && !strings.HasSuffix(javaType, "[]") {
		return normalizeJNISig(javaType)
	}

	// Array types.
	if strings.HasSuffix(javaType, "[]") {
		return "[" + JNITypeSignature(strings.TrimSuffix(javaType, "[]"))
	}

	// Strip generic type parameters (JNI uses erased types).
	javaType = stripGenerics(javaType)

	// Primitives.
	if tc, ok := primitiveTypeMap[javaType]; ok {
		return tc.JNISig
	}

	// String shorthand.
	if javaType == "String" {
		return "Ljava/lang/String;"
	}

	// Known java.lang short names.
	if full, ok := javaLangShortNames[javaType]; ok {
		return "L" + JavaClassToSlash(full) + ";"
	}

	// Fully-qualified class.
	if strings.Contains(javaType, ".") {
		return "L" + JavaClassToSlash(javaType) + ";"
	}

	// Unknown short name: assume java.lang.
	return "Ljava/lang/" + javaType + ";"
}

// ParamConversionCode generates Go code to convert a Go parameter to its JNI
// representation before a JNI call.
func ParamConversionCode(p MergedParam) string {
	varName := "j" + strings.Title(p.GoName) //nolint:staticcheck // strings.Title is fine here
	if p.IsString {
		return fmt.Sprintf("%s, err := env.NewStringUTF(%s)\n\t\tif err != nil {\n\t\t\treturn %s\n\t\t}\n",
			varName, p.GoName, "err")
	}
	if p.IsBool {
		return fmt.Sprintf("var %s uint8\n\t\tif %s {\n\t\t\t%s = 1\n\t\t}\n",
			varName, p.GoName, varName)
	}
	return ""
}

// ReturnConversionCode generates Go code to convert a JNI return value to Go.
func ReturnConversionCode(javaType, goType string) string {
	tc := ResolveType(javaType)
	switch {
	case javaType == "void":
		return ""
	case javaType == "boolean":
		return "result != 0"
	case javaType == "String" || javaType == "java.lang.String":
		return "env.GoString((*jni.String)(unsafe.Pointer(resultObj)))"
	case tc.IsObject:
		return "resultObj"
	default:
		return fmt.Sprintf("%s(result)", goType)
	}
}

// normalizeJNISig normalizes a pre-formatted JNI signature that may contain dots
// instead of slashes. For example, "[Landroid.foo.Bar;" becomes "[Landroid/foo/Bar;".
func normalizeJNISig(sig string) string {
	return strings.ReplaceAll(sig, ".", "/")
}

// stripGenerics removes Java generic type parameters from a type name.
// JNI uses erased types, so generics must be stripped before building
// JNI signatures. Handles nested generics:
//
//	"java.util.Set<java.lang.String>"                → "java.util.Set"
//	"java.util.Map<java.lang.String, java.util.List<java.lang.Integer>>" → "java.util.Map"
func stripGenerics(javaType string) string {
	idx := strings.IndexByte(javaType, '<')
	if idx < 0 {
		return javaType
	}
	// Everything before the first '<' is the raw type.
	return javaType[:idx]
}

// jniValueFunc returns the jni.XxxValue() function name for a given JNI call suffix.
func jniValueFunc(callSuffix string) string {
	switch callSuffix {
	case "Boolean":
		return "BooleanValue"
	case "Byte":
		return "ByteValue"
	case "Char":
		return "CharValue"
	case "Short":
		return "ShortValue"
	case "Int":
		return "IntValue"
	case "Long":
		return "LongValue"
	case "Float":
		return "FloatValue"
	case "Double":
		return "DoubleValue"
	case "Object":
		return "ObjectValue"
	default:
		return "IntValue"
	}
}
