package os

import (
	"github.com/AndroidGoLab/jni"
)

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
