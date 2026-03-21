package jnigen

// RenderMethod is a pre-computed method ready for template rendering.
type RenderMethod struct {
	GoName            string
	GoParamList       string
	GoParams          []string // individual "name Type" entries for multi-line rendering
	GoReturnList      string
	Transforms        []string
	HasReturn         bool
	CapiName          string
	CapiArgs          string
	PostTransforms    []string
	CheckException    bool
	ZeroReturn        string
	ReturnConversion  string
	NullableReturn    bool   // true if _ret == 0 means Java null → Go nil
	NullReturn        string // Go expression for the null/nil return (e.g. "nil" or "nil, nil")
	ReturnsError      bool
	JNIErrorTransform bool // true if the only return is error with jni_error transform
	Doc               string
}
