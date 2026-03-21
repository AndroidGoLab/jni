package app

import "github.com/AndroidGoLab/jni"

// ContextFromObject wraps an existing Android Context JNI global reference
// into an app.Context. It creates its own GlobalRef so that calling Close()
// is safe and will not affect the caller's reference.
func ContextFromObject(vm *jni.VM, obj *jni.GlobalRef) (*Context, error) {
	ctx := &Context{VM: vm}
	err := vm.Do(func(env *jni.Env) error {
		ctx.Obj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ctx, nil
}
