package jnigen

// CapiRenderData holds all data needed for capi template rendering.
type CapiRenderData struct {
	EnvFunctions []MergedCapiFunc
	VMFunctions  []MergedCapiFunc
	AllFunctions []MergedCapiFunc
	Types        []MergedRefType
	OpaqueTypes  []MergedOpaqueType
	Constants    []MergedConstant
	Primitives   []Primitive
}

// BuildCapiData prepares data for capi templates.
func BuildCapiData(merged *MergedSpec) *CapiRenderData {
	data := &CapiRenderData{
		Types:       merged.ReferenceTypes,
		OpaqueTypes: merged.OpaqueTypes,
		Constants:   merged.Constants,
		Primitives:  merged.Primitives,
	}

	for _, f := range merged.CapiFunctions {
		data.AllFunctions = append(data.AllFunctions, f)
		switch f.Vtable {
		case "JNIEnv":
			data.EnvFunctions = append(data.EnvFunctions, f)
		case "JavaVM":
			data.VMFunctions = append(data.VMFunctions, f)
		}
	}

	return data
}
