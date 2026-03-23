package os

import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

// NewHandlerThread creates and starts a new HandlerThread with the given name.
// The thread is started immediately; call GetLooper() to obtain its Looper
// for registering callbacks.
func NewHandlerThread(
	vm *jni.VM,
	name string,
) (*HandlerThread, error) {
	if vm == nil {
		return nil, fmt.Errorf("os.NewHandlerThread: nil VM")
	}
	var ht HandlerThread
	ht.VM = vm

	err := ht.VM.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}

		cls := (*jni.Class)(unsafe.Pointer(clsHandlerThread))

		initMid, err := env.GetMethodID(cls, "<init>", "(Ljava/lang/String;)V")
		if err != nil {
			return fmt.Errorf("get HandlerThread(String) constructor: %w", err)
		}
		jName, err := env.NewStringUTF(name)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jName.Object)

		obj, err := env.NewObject(cls, initMid, jni.ObjectValue(&jName.Object))
		if err != nil {
			return fmt.Errorf("new HandlerThread(%q): %w", name, err)
		}

		ht.Obj = env.NewGlobalRef(obj)
		env.DeleteLocalRef(obj)

		// Start the thread immediately. Thread.start() is inherited from
		// java.lang.Thread and not in the generated method set.
		startMid, err := env.GetMethodID(cls, "start", "()V")
		if err != nil {
			return fmt.Errorf("get HandlerThread.start: %w", err)
		}
		return env.CallVoidMethod(ht.Obj, startMid)
	})
	if err != nil {
		return nil, err
	}
	return &ht, nil
}

// Close quits the thread safely and releases the global reference.
func (m *HandlerThread) Close() {
	if m.Obj == nil {
		return
	}
	// QuitSafely calls ensureInit and vm.Do internally.
	_, _ = m.QuitSafely()
	_ = m.VM.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(m.Obj)
		m.Obj = nil
		return nil
	})
}
