package javagen

import "strings"

// MergedAbstractCallbackMethod is a resolved method in an abstract callback class.
type MergedAbstractCallbackMethod struct {
	JavaMethod string
	GoField    string
	Params     []MergedParam
	GoParams   string
	Returns    string
}

// JavaReturnType returns the Java return type for use in the adapter template.
func (m *MergedAbstractCallbackMethod) JavaReturnType() string {
	if m.Returns == "" || m.Returns == "void" {
		return "void"
	}
	return m.Returns
}

// JavaParamList returns the Java parameter declaration list for the adapter method
// (e.g. "int arg0, android.bluetooth.le.ScanResult arg1").
func (m *MergedAbstractCallbackMethod) JavaParamList() string {
	var parts []string
	for _, p := range m.Params {
		parts = append(parts, p.JavaType+" "+p.GoName)
	}
	return strings.Join(parts, ", ")
}

// JavaArgList returns the argument names for passing to GoAbstractDispatch.invoke
// (e.g. "arg0, arg1"). Primitives are autoboxed via valueOf wrappers.
func (m *MergedAbstractCallbackMethod) JavaArgList() string {
	var parts []string
	for _, p := range m.Params {
		parts = append(parts, javaAutoboxExpression(p.JavaType, p.GoName))
	}
	return strings.Join(parts, ", ")
}

// HasReturn reports whether this method returns a value (non-void).
func (m *MergedAbstractCallbackMethod) HasReturn() bool {
	return m.Returns != "" && m.Returns != "void"
}

// JavaCastReturn returns the Java cast expression needed to convert
// the Object return from GoAbstractDispatch.invoke to the method's return type.
func (m *MergedAbstractCallbackMethod) JavaCastReturn() string {
	switch m.Returns {
	case "int":
		return "(Integer)"
	case "long":
		return "(Long)"
	case "boolean":
		return "(Boolean)"
	case "float":
		return "(Float)"
	case "double":
		return "(Double)"
	case "byte":
		return "(Byte)"
	case "short":
		return "(Short)"
	case "char":
		return "(Character)"
	default:
		return "(" + m.Returns + ")"
	}
}

// JavaUnboxReturn returns the Java unboxing method call for primitive return types,
// or empty string for object/void returns.
func (m *MergedAbstractCallbackMethod) JavaUnboxReturn() string {
	switch m.Returns {
	case "int":
		return ".intValue()"
	case "long":
		return ".longValue()"
	case "boolean":
		return ".booleanValue()"
	case "float":
		return ".floatValue()"
	case "double":
		return ".doubleValue()"
	case "byte":
		return ".byteValue()"
	case "short":
		return ".shortValue()"
	case "char":
		return ".charValue()"
	default:
		return ""
	}
}

// javaAutoboxExpression returns a Java expression that autoboxes a primitive
// value for inclusion in an Object[] array. Object types pass through unchanged.
func javaAutoboxExpression(javaType, varName string) string {
	switch javaType {
	case "int":
		return "Integer.valueOf(" + varName + ")"
	case "long":
		return "Long.valueOf(" + varName + ")"
	case "boolean":
		return "Boolean.valueOf(" + varName + ")"
	case "float":
		return "Float.valueOf(" + varName + ")"
	case "double":
		return "Double.valueOf(" + varName + ")"
	case "byte":
		return "Byte.valueOf(" + varName + ")"
	case "short":
		return "Short.valueOf(" + varName + ")"
	case "char":
		return "Character.valueOf(" + varName + ")"
	default:
		return varName
	}
}
