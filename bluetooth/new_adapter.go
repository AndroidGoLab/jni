package bluetooth

import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
)

// NewAdapter obtains the BluetoothAdapter via BluetoothManager system service.
func NewAdapter(ctx *app.Context) (*Adapter, error) {
	if ctx == nil {
		return nil, fmt.Errorf("bluetooth.NewAdapter: nil Context")
	}
	var adapter Adapter
	adapter.VM = ctx.VM

	err := adapter.VM.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}

		svc, err := ctx.GetSystemService("bluetooth")
		if err != nil {
			return fmt.Errorf("get bluetooth service: %w", err)
		}
		if svc == nil || svc.Ref() == 0 {
			return fmt.Errorf("bluetooth service not available")
		}
		defer env.DeleteGlobalRef(svc)

		bmClass, err := env.FindClass("android/bluetooth/BluetoothManager")
		if err != nil {
			return fmt.Errorf("find BluetoothManager: %w", err)
		}
		defer env.DeleteLocalRef(&bmClass.Object)
		getAdapterMid, err := env.GetMethodID(bmClass, "getAdapter",
			"()Landroid/bluetooth/BluetoothAdapter;")
		if err != nil {
			return fmt.Errorf("get getAdapter: %w", err)
		}
		adapterObj, err := env.CallObjectMethod(svc, getAdapterMid)
		if err != nil {
			return fmt.Errorf("call getAdapter: %w", err)
		}
		if adapterObj == nil || adapterObj.Ref() == 0 {
			return fmt.Errorf("BluetoothAdapter is null (Bluetooth may be disabled)")
		}

		adapter.Obj = env.NewGlobalRef(adapterObj)
		env.DeleteLocalRef(adapterObj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &adapter, nil
}

// Close releases the global reference to the underlying Java object.
func (m *Adapter) Close() {
	if m.Obj != nil {
		_ = m.VM.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(m.Obj)
			m.Obj = nil
			return nil
		})
	}
}

var _ = unsafe.Pointer(nil)
