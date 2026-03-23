package app

import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

// midIntentGetParcelableExtra is lazily initialized on first call.
var midIntentGetParcelableExtra jni.MethodID

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

		if midIntentGetParcelableExtra == nil {
			mid, err := env.GetMethodID(
				(*jni.Class)(unsafe.Pointer(clsIntent)),
				"getParcelableExtra",
				"(Ljava/lang/String;)Landroid/os/Parcelable;",
			)
			if err != nil {
				callErr = fmt.Errorf("android.content.Intent.getParcelableExtra is not available on this device")
				return callErr
			}
			midIntentGetParcelableExtra = mid
		}

		jKey, err := env.NewStringUTF(key)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jKey.Object)

		result, callErr = env.CallObjectMethod(
			m.Obj,
			midIntentGetParcelableExtra, jni.ObjectValue(&jKey.Object),
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
