package omapi

import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

// NewService creates a new SEService instance via its no-arg constructor.
// The SEService(Context, Executor, OnConnectedListener) constructor is
// required on newer APIs; this uses a simplified no-arg path available
// in app_process context.
func NewService(vm *jni.VM) (*Service, error) {
	var svc Service
	svc.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}

		cls, err := env.FindClass("android/se/omapi/SEService")
		if err != nil {
			return fmt.Errorf("find SEService: %w", err)
		}
		initMid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("get SEService.<init>: %w", err)
		}
		obj, err := env.NewObject(cls, initMid)
		if err != nil {
			return fmt.Errorf("new SEService: %w", err)
		}
		svc.Obj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// Close releases the global reference to the underlying Java object.
func (m *Service) Close() {
	if m.Obj != nil {
		_ = m.VM.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(m.Obj)
			m.Obj = nil
			return nil
		})
	}
}

var _ = unsafe.Pointer(nil)
