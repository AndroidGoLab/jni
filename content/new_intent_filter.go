package content

import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

// NewIntentFilter creates a new IntentFilter with the given action.
// Pass an empty string to create a filter with no initial action.
func NewIntentFilter(
	vm *jni.VM,
	action string,
) (*IntentFilter, error) {
	if vm == nil {
		return nil, fmt.Errorf("content.NewIntentFilter: nil VM")
	}
	var filter IntentFilter
	filter.VM = vm

	err := filter.VM.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}

		cls := (*jni.Class)(unsafe.Pointer(clsIntentFilter))

		var obj *jni.Object
		switch {
		case action != "":
			initMid, err := env.GetMethodID(cls, "<init>", "(Ljava/lang/String;)V")
			if err != nil {
				return fmt.Errorf("get IntentFilter(String) constructor: %w", err)
			}
			jAction, err := env.NewStringUTF(action)
			if err != nil {
				return err
			}
			defer env.DeleteLocalRef(&jAction.Object)
			obj, err = env.NewObject(cls, initMid, jni.ObjectValue(&jAction.Object))
			if err != nil {
				return fmt.Errorf("new IntentFilter(%q): %w", action, err)
			}
		default:
			initMid, err := env.GetMethodID(cls, "<init>", "()V")
			if err != nil {
				return fmt.Errorf("get IntentFilter() constructor: %w", err)
			}
			obj, err = env.NewObject(cls, initMid)
			if err != nil {
				return fmt.Errorf("new IntentFilter(): %w", err)
			}
		}

		filter.Obj = env.NewGlobalRef(obj)
		env.DeleteLocalRef(obj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &filter, nil
}

// Close releases the global reference to the underlying Java object.
func (m *IntentFilter) Close() {
	if m.Obj != nil {
		_ = m.VM.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(m.Obj)
			m.Obj = nil
			return nil
		})
	}
}
