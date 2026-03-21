package jnigen

import (
	"fmt"
	"sort"
	"strings"
)

// Merge combines a Spec and Overlay into a MergedSpec.
func Merge(spec *Spec, overlay *Overlay) (*MergedSpec, error) {
	merged := &MergedSpec{
		Primitives:  spec.Primitives,
		TypeMethods: make(map[string][]MergedMethod),
	}

	// Build suffix-to-type maps for family expansion.
	suffixToCType := buildSuffixToCTypeMap(spec.Primitives)
	suffixToGoType := buildSuffixToGoTypeMap(spec.Primitives)

	// Merge reference types.
	for _, rt := range spec.ReferenceTypes {
		goType := overlay.TypeRenames[rt.CType]
		if goType == "" {
			goType = rt.CType
		}
		parentGo := ""
		if rt.Parent != nil {
			parentGo = overlay.TypeRenames[*rt.Parent]
			if parentGo == "" {
				parentGo = *rt.Parent
			}
		}
		isArray := strings.HasSuffix(rt.CType, "Array")
		elemType := ""
		if isArray && rt.CType != "jarray" {
			// Extract element type: e.g., jintArray -> Int
			base := strings.TrimSuffix(rt.CType, "Array")
			base = strings.TrimPrefix(base, "j")
			for _, p := range spec.Primitives {
				trimmed := strings.TrimPrefix(p.CType, "j")
				if trimmed == base {
					elemType = p.Suffix
					break
				}
			}
			if base == "object" {
				elemType = "Object"
			}
		}
		merged.ReferenceTypes = append(merged.ReferenceTypes, MergedRefType{
			CType:    rt.CType,
			GoType:   goType,
			Parent:   parentGo,
			IsArray:  isArray,
			ElemType: elemType,
		})
	}

	// Merge opaque types.
	for _, ot := range spec.OpaqueTypes {
		goType := overlay.TypeRenames[ot.CType]
		if goType == "" {
			if ot.GoType != "" {
				goType = ot.GoType
			} else {
				goType = ot.CType
			}
		}
		merged.OpaqueTypes = append(merged.OpaqueTypes, MergedOpaqueType{
			CType:    ot.CType,
			GoType:   goType,
			CapiType: camelCase(ot.CType),
		})
	}

	// Merge constants.
	for _, c := range spec.Constants {
		merged.Constants = append(merged.Constants, MergedConstant{
			CName:  c.Name,
			GoName: c.Name,
			Value:  c.Value,
			GoType: c.GoType,
		})
	}

	// Expand function families into individual functions.
	var expandedFuncs []expandedFunc
	for _, fam := range spec.FunctionFamilies {
		famOverlay, hasFamOverlay := overlay.FamilyOverlays[fam.Pattern]

		for _, suffix := range fam.Expand {
			cType := suffixToCType[suffix]
			goType := suffixToGoType[suffix]

			// Build the C function name by replacing {Type} with suffix.
			cName := strings.ReplaceAll(fam.Pattern, "{Type}", suffix)

			// Build capi params.
			var capiParams []CapiParam
			envParamName := "env"
			if fam.Vtable == "JavaVM" {
				envParamName = "vm"
			}
			capiParams = append(capiParams, CapiParam{
				CType: fam.Vtable + "*",
				CName: envParamName,
			})

			for i, p := range fam.Params {
				paramType := expandTypeParam(p, suffix, cType)
				paramName := fmt.Sprintf("p%d", i)
				capiParams = append(capiParams, CapiParam{
					CType: paramType,
					CName: paramName,
				})
			}

			// Build C return type.
			retType := expandReturnType(fam.Returns, suffix, cType)
			isVoid := retType == "void"

			capiFunc := MergedCapiFunc{
				CName:      cName,
				HelperName: "jni_" + cName,
				Vtable:     fam.Vtable,
				Params:     capiParams,
				Returns:    retType,
				IsVoid:     isVoid,
			}

			ef := expandedFunc{
				capi:      capiFunc,
				specName:  cName,
				vtable:    fam.Vtable,
				exception: fam.Exception,
				suffix:    suffix,
				cType:     cType,
				goType:    goType,
			}

			if hasFamOverlay {
				ef.famOverlay = &famOverlay
			}

			expandedFuncs = append(expandedFuncs, ef)
		}
	}

	// Process env_functions.
	for _, f := range spec.EnvFunctions {
		fo, hasFO := overlay.Functions[f.Name]
		if !hasFO {
			return nil, fmt.Errorf("missing overlay for env function %q", f.Name)
		}

		var capiParams []CapiParam
		capiParams = append(capiParams, CapiParam{CType: "JNIEnv*", CName: "env"})
		for i, p := range f.Params {
			capiParams = append(capiParams, CapiParam{
				CType: p,
				CName: fmt.Sprintf("p%d", i),
			})
		}

		retType := f.Returns
		if retType == "" {
			retType = "void"
		}
		isVoid := retType == "void"

		capiFunc := MergedCapiFunc{
			CName:      f.Name,
			HelperName: "jni_" + f.Name,
			Vtable:     "JNIEnv",
			Params:     capiParams,
			Returns:    retType,
			IsVoid:     isVoid,
		}

		merged.CapiFunctions = append(merged.CapiFunctions, capiFunc)

		// Build idiomatic method.
		method := buildMethod(f.Name, "JNIEnv", fo, overlay)
		merged.EnvMethods = append(merged.EnvMethods, method)
	}

	// Process vm_functions.
	for _, f := range spec.VMFunctions {
		fo, hasFO := overlay.Functions[f.Name]
		if !hasFO {
			return nil, fmt.Errorf("missing overlay for vm function %q", f.Name)
		}

		var capiParams []CapiParam
		capiParams = append(capiParams, CapiParam{CType: "JavaVM*", CName: "vm"})
		for i, p := range f.Params {
			cp := CapiParam{
				CType: p,
				CName: fmt.Sprintf("p%d", i),
			}
			// The C JNI API declares AttachCurrentThread with void** but
			// we use JNIEnv** for type safety; cast in the vtable call.
			if p == "JNIEnv**" {
				cp.VtableCast = "void**"
			}
			capiParams = append(capiParams, cp)
		}

		retType := f.Returns
		if retType == "" {
			retType = "void"
		}
		isVoid := retType == "void"

		capiFunc := MergedCapiFunc{
			CName:      f.Name,
			HelperName: "jni_" + f.Name,
			Vtable:     "JavaVM",
			Params:     capiParams,
			Returns:    retType,
			IsVoid:     isVoid,
		}

		merged.CapiFunctions = append(merged.CapiFunctions, capiFunc)

		method := buildMethod(f.Name, "JavaVM", fo, overlay)
		merged.VMMethods = append(merged.VMMethods, method)
	}

	// Add expanded family functions to capi and idiomatic layers.
	for _, ef := range expandedFuncs {
		merged.CapiFunctions = append(merged.CapiFunctions, ef.capi)

		if ef.famOverlay != nil {
			method := buildFamilyMethod(ef, overlay)
			merged.EnvMethods = append(merged.EnvMethods, method)
		}
	}

	// Sort capi functions for deterministic output.
	sort.Slice(merged.CapiFunctions, func(i, j int) bool {
		return merged.CapiFunctions[i].CName < merged.CapiFunctions[j].CName
	})
	sort.Slice(merged.EnvMethods, func(i, j int) bool {
		return merged.EnvMethods[i].GoName < merged.EnvMethods[j].GoName
	})
	sort.Slice(merged.VMMethods, func(i, j int) bool {
		return merged.VMMethods[i].GoName < merged.VMMethods[j].GoName
	})

	return merged, nil
}

type expandedFunc struct {
	capi       MergedCapiFunc
	specName   string
	vtable     string
	exception  bool
	suffix     string
	cType      string
	goType     string
	famOverlay *FamilyOverlay
}

func buildSuffixToCTypeMap(primitives []Primitive) map[string]string {
	m := make(map[string]string, len(primitives)+2)
	for _, p := range primitives {
		m[p.Suffix] = p.CType
	}
	m["Object"] = "jobject"
	m["Void"] = "void"
	return m
}

func buildSuffixToGoTypeMap(primitives []Primitive) map[string]string {
	m := make(map[string]string, len(primitives)+2)
	for _, p := range primitives {
		m[p.Suffix] = p.GoType
	}
	m["Object"] = "*Object"
	m["Void"] = ""
	return m
}

func expandTypeParam(param, suffix, cType string) string {
	result := strings.ReplaceAll(param, "{Type}", suffix)
	result = strings.ReplaceAll(result, "{type}", cType)
	return result
}

func expandReturnType(retPattern, suffix, cType string) string {
	if retPattern == "void" {
		return "void"
	}
	result := strings.ReplaceAll(retPattern, "{Type}", suffix)
	result = strings.ReplaceAll(result, "{type}", cType)
	return result
}

func buildMethod(
	name string,
	vtable string,
	fo FuncOverlay,
	overlay *Overlay,
) MergedMethod {
	receiver := overlay.Receivers["env_functions"]
	if vtable == "JavaVM" {
		receiver = overlay.Receivers["vm_functions"]
	}

	method := MergedMethod{
		GoName:         fo.GoName,
		Receiver:       receiver,
		CapiCall:       name,
		CheckException: fo.CheckException,
	}

	for _, p := range fo.Params {
		if p.Implicit != "" {
			method.AllParams = append(method.AllParams, CapiArgInfo{
				GoType:     p.GoType,
				CType:      p.CType,
				IsImplicit: true,
				Implicit:   p.Implicit,
			})
			continue
		}
		isVariadic := strings.HasPrefix(p.GoType, "...")
		method.Params = append(method.Params, GoParam{
			Name:       p.Name,
			GoType:     p.GoType,
			CType:      p.CType,
			IsVariadic: isVariadic,
		})
		method.AllParams = append(method.AllParams, CapiArgInfo{
			Name:       p.Name,
			GoType:     p.GoType,
			CType:      p.CType,
			IsVariadic: isVariadic,
		})
	}

	if fo.Returns != nil && fo.Returns.GoType != "" {
		isError := fo.Returns.GoType == "error"
		method.Returns = append(method.Returns, GoReturn{
			GoType:    fo.Returns.GoType,
			IsError:   isError,
			Transform: fo.Returns.Transform,
		})
	}

	if fo.CheckException {
		hasError := false
		for _, r := range method.Returns {
			if r.IsError {
				hasError = true
				break
			}
		}
		if !hasError {
			method.Returns = append(method.Returns, GoReturn{
				GoType:  "error",
				IsError: true,
			})
		}
	}

	return method
}

func buildFamilyMethod(ef expandedFunc, overlay *Overlay) MergedMethod {
	fo := ef.famOverlay
	receiver := overlay.Receivers["function_families"]

	goName := strings.ReplaceAll(fo.GoPattern, "{Type}", ef.suffix)

	method := MergedMethod{
		GoName:         goName,
		Receiver:       receiver,
		CapiCall:       ef.specName,
		CheckException: fo.CheckException || ef.exception,
	}

	for _, p := range fo.Params {
		goType := strings.ReplaceAll(p.GoType, "{Type}", ef.suffix)
		goType = strings.ReplaceAll(goType, "{go_type}", ef.goType)
		goType = strings.ReplaceAll(goType, "{type}", ef.cType)

		cType := strings.ReplaceAll(p.CType, "{Type}", ef.suffix)
		cType = strings.ReplaceAll(cType, "{type}", ef.cType)

		if p.Implicit != "" {
			implicit := strings.ReplaceAll(p.Implicit, "{Type}", ef.suffix)
			method.AllParams = append(method.AllParams, CapiArgInfo{
				GoType:     goType,
				CType:      cType,
				IsImplicit: true,
				Implicit:   implicit,
			})
			continue
		}

		isVariadic := strings.HasPrefix(goType, "...")
		method.Params = append(method.Params, GoParam{
			Name:       p.Name,
			GoType:     goType,
			CType:      cType,
			IsVariadic: isVariadic,
		})
		method.AllParams = append(method.AllParams, CapiArgInfo{
			Name:       p.Name,
			GoType:     goType,
			CType:      cType,
			IsVariadic: isVariadic,
		})
	}

	// Determine return type from family overlay.
	retMap := fo.FamilyReturnMap()
	if retMap != nil {
		var goRetType string
		if allRet, ok := retMap["_all"]; ok {
			// Single return type for all expansions.
			goRetType = strings.ReplaceAll(allRet, "{Type}", ef.suffix)
			goRetType = strings.ReplaceAll(goRetType, "{go_type}", ef.goType)
		} else {
			switch ef.suffix {
			case "Object":
				goRetType = retMap["Object"]
			case "Void":
				goRetType = retMap["Void"]
			default:
				// Primitive: use the "primitive" key.
				tmpl := retMap["primitive"]
				if tmpl != "" {
					goRetType = strings.ReplaceAll(tmpl, "{go_type}", ef.goType)
				}
			}
		}

		if goRetType != "" {
			method.Returns = append(method.Returns, GoReturn{
				GoType: goRetType,
			})
		}
	}

	// Add error return for exception-checking methods.
	if method.CheckException {
		hasError := false
		for _, r := range method.Returns {
			if r.IsError {
				hasError = true
				break
			}
		}
		if !hasError {
			method.Returns = append(method.Returns, GoReturn{
				GoType:  "error",
				IsError: true,
			})
		}
	}

	return method
}
