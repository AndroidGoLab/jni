package build

import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

// BuildInfo holds static field values from android.os.Build.
type BuildInfo struct {
	Device       string
	Model        string
	Product      string
	Manufacturer string
	Brand        string
	Board        string
	Hardware     string
}

// VersionInfo holds static field values from android.os.Build.VERSION.
type VersionInfo struct {
	Release     string
	SDKInt      int32
	Codename    string
	Incremental string
}

// GetBuildInfo reads static fields from android.os.Build.
func GetBuildInfo(vm *jni.VM) (*BuildInfo, error) {
	var info BuildInfo
	err := vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("android/os/Build")
		if err != nil {
			return fmt.Errorf("find Build: %w", err)
		}

		readString := func(name string) (string, error) {
			fid, err := env.GetStaticFieldID(cls, name, "Ljava/lang/String;")
			if err != nil {
				return "", err
			}
			obj := env.GetStaticObjectField(cls, fid)
			return env.GoString((*jni.String)(unsafe.Pointer(obj))), nil
		}

		info.Device, _ = readString("DEVICE")
		info.Model, _ = readString("MODEL")
		info.Product, _ = readString("PRODUCT")
		info.Manufacturer, _ = readString("MANUFACTURER")
		info.Brand, _ = readString("BRAND")
		info.Board, _ = readString("BOARD")
		info.Hardware, _ = readString("HARDWARE")
		return nil
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

		readString := func(name string) (string, error) {
			fid, err := env.GetStaticFieldID(cls, name, "Ljava/lang/String;")
			if err != nil {
				return "", err
			}
			obj := env.GetStaticObjectField(cls, fid)
			return env.GoString((*jni.String)(unsafe.Pointer(obj))), nil
		}

		info.Release, _ = readString("RELEASE")
		info.Codename, _ = readString("CODENAME")
		info.Incremental, _ = readString("INCREMENTAL")

		sdkFid, err := env.GetStaticFieldID(cls, "SDK_INT", "I")
		if err != nil {
			return fmt.Errorf("get SDK_INT: %w", err)
		}
		info.SDKInt = env.GetStaticIntField(cls, sdkFid)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &info, nil
}
