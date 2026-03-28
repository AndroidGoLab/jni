package javagen

// MergedAbstractCallback is a resolved abstract callback class.
type MergedAbstractCallback struct {
	JavaClass      string
	JavaClassSlash string
	GoType         string
	Methods        []MergedAbstractCallbackMethod
}

// AdapterClassName returns the Java adapter class name for this abstract callback.
// The name uses the Java simple class name (not the Go type) with an "Adapter"
// suffix, matching the naming convention that tryAbstractAdapter in proxy.go
// searches for (e.g. "ScanCallbackAdapter" for "android.bluetooth.le.ScanCallback").
func (m *MergedAbstractCallback) AdapterClassName() string {
	return m.JavaSimpleName() + "Adapter"
}

// JavaSimpleName returns the simple (unqualified) Java class name
// (e.g. "ScanCallback" from "android.bluetooth.le.ScanCallback").
func (m *MergedAbstractCallback) JavaSimpleName() string {
	for i := len(m.JavaClass) - 1; i >= 0; i-- {
		if m.JavaClass[i] == '.' {
			return m.JavaClass[i+1:]
		}
	}
	return m.JavaClass
}

// JavaPackage returns the Java package of the abstract class
// (e.g. "android.bluetooth.le" from "android.bluetooth.le.ScanCallback").
func (m *MergedAbstractCallback) JavaPackage() string {
	for i := len(m.JavaClass) - 1; i >= 0; i-- {
		if m.JavaClass[i] == '.' {
			return m.JavaClass[:i]
		}
	}
	return ""
}
