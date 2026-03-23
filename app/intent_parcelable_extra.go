package app

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

var (
	parcelableExtraOnce sync.Once
	parcelableExtraMid  jni.MethodID
	parcelableExtraErr  error
)

// GetParcelableExtra calls android.content.Intent.getParcelableExtra(String).
// This method is deprecated in API 33 but remains the standard way to extract
// Parcelable extras on older API levels.
func (m *Intent) GetParcelableExtra(
	key string,
) (*jni.Object, error) {
	var result *jni.Object
	var callErr error
	callErr = m.VM.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}

		parcelableExtraOnce.Do(func() {
			parcelableExtraMid, parcelableExtraErr = env.GetMethodID(
				(*jni.Class)(unsafe.Pointer(clsIntent)),
				"getParcelableExtra",
				"(Ljava/lang/String;)Landroid/os/Parcelable;",
			)
		})
		if parcelableExtraErr != nil {
			callErr = fmt.Errorf("android.content.Intent.getParcelableExtra is not available on this device")
			return callErr
		}

		jKey, err := env.NewStringUTF(key)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jKey.Object)

		result, callErr = env.CallObjectMethod(
			m.Obj,
			parcelableExtraMid, jni.ObjectValue(&jKey.Object),
		)
		if callErr != nil {
			return callErr
		}
		if result != nil {
			localRef := result
			result = env.NewGlobalRef(localRef)
			env.DeleteLocalRef(localRef)
		}
		return callErr
	})
	return result, callErr
}
