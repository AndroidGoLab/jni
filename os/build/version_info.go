package build

// VersionInfo holds static field values from android.os.Build.VERSION.
type VersionInfo struct {
	Release     string
	SDKInt      int32
	Codename    string
	Incremental string
}
