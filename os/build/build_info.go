package build

// BuildInfo holds static field values from android.os.Build.
type BuildInfo struct {
	Device       string
	Model        string
	Product      string
	Manufacturer string
	Brand        string
	Board        string
	Hardware     string
}
