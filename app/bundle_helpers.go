package app

import "github.com/AndroidGoLab/jni"

// ExtractBundle wraps a JNI Bundle object for Go access.
// Returns a non-nil Bundle if the object is valid.
func ExtractBundle(env *jni.Env, obj *jni.Object) (*Bundle, error) {
	// The Bundle type is the generated wrapper. Populate VM and Obj.
	// Since we're inside a VM.Do callback, we don't have a VM pointer here,
	// so we just verify the object is valid and return a minimal wrapper.
	if obj == nil || obj.Ref() == 0 {
		return nil, nil
	}
	return &Bundle{Obj: env.NewGlobalRef(obj)}, nil
}
