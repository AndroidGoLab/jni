package javagen

// MergedCallbackMethod is a resolved callback method.
type MergedCallbackMethod struct {
	JavaMethod string
	GoField    string
	Params     []MergedParam
	GoParams   string
}
