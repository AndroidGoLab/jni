package javagen

// MergedCallback is a resolved callback interface.
type MergedCallback struct {
	JavaInterface string
	GoType        string
	Methods       []MergedCallbackMethod
}
