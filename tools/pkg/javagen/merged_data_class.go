package javagen

// MergedDataClass is a resolved data class (getter extraction).
type MergedDataClass struct {
	JavaClass      string
	JavaClassSlash string
	GoType         string
	Fields         []MergedField
}
