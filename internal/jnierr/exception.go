// Package jnierr provides JNI exception-to-Go-error conversion.
//
// This is hand-written because exception extraction requires careful
// use of raw capi calls to avoid infinite recursion (every generated
// Env method calls CheckException, so it must not use Env methods).
package jnierr

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/AndroidGoLab/jni/capi"
)

// dummyJvalue is a zero-valued jvalue passed to Call*MethodA for
// zero-argument methods. The JNI spec does not guarantee NULL is valid
// for the const jvalue* parameter; OpenJ9 segfaults on NULL
// (eclipse-openj9/openj9#10480). The dummy is never read by the JVM
// when the method takes no arguments.
var dummyJvalue capi.Jvalue

// JavaException represents a Java exception caught by JNI.
type JavaException struct {
	ClassName string
	Message   string
}

// Error returns a human-readable representation of the Java exception.
func (e *JavaException) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("java %s: %s", e.ClassName, e.Message)
	}
	return fmt.Sprintf("java %s", e.ClassName)
}

// CheckException checks for a pending Java exception on the given JNI
// env pointer and converts it to a Go error. Returns nil if no exception
// is pending.
//
// Call sequence:
//  1. ExceptionCheck  — is an exception pending?
//  2. ExceptionOccurred — get the jthrowable reference
//  3. ExceptionClear  — MUST clear before making more JNI calls
//  4. Extract class name and message from the throwable
//  5. Delete the throwable local ref
//  6. Return *JavaException
func CheckException(env *capi.Env) error {
	if capi.ExceptionCheck(env) == capi.JNI_FALSE {
		return nil
	}

	throwable := capi.ExceptionOccurred(env)
	capi.ExceptionClear(env)

	className := extractClassName(env, throwable)
	message := extractMessage(env, throwable)

	capi.DeleteLocalRef(env, capi.Object(throwable))

	return &JavaException{
		ClassName: className,
		Message:   message,
	}
}

// cstr converts a Go *byte (null-terminated C string) to *capi.Cchar.
func cstr(b *byte) *capi.Cchar {
	return (*capi.Cchar)(unsafe.Pointer(b))
}

func extractClassName(env *capi.Env, throwable capi.Throwable) string {
	cls := capi.GetObjectClass(env, capi.Object(throwable))
	if cls == 0 {
		return "unknown"
	}
	defer capi.DeleteLocalRef(env, capi.Object(cls))

	classClass := capi.GetObjectClass(env, capi.Object(cls))
	if classClass == 0 {
		return "unknown"
	}
	defer capi.DeleteLocalRef(env, capi.Object(classClass))

	getNameMID := capi.GetMethodID(env, classClass, cstr(cstrGetName()), cstr(cstrVoidToString()))
	if getNameMID == nil {
		capi.ExceptionClear(env)
		return "unknown"
	}

	nameObj := capi.CallObjectMethodA(env, capi.Object(cls), getNameMID, &dummyJvalue)
	if capi.ExceptionCheck(env) == capi.JNI_TRUE {
		capi.ExceptionClear(env)
		return "unknown"
	}
	if nameObj == 0 {
		return "unknown"
	}
	defer capi.DeleteLocalRef(env, nameObj)

	return extractGoString(env, capi.String(nameObj))
}

func extractMessage(env *capi.Env, throwable capi.Throwable) string {
	cls := capi.GetObjectClass(env, capi.Object(throwable))
	if cls == 0 {
		return ""
	}
	defer capi.DeleteLocalRef(env, capi.Object(cls))

	getMsgMID := capi.GetMethodID(env, cls, cstr(cstrGetMessage()), cstr(cstrVoidToString()))
	if getMsgMID == nil {
		capi.ExceptionClear(env)
		return ""
	}

	msgObj := capi.CallObjectMethodA(env, capi.Object(throwable), getMsgMID, &dummyJvalue)
	if capi.ExceptionCheck(env) == capi.JNI_TRUE {
		capi.ExceptionClear(env)
		return ""
	}
	if msgObj == 0 {
		return ""
	}
	defer capi.DeleteLocalRef(env, msgObj)

	return extractGoString(env, capi.String(msgObj))
}

func extractGoString(env *capi.Env, jstr capi.String) string {
	if jstr == 0 {
		return ""
	}
	length := capi.GetStringUTFLength(env, jstr)
	if length == 0 {
		return ""
	}
	chars := capi.GetStringUTFChars(env, jstr, nil)
	if chars == nil {
		return ""
	}
	goStr := unsafe.String((*byte)(unsafe.Pointer(chars)), int(length))
	result := string([]byte(goStr))
	capi.ReleaseStringUTFChars(env, jstr, chars)
	return result
}

var (
	cstrGetName      = sync.OnceValue(func() *byte { return newCString("getName") })
	cstrGetMessage   = sync.OnceValue(func() *byte { return newCString("getMessage") })
	cstrVoidToString = sync.OnceValue(func() *byte { return newCString("()Ljava/lang/String;") })
)

func newCString(s string) *byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return &b[0]
}
