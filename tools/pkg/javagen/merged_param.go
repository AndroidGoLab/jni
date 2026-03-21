package javagen

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
