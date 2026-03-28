package content

import (
	"github.com/AndroidGoLab/jni"
)

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
