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
		defer env.DeleteLocalRef(&cls.Object)
		mid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("get Intent.<init>: %w", err)
		}
		obj, err := env.NewObject(cls, mid)
		if err != nil {
			return fmt.Errorf("new Intent: %w", err)
		}
		intent.Obj = env.NewGlobalRef(obj)
		env.DeleteLocalRef(obj)
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
) error {
	return i.VM.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(i.Obj)
		defer env.DeleteLocalRef(&cls.Object)
		mid, err := env.GetMethodID(cls, "putExtra",
			"(Ljava/lang/String;Ljava/lang/String;)Landroid/content/Intent;")
		if err != nil {
			return err
		}
		jKey, err := env.NewStringUTF(key)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jKey.Object)
		jVal, err := env.NewStringUTF(value)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jVal.Object)
		retObj, err := env.CallObjectMethod(i.Obj, mid,
			jni.ObjectValue(&jKey.Object), jni.ObjectValue(&jVal.Object))
		if retObj != nil {
			env.DeleteLocalRef(retObj)
		}
		return err
	})
}

// PutIntExtra calls Intent.putExtra(key, value) for int values.
func (i *Intent) PutIntExtra(
	key string,
	value int32,
) error {
	return i.VM.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(i.Obj)
		defer env.DeleteLocalRef(&cls.Object)
		mid, err := env.GetMethodID(cls, "putExtra",
			"(Ljava/lang/String;I)Landroid/content/Intent;")
		if err != nil {
			return err
		}
		jKey, err := env.NewStringUTF(key)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jKey.Object)
		retObj, err := env.CallObjectMethod(i.Obj, mid,
			jni.ObjectValue(&jKey.Object), jni.IntValue(value))
		if retObj != nil {
			env.DeleteLocalRef(retObj)
		}
		return err
	})
}

// PutBoolExtra calls Intent.putExtra(key, value) for boolean values.
func (i *Intent) PutBoolExtra(
	key string,
	value bool,
) error {
	return i.VM.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(i.Obj)
		defer env.DeleteLocalRef(&cls.Object)
		mid, err := env.GetMethodID(cls, "putExtra",
			"(Ljava/lang/String;Z)Landroid/content/Intent;")
		if err != nil {
			return err
		}
		jKey, err := env.NewStringUTF(key)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jKey.Object)
		var boolVal uint8
		if value {
			boolVal = 1
		}
		retObj, err := env.CallObjectMethod(i.Obj, mid,
			jni.ObjectValue(&jKey.Object), jni.BooleanValue(boolVal))
		if retObj != nil {
			env.DeleteLocalRef(retObj)
		}
		return err
	})
}

// GetBoolExtra calls Intent.getBooleanExtra(key, defaultValue).
func (i *Intent) GetBoolExtra(
	key string,
	defaultValue bool,
) (bool, error) {
	return i.GetBooleanExtra(key, defaultValue)
}
