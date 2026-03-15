package app

import "github.com/AndroidGoLab/jni"

// ContextFromObject wraps an existing Android Context JNI global reference
// into an app.Context. The caller is responsible for ensuring the object
// is a valid android.content.Context and that it is a global reference.
// The returned Context does NOT own the reference — calling Close() is a no-op.
func ContextFromObject(vm *jni.VM, obj *jni.GlobalRef) *Context {
	return &Context{
		VM:  vm,
		Obj: obj,
	}
}
