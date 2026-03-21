package jnigen

// ErrorCode holds data for a single JNI error constant.
type ErrorCode struct {
	GoName      string
	Value       string
	Description string
}
