package jnigen

// TypeRenderData holds data for rendering a single reference type file.
type TypeRenderData struct {
	GoName   string
	CType    string
	CapiType string
	Parent   *TypeParent
	IsArray  bool
	ElemType string
}
