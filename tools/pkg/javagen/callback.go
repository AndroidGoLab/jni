package javagen

import (
	"fmt"
	"strings"
)

// callbackArgConversion returns Go code that extracts and converts a callback
// argument from the args slice at the given index.
func callbackArgConversion(p MergedParam, idx int) string {
	if p.IsString {
		return fmt.Sprintf("%s := env.GoString((*jni.String)(unsafe.Pointer(args[%d])))", p.GoName, idx)
	}
	// All non-string args arrive as autoboxed *jni.Object via the JNI proxy.
	return fmt.Sprintf("%s := args[%d]", p.GoName, idx)
}

// BuildCallbackDispatch generates the Go code for a callback's switch/case
// dispatch inside a jni.NewProxy handler.
func BuildCallbackDispatch(cb *MergedCallback) string {
	var sb strings.Builder
	sb.WriteString("switch methodName {\n")
	for _, m := range cb.Methods {
		fmt.Fprintf(&sb, "\tcase %q:\n", m.JavaMethod)
		if len(m.Params) == 0 {
			fmt.Fprintf(&sb, "\t\tif cb.%s != nil {\n", m.GoField)
			fmt.Fprintf(&sb, "\t\t\tcb.%s()\n", m.GoField)
			sb.WriteString("\t\t}\n")
		} else {
			for i, p := range m.Params {
				fmt.Fprintf(&sb, "\t\t%s\n", callbackArgConversion(p, i))
			}
			fmt.Fprintf(&sb, "\t\tif cb.%s != nil {\n", m.GoField)
			var argNames []string
			for _, p := range m.Params {
				argNames = append(argNames, p.GoName)
			}
			fmt.Fprintf(&sb, "\t\t\tcb.%s(%s)\n", m.GoField, strings.Join(argNames, ", "))
			sb.WriteString("\t\t}\n")
		}
	}
	sb.WriteString("\t}")
	return sb.String()
}

// BuildCallbackType generates the Go struct type definition with function
// fields for a callback interface.
func BuildCallbackType(cb *MergedCallback) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "type %s struct {\n", cb.GoType)
	for _, m := range cb.Methods {
		if m.GoParams == "" {
			fmt.Fprintf(&sb, "\t%s func()\n", m.GoField)
		} else {
			fmt.Fprintf(&sb, "\t%s func(%s)\n", m.GoField, m.GoParams)
		}
	}
	sb.WriteString("}")
	return sb.String()
}
