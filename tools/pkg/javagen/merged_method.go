package javagen

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

	GoParamList          string
	GoParamListMultiLine string
	GoReturnList         string
	GoReturnVars         string
	GoReturnValues       string
	JNIArgs              string
	HasError             bool
}
