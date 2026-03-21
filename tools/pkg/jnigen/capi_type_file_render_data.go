package jnigen

// CapiTypeEntry represents a single type alias in the capi layer.
type CapiTypeEntry struct {
	GoType string // exported Go name (e.g. "Object", "Jint")
	CType  string // C type name (e.g. "jobject", "jint")
}

// CapiTypeFileRenderData holds all types for a single generated capi type file.
type CapiTypeFileRenderData struct {
	Types []CapiTypeEntry
}
