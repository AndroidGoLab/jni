package javagen

// MergedSpec is the fully resolved specification ready for template rendering.
type MergedSpec struct {
	Package         string
	GoImport        string
	JavaPackageDesc string

	Classes            []MergedClass
	DataClasses        []MergedDataClass
	Callbacks          []MergedCallback
	AbstractCallbacks  []MergedAbstractCallback
	ConstantGroups     []MergedConstantGroup
}
