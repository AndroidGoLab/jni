package build

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

func readStaticStringField(
	env *jni.Env,
	cls *jni.Class,
	name string,
) (string, error) {
	fid, err := env.GetStaticFieldID(cls, name, "Ljava/lang/String;")
	if err != nil {
		return "", fmt.Errorf("get field %s: %w", name, err)
	}
	obj := env.GetStaticObjectField(cls, fid)
	return env.GoString((*jni.String)(unsafe.Pointer(obj))), nil
}

// GetBuildInfo reads static fields from android.os.Build.
func GetBuildInfo(vm *jni.VM) (*BuildInfo, error) {
	var info BuildInfo
	err := vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("android/os/Build")
		if err != nil {
			return fmt.Errorf("find Build: %w", err)
		}

		var errs []error
		info.Device, err = readStaticStringField(env, cls, "DEVICE")
		errs = append(errs, err)
		info.Model, err = readStaticStringField(env, cls, "MODEL")
		errs = append(errs, err)
		info.Product, err = readStaticStringField(env, cls, "PRODUCT")
		errs = append(errs, err)
		info.Manufacturer, err = readStaticStringField(env, cls, "MANUFACTURER")
		errs = append(errs, err)
		info.Brand, err = readStaticStringField(env, cls, "BRAND")
		errs = append(errs, err)
		info.Board, err = readStaticStringField(env, cls, "BOARD")
		errs = append(errs, err)
		info.Hardware, err = readStaticStringField(env, cls, "HARDWARE")
		errs = append(errs, err)
		return errors.Join(errs...)
	})
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetVersionInfo reads static fields from android.os.Build.VERSION.
func GetVersionInfo(vm *jni.VM) (*VersionInfo, error) {
	var info VersionInfo
	err := vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("android/os/Build$VERSION")
		if err != nil {
			return fmt.Errorf("find Build.VERSION: %w", err)
		}

		var errs []error
		info.Release, err = readStaticStringField(env, cls, "RELEASE")
		errs = append(errs, err)
		info.Codename, err = readStaticStringField(env, cls, "CODENAME")
		errs = append(errs, err)
		info.Incremental, err = readStaticStringField(env, cls, "INCREMENTAL")
		errs = append(errs, err)

		sdkFid, err := env.GetStaticFieldID(cls, "SDK_INT", "I")
		if err != nil {
			return fmt.Errorf("get SDK_INT: %w", err)
		}
		info.SDKInt = env.GetStaticIntField(cls, sdkFid)

		return errors.Join(errs...)
	})
	if err != nil {
		return nil, err
	}
	return &info, nil
}
