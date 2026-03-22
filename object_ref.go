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

// ObjectFromUintptr wraps a uintptr-encoded jobject in an Object.
// This is the convention used by gomobile (RunOnJVM passes context
// as uintptr) and other Go Android frameworks.
//
// unsafe.Add(nil, ptr) is used instead of unsafe.Pointer(ptr) to avoid
// a go vet "possible misuse of unsafe.Pointer" false positive.
func ObjectFromUintptr(ptr uintptr) *Object {
	return ObjectFromPtr(unsafe.Add(unsafe.Pointer(nil), ptr))
}
