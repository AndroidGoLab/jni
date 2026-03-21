package specgen

// PackageMapping defines how a Java package maps to a Go package.
type PackageMapping struct {
	JavaPrefix string // e.g. "android.app.admin"
	Package    string // e.g. "admin"
	GoImport   string // e.g. "github.com/AndroidGoLab/jni/app/admin"
}
