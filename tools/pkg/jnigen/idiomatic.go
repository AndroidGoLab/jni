package jnigen

import (
	"fmt"
	"strings"
)

// BuildTypesData prepares data for the types.go.tmpl template.
func BuildTypesData(merged *MergedSpec) *TypesRenderData {
	data := &TypesRenderData{}
	for _, c := range merged.Constants {
		if strings.HasPrefix(c.CName, "JNI_VERSION_") {
			data.VersionConstants = append(data.VersionConstants, VersionConstant{
				GoName: c.CName,
			})
		}
	}
	return data
}

// BuildEnvData prepares data for the env.go.tmpl template.
func BuildEnvData(merged *MergedSpec) *EnvRenderData {
	data := &EnvRenderData{}
	for _, m := range merged.EnvMethods {
		data.Methods = append(data.Methods, buildRenderMethod(m))
	}
	return data
}

// BuildVMData prepares data for the vm.go.tmpl template.
func BuildVMData(merged *MergedSpec) *VMRenderData {
	data := &VMRenderData{}
	for _, m := range merged.VMMethods {
		data.Methods = append(data.Methods, buildRenderMethod(m))
	}
	return data
}

// BuildTypeData prepares data for a single reference type.
func BuildTypeData(rt MergedRefType) *TypeRenderData {
	data := &TypeRenderData{
		GoName:   rt.GoType,
		CType:    rt.CType,
		CapiType: rt.GoType, // matches the capi export name from overlay renames
		IsArray:  rt.IsArray,
		ElemType: rt.ElemType,
	}
	if rt.Parent != "" {
		data.Parent = &TypeParent{GoName: rt.Parent}
	}
	return data
}

// BuildValueData prepares data for the jvalue.go.tmpl template.
func BuildValueData(merged *MergedSpec) *ValueRenderData {
	data := &ValueRenderData{}
	for _, p := range merged.Primitives {
		data.Primitives = append(data.Primitives, ValuePrimitive{
			GoName:  p.Suffix,
			GoType:  p.GoType,
			CGoType: "capi." + camelCase(p.CType),
		})
	}
	return data
}

// BuildErrorData prepares data for the errors.go.tmpl template.
func BuildErrorData(merged *MergedSpec) *ErrorRenderData {
	data := &ErrorRenderData{}
	for _, c := range merged.Constants {
		// Only include JNI error codes (negative value or JNI_OK).
		if !strings.HasPrefix(c.CName, "JNI_E") && c.CName != "JNI_ERR" && c.CName != "JNI_OK" {
			continue
		}
		desc := errorDescription(c.CName)
		data.ErrorCodes = append(data.ErrorCodes, ErrorCode{
			GoName:      c.CName,
			Value:       c.Value,
			Description: desc,
		})
	}
	return data
}

func errorDescription(name string) string {
	switch name {
	case "JNI_OK":
		return "success"
	case "JNI_ERR":
		return "general error"
	case "JNI_EDETACHED":
		return "thread detached"
	case "JNI_EVERSION":
		return "version error"
	case "JNI_ENOMEM":
		return "out of memory"
	case "JNI_EEXIST":
		return "VM already exists"
	case "JNI_EINVAL":
		return "invalid argument"
	default:
		return "unknown error"
	}
}

// isTypedWrapperReplaced returns true for Env methods whose functionality
// is fully covered by the typed wrapper packages (app/, content/, provider/,
// etc.). These are the methods that require JNI class names and method
// signatures as magic strings.
func isTypedWrapperReplaced(goName string) bool {
	switch {
	case goName == "FindClass",
		goName == "DefineClass",
		goName == "GetObjectClass",
		goName == "NewObject",
		goName == "AllocObject":
		return true
	case strings.HasPrefix(goName, "Call"):
		return true
	case strings.HasPrefix(goName, "GetMethodID"),
		strings.HasPrefix(goName, "GetStaticMethodID"),
		strings.HasPrefix(goName, "GetFieldID"),
		strings.HasPrefix(goName, "GetStaticFieldID"):
		return true
	case strings.HasPrefix(goName, "Get") && strings.HasSuffix(goName, "Field"):
		return true
	case strings.HasPrefix(goName, "Set") && strings.HasSuffix(goName, "Field"):
		return true
	case strings.HasPrefix(goName, "GetStatic") && strings.HasSuffix(goName, "Field"):
		return true
	case strings.HasPrefix(goName, "SetStatic") && strings.HasSuffix(goName, "Field"):
		return true
	default:
		return false
	}
}

func buildRenderMethod(m MergedMethod) RenderMethod {
	rm := RenderMethod{
		GoName:         m.GoName,
		CapiName:       m.CapiCall,
		CheckException: m.CheckException,
	}

	// Build doc comment.
	rm.Doc = fmt.Sprintf("// %s wraps the JNI %s function.", m.GoName, m.CapiCall)
	if isTypedWrapperReplaced(m.GoName) {
		rm.Doc += "\n//\n// This is a low-level call. Prefer the typed wrappers (e.g. packages under\n// app/, content/, provider/) which handle JNI signatures, local/global\n// reference management, and error checking automatically. Use Env methods\n// directly only as a last resort for functionality not yet covered by a\n// typed wrapper."
	}

	// Build Go parameter list.
	var goParams []string
	for _, p := range m.Params {
		goParams = append(goParams, p.Name+" "+p.GoType)
	}
	rm.GoParams = goParams
	rm.GoParamList = strings.Join(goParams, ", ")

	// Build Go return list.
	var goReturns []string
	for _, r := range m.Returns {
		if r.IsError {
			goReturns = append(goReturns, "error")
		} else {
			goReturns = append(goReturns, r.GoType)
		}
	}
	switch len(goReturns) {
	case 0:
		rm.GoReturnList = ""
	case 1:
		rm.GoReturnList = goReturns[0]
	default:
		rm.GoReturnList = "(" + strings.Join(goReturns, ", ") + ")"
	}

	// Determine return characteristics.
	rm.HasReturn = false
	rm.ReturnsError = false
	var nonErrorReturn *GoReturn
	hasJNIErrorTransform := false
	for i := range m.Returns {
		if m.Returns[i].IsError {
			rm.ReturnsError = true
			if m.Returns[i].Transform == "jni_error" {
				hasJNIErrorTransform = true
			}
		} else {
			rm.HasReturn = true
			nonErrorReturn = &m.Returns[i]
		}
	}

	// If the method's only return is error with jni_error transform,
	// we need to capture the capi result and convert it.
	if hasJNIErrorTransform && !rm.HasReturn {
		rm.JNIErrorTransform = true
	}

	// Build transforms (pre-call conversions).
	rm.Transforms = buildTransforms(m.Params)

	// Build capi call args (the part after the env/vm pointer).
	rm.CapiArgs = buildCapiArgs(m.AllParams)

	// Build post-transforms (after capi call).
	rm.PostTransforms = buildPostTransforms(m.Params)

	// Build zero return for early error returns.
	if rm.HasReturn && nonErrorReturn != nil {
		rm.ZeroReturn = zeroValue(nonErrorReturn.GoType) + ", "
	}

	// Build return conversion.
	if rm.HasReturn && nonErrorReturn != nil {
		rm.ReturnConversion = buildReturnConversion(nonErrorReturn.GoType, nonErrorReturn.Transform)
		// Object-like return types can be Java null (zero ref). In Go we
		// represent that as a nil pointer so callers can check naturally.
		rm.NullableReturn = isNullableReturnType(nonErrorReturn.GoType)
		if rm.NullableReturn {
			if rm.CheckException {
				rm.NullReturn = "nil, nil"
			} else {
				rm.NullReturn = "nil"
			}
		}
	}

	return rm
}

func buildTransforms(params []GoParam) []string {
	var transforms []string
	for _, p := range params {
		cn := "c" + capitalizeFirst(p.Name)
		switch {
		case p.GoType == "string":
			// Allocate a null-terminated byte slice for C string interop.
			transforms = append(transforms,
				fmt.Sprintf("%s := append([]byte(%s), 0)", cn, p.Name))
		case p.GoType == "bool" && !p.IsVariadic:
			transforms = append(transforms,
				fmt.Sprintf("var %s capi.Jboolean; if %s { %s = capi.JNI_TRUE }",
					cn, p.Name, cn))
		case p.GoType == "...Value":
			transforms = append(transforms,
				fmt.Sprintf(
					"// _dummyJvalue ensures a non-nil jvalue pointer for zero-argument calls.\n"+
						"\t// The JNI spec does not guarantee NULL is valid for the const jvalue*\n"+
						"\t// parameter of Call<Type>MethodA; OpenJ9 rejects it as an error\n"+
						"\t// (eclipse-openj9/openj9#10480). The dummy is never read by the JVM\n"+
						"\t// when the method takes no arguments.\n"+
						"\tvar _dummyJvalue capi.Jvalue; %s := &_dummyJvalue; if len(%s) > 0 { %s = &%s[0].val }",
					cn, p.Name, cn, p.Name))
		}
	}
	return transforms
}

func buildCapiArgs(allParams []CapiArgInfo) string {
	var args []string
	for _, ai := range allParams {
		if ai.IsImplicit {
			capiType := goTypeToCapi(ai.GoType)
			if capiType != "" {
				args = append(args, fmt.Sprintf("%s(%s)", capiType, ai.Implicit))
			} else {
				args = append(args, ai.Implicit)
			}
		} else {
			arg := capiArgExpr(ai)
			args = append(args, arg)
		}
	}
	if len(args) == 0 {
		return ""
	}
	return ", " + strings.Join(args, ", ")
}

// capiArgExpr generates the Go expression to pass a param to a capi function.
func capiArgExpr(ai CapiArgInfo) string {
	goType := ai.GoType
	name := ai.Name

	cn := "c" + capitalizeFirst(name)
	switch {
	case goType == "string":
		return fmt.Sprintf("(*capi.Cchar)(unsafe.Pointer(&%s[0]))", cn)
	case goType == "bool" && !ai.IsVariadic:
		return cn
	case goType == "...Value":
		return cn

	// Reference type parameters: extract the capi ref.
	case goType == "*Object":
		return name + ".Ref()"
	case goType == "*Class":
		return "capi.Class(" + name + ".Ref())"
	case goType == "*String":
		return "capi.String(" + name + ".Ref())"
	case goType == "*Throwable":
		return "capi.Throwable(" + name + ".Ref())"
	case goType == "*Array":
		return "capi.Array(" + name + ".Ref())"
	case goType == "*ObjectArray":
		return "capi.ObjectArray(" + name + ".Ref())"
	case goType == "*WeakRef":
		return "capi.WeakRef(" + name + ".Ref())"

	// Typed arrays — capi export name matches the Go type name.
	case strings.HasPrefix(goType, "*") && strings.HasSuffix(goType, "Array"):
		typeName := strings.TrimPrefix(goType, "*")
		return fmt.Sprintf("capi.%s(%s.Ref())", typeName, name)

	// Opaque types: pass through.
	case goType == "MethodID" || goType == "FieldID":
		return name

	// Primitive types that match capi aliases.
	case goType == "int32" || goType == "int64" || goType == "float32" || goType == "float64":
		return capiCast(goType, name)
	case goType == "uint8" || goType == "int8" || goType == "uint16" || goType == "int16":
		return capiCast(goType, name)

	// unsafe.Pointer: may need a typed cast based on the original C type.
	case goType == "unsafe.Pointer":
		cast := cTypeToCapiPtrCast(ai.CType)
		if cast != "" {
			return fmt.Sprintf("%s(%s)", cast, name)
		}
		return name
	case goType == "*uint8":
		return fmt.Sprintf("(*capi.Jboolean)(%s)", name)
	case goType == "[]byte":
		cast := cTypeToCapiPtrCast(ai.CType)
		if cast == "" {
			cast = "(*capi.Jbyte)"
		}
		return fmt.Sprintf("%s(unsafe.Pointer(&%s[0]))", cast, name)
	case goType == "[]uint16":
		cast := cTypeToCapiPtrCast(ai.CType)
		if cast == "" {
			cast = "(*capi.Jchar)"
		}
		return fmt.Sprintf("%s(unsafe.Pointer(&%s[0]))", cast, name)
	case goType == "ObjectRefType":
		return name

	default:
		return name
	}
}

// cTypeToCapiPtrCast maps a C pointer type to its capi pointer cast expression.
// Returns "" if no cast is needed (e.g., void* maps to unsafe.Pointer natively).
func cTypeToCapiPtrCast(cType string) string {
	base := strings.TrimPrefix(cType, "const ")
	base = strings.TrimSuffix(base, "*")
	base = strings.TrimSpace(base)
	switch base {
	case "jboolean":
		return "(*capi.Jboolean)"
	case "jbyte":
		return "(*capi.Jbyte)"
	case "jchar":
		return "(*capi.Jchar)"
	case "jshort":
		return "(*capi.Jshort)"
	case "jint":
		return "(*capi.Jint)"
	case "jlong":
		return "(*capi.Jlong)"
	case "jfloat":
		return "(*capi.Jfloat)"
	case "jdouble":
		return "(*capi.Jdouble)"
	case "char":
		return "(*capi.Cchar)"
	case "JNINativeMethod":
		return "(*capi.JNINativeMethod)"
	default:
		return ""
	}
}

func capiCast(goType, name string) string {
	capiType := goTypeToCapi(goType)
	if capiType != "" {
		return fmt.Sprintf("%s(%s)", capiType, name)
	}
	return name
}

func goTypeToCapi(goType string) string {
	switch goType {
	case "int32":
		return "capi.Jint"
	case "int64":
		return "capi.Jlong"
	case "float32":
		return "capi.Jfloat"
	case "float64":
		return "capi.Jdouble"
	case "uint8":
		return "capi.Jboolean"
	case "int8":
		return "capi.Jbyte"
	case "uint16":
		return "capi.Jchar"
	case "int16":
		return "capi.Jshort"
	default:
		return ""
	}
}

func buildPostTransforms(params []GoParam) []string {
	// Currently no post-transforms needed.
	return nil
}

func buildReturnConversion(goType, transform string) string {
	switch goType {
	case "*Object":
		return "&Object{ref: _ret}"
	case "*Class":
		return "&Class{Object{ref: capi.Object(_ret)}}"
	case "*String":
		return "&String{Object{ref: capi.Object(_ret)}}"
	case "*Throwable":
		return "&Throwable{Object{ref: capi.Object(_ret)}}"
	case "*Array":
		return "&Array{Object{ref: capi.Object(_ret)}}"
	case "*ObjectArray":
		return "&ObjectArray{Array{Object{ref: capi.Object(_ret)}}}"
	case "*WeakRef":
		return "&WeakRef{Object{ref: capi.Object(_ret)}}"
	case "bool":
		return "_ret != 0"
	case "int32":
		return "int32(_ret)"
	case "int64":
		return "int64(_ret)"
	case "float32":
		return "float32(_ret)"
	case "float64":
		return "float64(_ret)"
	case "uint8":
		return "uint8(_ret)"
	case "int8":
		return "int8(_ret)"
	case "uint16":
		return "uint16(_ret)"
	case "int16":
		return "int16(_ret)"
	case "MethodID":
		return "_ret"
	case "FieldID":
		return "_ret"
	case "unsafe.Pointer":
		return "unsafe.Pointer(_ret)"
	case "error":
		// jni_error transform
		return ""
	case "ObjectRefType":
		return "ObjectRefType(_ret)"
	default:
		// Typed arrays
		if strings.HasPrefix(goType, "*") && strings.HasSuffix(goType, "Array") {
			typeName := strings.TrimPrefix(goType, "*")
			return "&" + typeName + "{Array{Object{ref: capi.Object(_ret)}}}"
		}
		return "_ret"
	}
}

// isNullableReturnType reports whether the Go return type wraps a JNI
// reference that can be Java null (zero). When true, the generated code
// must check _ret == 0 and return Go nil instead of a non-nil wrapper
// around a zero ref.
func isNullableReturnType(goType string) bool {
	switch goType {
	case "*Object", "*Class", "*String", "*Throwable",
		"*Array", "*ObjectArray", "*WeakRef":
		return true
	}
	// Typed arrays (*BooleanArray, *ByteArray, etc.)
	if strings.HasPrefix(goType, "*") && strings.HasSuffix(goType, "Array") {
		return true
	}
	return false
}

func zeroValue(goType string) string {
	switch goType {
	case "bool":
		return "false"
	case "int32", "int64", "float32", "float64", "uint8", "int8", "uint16", "int16":
		return "0"
	case "MethodID":
		return "nil"
	case "FieldID":
		return "nil"
	case "unsafe.Pointer":
		return "nil"
	case "ObjectRefType":
		return "0"
	default:
		if strings.HasPrefix(goType, "*") {
			return "nil"
		}
		return "0"
	}
}
