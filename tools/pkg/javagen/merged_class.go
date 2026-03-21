package javagen

// MergedClass is a resolved Java class wrapper.
type MergedClass struct {
	JavaClass      string
	JavaClassSlash string
	GoType         string
	Obtain         string
	ServiceName    string
	Kind           string
	Close          bool
	Methods        []MergedMethod
}
