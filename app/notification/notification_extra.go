package notification

import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

// NewChannel creates a new android.app.NotificationChannel(id, name, importance).
func NewChannel(vm *jni.VM, id, name string, importance int32) (*Channel, error) {
	var ch Channel
	ch.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}
		mid, err := env.GetMethodID(
			(*jni.Class)(unsafe.Pointer(clsChannel)),
			"<init>",
			"(Ljava/lang/String;Ljava/lang/CharSequence;I)V",
		)
		if err != nil {
			return fmt.Errorf("get NotificationChannel.<init>: %w", err)
		}
		jID, err := env.NewStringUTF(id)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jID.Object)
		jName, err := env.NewStringUTF(name)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jName.Object)
		obj, err := env.NewObject(
			(*jni.Class)(unsafe.Pointer(clsChannel)),
			mid,
			jni.ObjectValue(&jID.Object),
			jni.ObjectValue(&jName.Object),
			jni.IntValue(importance),
		)
		if err != nil {
			return fmt.Errorf("new NotificationChannel: %w", err)
		}
		ch.Obj = env.NewGlobalRef(obj)
		env.DeleteLocalRef(obj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// NewBuilder creates a new android.app.Notification$Builder(context, channelID).
func NewBuilder(vm *jni.VM, context *jni.Object, channelID string) (*Builder, error) {
	var b Builder
	b.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := ensureInit(env); err != nil {
			return err
		}
		mid, err := env.GetMethodID(
			(*jni.Class)(unsafe.Pointer(clsBuilder)),
			"<init>",
			"(Landroid/content/Context;Ljava/lang/String;)V",
		)
		if err != nil {
			return fmt.Errorf("get Notification$Builder.<init>: %w", err)
		}
		jChannelID, err := env.NewStringUTF(channelID)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jChannelID.Object)
		obj, err := env.NewObject(
			(*jni.Class)(unsafe.Pointer(clsBuilder)),
			mid,
			jni.ObjectValue(context),
			jni.ObjectValue(&jChannelID.Object),
		)
		if err != nil {
			return fmt.Errorf("new Notification$Builder: %w", err)
		}
		b.Obj = env.NewGlobalRef(obj)
		env.DeleteLocalRef(obj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &b, nil
}
