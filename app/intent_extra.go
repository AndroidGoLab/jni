package app

import (
	"fmt"

	"github.com/AndroidGoLab/jni"
)

// NewIntent creates a new android.content.Intent via its no-arg constructor.
func NewIntent(vm *jni.VM) (*Intent, error) {
	var intent Intent
	intent.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}
		cls, err := env.FindClass("android/content/Intent")
		if err != nil {
			return fmt.Errorf("find Intent: %w", err)
		}
		mid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("get Intent.<init>: %w", err)
		}
		obj, err := env.NewObject(cls, mid)
		if err != nil {
			return fmt.Errorf("new Intent: %w", err)
		}
		intent.Obj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &intent, nil
}

// PutStringExtra calls Intent.putExtra(key, value) for string values.
func (i *Intent) PutStringExtra(
	key string,
	value string,
) {
	i.VM.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(i.Obj)
		mid, err := env.GetMethodID(cls, "putExtra",
			"(Ljava/lang/String;Ljava/lang/String;)Landroid/content/Intent;")
		if err != nil {
			return err
		}
		jKey, err := env.NewStringUTF(key)
		if err != nil {
			return err
		}
		jVal, err := env.NewStringUTF(value)
		if err != nil {
			return err
		}
		_, err = env.CallObjectMethod(i.Obj, mid,
			jni.ObjectValue(&jKey.Object), jni.ObjectValue(&jVal.Object))
		return err
	})
}

// PutIntExtra calls Intent.putExtra(key, value) for int values.
func (i *Intent) PutIntExtra(
	key string,
	value int32,
) {
	i.VM.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(i.Obj)
		mid, err := env.GetMethodID(cls, "putExtra",
			"(Ljava/lang/String;I)Landroid/content/Intent;")
		if err != nil {
			return err
		}
		jKey, err := env.NewStringUTF(key)
		if err != nil {
			return err
		}
		_, err = env.CallObjectMethod(i.Obj, mid,
			jni.ObjectValue(&jKey.Object), jni.IntValue(value))
		return err
	})
}

// PutBoolExtra calls Intent.putExtra(key, value) for boolean values.
func (i *Intent) PutBoolExtra(
	key string,
	value bool,
) {
	i.VM.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(i.Obj)
		mid, err := env.GetMethodID(cls, "putExtra",
			"(Ljava/lang/String;Z)Landroid/content/Intent;")
		if err != nil {
			return err
		}
		jKey, err := env.NewStringUTF(key)
		if err != nil {
			return err
		}
		var boolVal uint8
		if value {
			boolVal = 1
		}
		_, err = env.CallObjectMethod(i.Obj, mid,
			jni.ObjectValue(&jKey.Object), jni.BooleanValue(boolVal))
		return err
	})
}

// GetBoolExtra calls Intent.getBooleanExtra(key, defaultValue).
func (i *Intent) GetBoolExtra(
	key string,
	defaultValue bool,
) bool {
	result, _ := i.GetBooleanExtra(key, defaultValue)
	return result
}
