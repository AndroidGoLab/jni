package app

import "github.com/AndroidGoLab/jni"

// Close releases the GlobalRef held by this Context.
func (m *Context) Close() {
	if m.Obj != nil {
		_ = m.VM.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(m.Obj)
			m.Obj = nil
			return nil
		})
	}
}
