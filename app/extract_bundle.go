package app

import "github.com/AndroidGoLab/jni"

// ExtractBundle wraps a JNI Bundle object for Go access.
// Returns a non-nil Bundle if the object is valid.
// The caller must provide the VM so that the returned Bundle can call
// JNI methods via VM.Do.
func ExtractBundle(vm *jni.VM, env *jni.Env, obj *jni.Object) (*Bundle, error) {
	if obj == nil || obj.Ref() == 0 {
		return nil, nil
	}
	return &Bundle{VM: vm, Obj: env.NewGlobalRef(obj)}, nil
}
