package jni

import (
	"unsafe"

	"github.com/xaionaro-go/jni/capi"
)

// GoString converts a JNI String to a Go string using modified UTF-8 encoding.
// If str is nil, it returns an empty string.
func (e *Env) GoString(str *String) string {
	if str == nil {
		return ""
	}
	jstr := capi.String(str.Ref())
	if jstr == 0 {
		return ""
	}
	length := capi.GetStringUTFLength(e.ptr, jstr)
	if length == 0 {
		return ""
	}
	chars := capi.GetStringUTFChars(e.ptr, jstr, nil)
	if chars == nil {
		return ""
	}
	s := unsafe.String((*byte)(unsafe.Pointer(chars)), int(length))
	result := string([]byte(s)) // copy to detach from JNI memory
	capi.ReleaseStringUTFChars(e.ptr, jstr, chars)
	return result
}
