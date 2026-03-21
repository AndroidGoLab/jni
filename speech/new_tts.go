package speech

import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

// NewTTS creates a new TextToSpeech instance via the no-arg constructor
// available in app_process context.
func NewTTS(vm *jni.VM) (*TextToSpeech, error) {
	var tts TextToSpeech
	tts.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}

		cls, err := env.FindClass("android/speech/tts/TextToSpeech")
		if err != nil {
			return fmt.Errorf("find TextToSpeech: %w", err)
		}
		defer env.DeleteLocalRef(&cls.Object)
		initMid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("get TextToSpeech.<init>: %w", err)
		}
		obj, err := env.NewObject(cls, initMid)
		if err != nil {
			return fmt.Errorf("new TextToSpeech: %w", err)
		}
		tts.Obj = env.NewGlobalRef(obj)
		env.DeleteLocalRef(obj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &tts, nil
}

// Close releases the global reference to the underlying Java object.
func (m *TextToSpeech) Close() {
	if m.Obj != nil {
		_ = m.VM.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(m.Obj)
			m.Obj = nil
			return nil
		})
	}
}

var _ = unsafe.Pointer(nil)
