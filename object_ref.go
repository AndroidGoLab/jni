package jni

import (
	"unsafe"

	"github.com/AndroidGoLab/jni/capi"
)

// CAPIObject is an alias for capi.Object (C.jobject), exported so that
// packages outside capi can reference the type for unsafe conversions
// between different CGO compilation units.
type CAPIObject = capi.Object

// ObjectFromRef wraps a raw JNI jobject reference in an Object.
// The caller is responsible for ensuring the reference is valid.
func ObjectFromRef(ref capi.Object) *Object {
	return &Object{ref: ref}
}

// ObjectFromPtr wraps a raw C jobject pointer in an Object, mirroring
// VMFromPtr. Use this from NativeActivity callbacks to avoid importing
// the capi package:
//
//	jni.ObjectFromPtr(unsafe.Pointer(activity.clazz))
func ObjectFromPtr(ptr unsafe.Pointer) *Object {
	return &Object{ref: capi.Object(uintptr(ptr))}
}
