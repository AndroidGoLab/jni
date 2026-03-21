package specgen

// JavapClass holds the parsed output of javap for one class.
type JavapClass struct {
	FullName    string // e.g. "android.app.KeyguardManager"
	IsInterface bool
	IsAbstract  bool
	IsFinal     bool
	Constants   []JavapConstant
	Methods     []JavapMethod
	Implements  []string
}
