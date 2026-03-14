package javagen

import (
	"fmt"
	"strings"
)

// MergedSpec is the fully resolved specification ready for template rendering.
type MergedSpec struct {
	Package         string
	GoImport        string
	JavaPackageDesc string

	Classes        []MergedClass
	DataClasses    []MergedDataClass
	Callbacks      []MergedCallback
	ConstantGroups []MergedConstantGroup
}

// MergedClass is a resolved Java class wrapper.
type MergedClass struct {
	JavaClass      string
	JavaClassSlash string
	GoType         string
	Obtain         string
	ServiceName    string
	Kind           string
	Close          bool
	Methods        []MergedMethod
}

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

// MergedMethod is a resolved method with all computed helpers.
type MergedMethod struct {
	JavaMethod string
	GoName     string
	Static     bool
	Params     []MergedParam
	Returns    string
	GoReturn   string
	Error      bool
	JNISig     string
	CallSuffix string
	ReturnKind ReturnKind

	GoParamList    string
	GoReturnList   string
	GoReturnVars   string
	GoReturnValues string
	JNIArgs        string
	HasError       bool
}

// MergedParam is a resolved parameter.
type MergedParam struct {
	JavaType       string
	GoName         string
	GoType         string
	ConversionCode string
	IsString       bool
	IsBool         bool
	IsObject       bool
}

// MergedDataClass is a resolved data class (getter extraction).
type MergedDataClass struct {
	JavaClass      string
	JavaClassSlash string
	GoType         string
	Fields         []MergedField
}

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

// MergedCallback is a resolved callback interface.
type MergedCallback struct {
	JavaInterface string
	GoType        string
	Methods       []MergedCallbackMethod
}

// MergedCallbackMethod is a resolved callback method.
type MergedCallbackMethod struct {
	JavaMethod string
	GoField    string
	Params     []MergedParam
	GoParams   string
}

// MergedConstantGroup groups constants by their Go type.
type MergedConstantGroup struct {
	GoType   string
	BaseType string
	Values   []MergedConstant
}

// MergedConstant is a resolved constant value.
type MergedConstant struct {
	GoName string
	Value  string
}

// Merge combines a Spec and Overlay into a fully resolved MergedSpec.
func Merge(spec *Spec, overlay *Overlay) (*MergedSpec, error) {
	merged := &MergedSpec{
		Package:  spec.Package,
		GoImport: spec.GoImport,
	}

	for _, cls := range spec.Classes {
		if cls.Kind == "data_class" {
			dc, err := mergeDataClass(&cls, overlay)
			if err != nil {
				return nil, fmt.Errorf("merge data class %s: %w", cls.GoType, err)
			}
			merged.DataClasses = append(merged.DataClasses, *dc)
			continue
		}

		mc, err := mergeClass(&cls, overlay)
		if err != nil {
			return nil, fmt.Errorf("merge class %s: %w", cls.GoType, err)
		}
		merged.Classes = append(merged.Classes, *mc)
	}

	for _, cb := range spec.Callbacks {
		mcb, err := mergeCallback(&cb, overlay)
		if err != nil {
			return nil, fmt.Errorf("merge callback %s: %w", cb.GoType, err)
		}
		merged.Callbacks = append(merged.Callbacks, *mcb)
	}

	merged.ConstantGroups = mergeConstants(spec.Constants)

	return merged, nil
}

func mergeClass(cls *Class, overlay *Overlay) (*MergedClass, error) {
	mc := &MergedClass{
		JavaClass:      cls.JavaClass,
		JavaClassSlash: JavaClassToSlash(cls.JavaClass),
		GoType:         cls.GoType,
		Obtain:         cls.Obtain,
		ServiceName:    cls.ServiceName,
		Kind:           cls.Kind,
		Close:          cls.Close,
	}

	methods := cls.Methods
	if overlay != nil && len(overlay.ExtraMethods) > 0 {
		methods = append(methods, overlay.ExtraMethods...)
	}

	for _, m := range methods {
		mm, err := mergeMethod(&m, overlay)
		if err != nil {
			return nil, fmt.Errorf("method %s: %w", m.GoName, err)
		}
		mc.Methods = append(mc.Methods, *mm)
	}

	return mc, nil
}

func mergeMethod(m *Method, overlay *Overlay) (*MergedMethod, error) {
	goName := m.GoName
	if overlay != nil {
		if override, ok := overlay.GoNameOverrides[m.JavaMethod]; ok {
			goName = override
		}
	}

	retType := m.Returns
	if retType == "" {
		retType = "void"
	}

	retConv := ResolveType(retType)
	goReturn := retConv.GoType
	if overlay != nil {
		if override, ok := overlay.TypeOverrides[retType]; ok {
			goReturn = override
		}
	}

	retKind := classifyReturn(retType, retConv)

	var params []MergedParam
	for _, p := range m.Params {
		mp := mergeParam(p, overlay)
		params = append(params, mp)
	}

	jniSig := JNISignature(m.Params, retType)

	mm := &MergedMethod{
		JavaMethod: m.JavaMethod,
		GoName:     goName,
		Static:     m.Static,
		Params:     params,
		Returns:    retType,
		GoReturn:   goReturn,
		Error:      m.Error,
		JNISig:     jniSig,
		CallSuffix: retConv.CallSuffix,
		ReturnKind: retKind,
		HasError:   m.Error,
	}

	isVoid := retKind == ReturnVoid
	mm.GoParamList = buildGoParamList(params)
	mm.GoReturnList = buildGoReturnList(goReturn, isVoid, m.Error)
	mm.GoReturnVars = buildGoReturnVars(goReturn, isVoid)
	mm.GoReturnValues = buildGoReturnValues(goReturn, isVoid, m.Error)
	mm.JNIArgs = buildJNIArgs(params)

	return mm, nil
}

func mergeParam(p Param, overlay *Overlay) MergedParam {
	tc := ResolveType(p.JavaType)
	goType := p.GoType
	if goType == "" {
		goType = tc.GoType
	}
	if overlay != nil {
		if override, ok := overlay.TypeOverrides[p.JavaType]; ok && p.GoType == "" {
			goType = override
		}
	}

	isString := p.JavaType == "String" || p.JavaType == "java.lang.String"
	isBool := p.JavaType == "boolean"

	goName := sanitizeGoName(p.GoName)

	return MergedParam{
		JavaType: p.JavaType,
		GoName:   goName,
		GoType:   goType,
		IsString: isString,
		IsBool:   isBool,
		IsObject: tc.IsObject,
	}
}

// goKeywords is the set of Go reserved keywords that cannot be used as identifiers.
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

// sanitizeGoName appends an underscore to names that are Go reserved keywords.
func sanitizeGoName(name string) string {
	if goKeywords[name] {
		return name + "_"
	}
	return name
}

func mergeDataClass(cls *Class, _ *Overlay) (*MergedDataClass, error) {
	dc := &MergedDataClass{
		JavaClass:      cls.JavaClass,
		JavaClassSlash: JavaClassToSlash(cls.JavaClass),
		GoType:         cls.GoType,
	}

	for _, f := range cls.Fields {
		retConv := ResolveType(f.Returns)
		goType := f.GoType
		if goType == "" {
			goType = retConv.GoType
		}

		isField := f.JavaField != ""
		var javaMethod, javaName, jniSig string
		if isField {
			javaName = f.JavaField
			jniSig = JNITypeSignature(f.Returns) // field sig: no () wrapper
		} else {
			javaMethod = f.GetterName()
			jniSig = JNISignature(nil, f.Returns) // method sig: ()RetType
		}

		dc.Fields = append(dc.Fields, MergedField{
			JavaMethod: javaMethod,
			JavaName:   javaName,
			GoName:     f.GoName,
			GoType:     goType,
			JNISig:     jniSig,
			CallSuffix: retConv.CallSuffix,
			IsField:    isField,
		})
	}

	return dc, nil
}

func mergeCallback(cb *Callback, overlay *Overlay) (*MergedCallback, error) {
	mcb := &MergedCallback{
		JavaInterface: cb.JavaInterface,
		GoType:        cb.GoType,
	}

	for _, m := range cb.Methods {
		var params []MergedParam
		for i, jt := range m.Params {
			goName := fmt.Sprintf("arg%d", i)
			isString := jt == "String" || jt == "java.lang.String"
			// Callback args arrive as autoboxed Object[] via JNI proxy.
			// Only strings get converted to Go string; everything else is *jni.Object.
			goType := "*jni.Object"
			if isString {
				goType = "string"
			}
			params = append(params, MergedParam{
				JavaType: jt,
				GoName:   goName,
				GoType:   goType,
				IsString: isString,
				IsObject: true,
			})
		}

		goParams := buildGoParamList(params)

		mcb.Methods = append(mcb.Methods, MergedCallbackMethod{
			JavaMethod: m.JavaMethod,
			GoField:    m.GoField,
			Params:     params,
			GoParams:   goParams,
		})
	}

	return mcb, nil
}

// inferBaseType guesses the Go base type from a constant value literal.
// Quoted strings → "string", everything else → "int".
func inferBaseType(value string) string {
	if len(value) >= 2 && value[0] == '"' {
		return "string"
	}
	return "int"
}

// isBuiltinType reports whether t is a Go builtin type that must not
// be used as a named type alias (e.g. "type string string" shadows the builtin).
func isBuiltinType(t string) bool {
	switch t {
	case "string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "byte", "rune":
		return true
	}
	return false
}

func mergeConstants(constants []Constant) []MergedConstantGroup {
	if len(constants) == 0 {
		return nil
	}

	groups := make(map[string]*MergedConstantGroup)
	var order []string

	for _, c := range constants {
		key := c.GoType
		if key == "" {
			key = "_untyped"
		}
		g, ok := groups[key]
		if !ok {
			baseType := c.GoUnderlying
			if baseType == "" {
				baseType = c.GoType
			}
			g = &MergedConstantGroup{
				BaseType: baseType,
			}
			// Only create a named type for non-builtin types.
			// Builtins like "string" or "int" would shadow Go builtins.
			if !isBuiltinType(key) && key != "_untyped" {
				g.GoType = key
				// If BaseType == GoType (self-referential), infer from values.
				if g.BaseType == g.GoType {
					g.BaseType = inferBaseType(c.Value)
				}
			}
			groups[key] = g
			order = append(order, key)
		}
		g.Values = append(g.Values, MergedConstant{
			GoName: c.GoName,
			Value:  c.Value,
		})
	}

	var result []MergedConstantGroup
	for _, key := range order {
		result = append(result, *groups[key])
	}
	return result
}

func buildGoParamList(params []MergedParam) string {
	var parts []string
	for _, p := range params {
		parts = append(parts, p.GoName+" "+p.GoType)
	}
	return strings.Join(parts, ", ")
}

func buildGoReturnList(goReturn string, isVoid, hasError bool) string {
	switch {
	case isVoid && hasError:
		return "error"
	case isVoid && !hasError:
		return ""
	case !isVoid && hasError:
		return "(" + goReturn + ", error)"
	default:
		return goReturn
	}
}

func buildGoReturnVars(goReturn string, isVoid bool) string {
	if isVoid {
		return ""
	}
	return "var result " + goReturn
}

func buildGoReturnValues(goReturn string, isVoid, hasError bool) string {
	switch {
	case isVoid && hasError:
		return "callErr"
	case isVoid && !hasError:
		return ""
	case !isVoid && hasError:
		return "result, callErr"
	default:
		return "result"
	}
}

func classifyReturn(retType string, retConv TypeConv) ReturnKind {
	switch {
	case retType == "void":
		return ReturnVoid
	case retType == "String" || retType == "java.lang.String":
		return ReturnString
	case retType == "boolean":
		return ReturnBool
	case retConv.IsObject:
		return ReturnObject
	default:
		return ReturnPrimitive
	}
}

func buildJNIArgs(params []MergedParam) string {
	if len(params) == 0 {
		return ""
	}
	var parts []string
	for _, p := range params {
		tc := ResolveType(p.JavaType)
		valFunc := jniValueFunc(tc.CallSuffix)
		argName := p.GoName
		switch {
		case p.IsString:
			argName = "&j" + strings.Title(p.GoName) + ".Object" //nolint:staticcheck // strings.Title is fine here
		case p.IsBool:
			argName = "j" + strings.Title(p.GoName) //nolint:staticcheck // strings.Title is fine here
		}
		parts = append(parts, fmt.Sprintf("jni.%s(%s)", valFunc, argName))
	}
	return ", " + strings.Join(parts, ", ")
}
