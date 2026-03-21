package pdf

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

// NewCanvas creates a new android.graphics.Canvas(bitmap).
func NewCanvas(vm *jni.VM, bitmap *jni.Object) (*Canvas, error) {
	var c Canvas
	c.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}
		mid, err := env.GetMethodID(
			(*jni.Class)(unsafe.Pointer(clsCanvas)),
			"<init>",
			"(Landroid/graphics/Bitmap;)V",
		)
		if err != nil {
			return fmt.Errorf("get Canvas.<init>: %w", err)
		}
		obj, err := env.NewObject(
			(*jni.Class)(unsafe.Pointer(clsCanvas)),
			mid,
			jni.ObjectValue(bitmap),
		)
		if err != nil {
			return fmt.Errorf("new Canvas: %w", err)
		}
		c.Obj = env.NewGlobalRef(obj)
		env.DeleteLocalRef(obj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// NewPaint creates a new android.graphics.Paint() with no arguments.
func NewPaint(vm *jni.VM) (*Paint, error) {
	var p Paint
	p.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}
		mid, err := env.GetMethodID(
			(*jni.Class)(unsafe.Pointer(clsPaint)),
			"<init>",
			"()V",
		)
		if err != nil {
			return fmt.Errorf("get Paint.<init>: %w", err)
		}
		obj, err := env.NewObject(
			(*jni.Class)(unsafe.Pointer(clsPaint)),
			mid,
		)
		if err != nil {
			return fmt.Errorf("new Paint: %w", err)
		}
		p.Obj = env.NewGlobalRef(obj)
		env.DeleteLocalRef(obj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

var (
	argb8888Once sync.Once
	argb8888Obj  *jni.Object
	argb8888Err  error
)

// ARGB8888 returns the Bitmap.Config.ARGB_8888 enum value.
// The result is cached as a global ref and reused across calls.
func ARGB8888(vm *jni.VM) (*jni.Object, error) {
	argb8888Once.Do(func() {
		argb8888Err = vm.Do(func(env *jni.Env) error {
			if err := ensureInit(env); err != nil {
				return err
			}
			fid, err := env.GetStaticFieldID(
				(*jni.Class)(unsafe.Pointer(clsBitmapConfig)),
				"ARGB_8888",
				"Landroid/graphics/Bitmap$Config;",
			)
			if err != nil {
				return fmt.Errorf("get ARGB_8888 field: %w", err)
			}
			obj := env.GetStaticObjectField(
				(*jni.Class)(unsafe.Pointer(clsBitmapConfig)),
				fid,
			)
			if obj != nil {
				argb8888Obj = env.NewGlobalRef(obj)
				env.DeleteLocalRef(obj)
			}
			return nil
		})
	})
	return argb8888Obj, argb8888Err
}

var (
	monospaceOnce sync.Once
	monospaceObj  *jni.Object
	monospaceErr  error
)

// MonospaceTypeface returns the Typeface.MONOSPACE static field.
// The result is cached as a global ref and reused across calls.
func MonospaceTypeface(vm *jni.VM) (*jni.Object, error) {
	monospaceOnce.Do(func() {
		monospaceErr = vm.Do(func(env *jni.Env) error {
			if err := ensureInit(env); err != nil {
				return err
			}
			fid, err := env.GetStaticFieldID(
				(*jni.Class)(unsafe.Pointer(clsTypeface)),
				"MONOSPACE",
				"Landroid/graphics/Typeface;",
			)
			if err != nil {
				return fmt.Errorf("get MONOSPACE field: %w", err)
			}
			obj := env.GetStaticObjectField(
				(*jni.Class)(unsafe.Pointer(clsTypeface)),
				fid,
			)
			if obj != nil {
				monospaceObj = env.NewGlobalRef(obj)
				env.DeleteLocalRef(obj)
			}
			return nil
		})
	})
	return monospaceObj, monospaceErr
}
