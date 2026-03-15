//go:build android

package main

import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/app/alarm"
)

// testAppBundle creates an android.os.Bundle via JNI, puts string/int/bool
// values into it, reads them back, and verifies the round-trip. It also
// validates that app.ExtractBundle succeeds on the live object.
func testAppBundle(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return fmt.Errorf("app.Init: %w", err)
		}

		// Find android.os.Bundle and its constructor.
		bundleCls, err := env.FindClass("android/os/Bundle")
		if err != nil {
			return fmt.Errorf("FindClass Bundle: %w", err)
		}
		initMid, err := env.GetMethodID(bundleCls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("GetMethodID Bundle.<init>: %w", err)
		}
		bundle, err := env.NewObject(bundleCls, initMid)
		if err != nil {
			return fmt.Errorf("NewObject Bundle: %w", err)
		}

		// Resolve put/get method IDs.
		putStringMid, err := env.GetMethodID(bundleCls, "putString", "(Ljava/lang/String;Ljava/lang/String;)V")
		if err != nil {
			return fmt.Errorf("GetMethodID putString: %w", err)
		}
		getStringMid, err := env.GetMethodID(bundleCls, "getString", "(Ljava/lang/String;)Ljava/lang/String;")
		if err != nil {
			return fmt.Errorf("GetMethodID getString: %w", err)
		}
		putIntMid, err := env.GetMethodID(bundleCls, "putInt", "(Ljava/lang/String;I)V")
		if err != nil {
			return fmt.Errorf("GetMethodID putInt: %w", err)
		}
		getIntMid, err := env.GetMethodID(bundleCls, "getInt", "(Ljava/lang/String;)I")
		if err != nil {
			return fmt.Errorf("GetMethodID getInt: %w", err)
		}
		putBoolMid, err := env.GetMethodID(bundleCls, "putBoolean", "(Ljava/lang/String;Z)V")
		if err != nil {
			return fmt.Errorf("GetMethodID putBoolean: %w", err)
		}
		getBoolMid, err := env.GetMethodID(bundleCls, "getBoolean", "(Ljava/lang/String;)Z")
		if err != nil {
			return fmt.Errorf("GetMethodID getBoolean: %w", err)
		}

		// Put a string and verify round-trip.
		{
			jKey, err := env.NewStringUTF("greeting")
			if err != nil {
				return err
			}
			jVal, err := env.NewStringUTF("hello from Go")
			if err != nil {
				return err
			}
			if err := env.CallVoidMethod(bundle, putStringMid,
				jni.ObjectValue(&jKey.Object), jni.ObjectValue(&jVal.Object)); err != nil {
				return fmt.Errorf("putString: %w", err)
			}
			jKey2, err := env.NewStringUTF("greeting")
			if err != nil {
				return err
			}
			gotObj, err := env.CallObjectMethod(bundle, getStringMid,
				jni.ObjectValue(&jKey2.Object))
			if err != nil {
				return fmt.Errorf("getString: %w", err)
			}
			got := env.GoString((*jni.String)(unsafe.Pointer(gotObj)))
			if got != "hello from Go" {
				return fmt.Errorf("getString = %q, want %q", got, "hello from Go")
			}
		}

		// Put an int and verify round-trip.
		{
			jKey, err := env.NewStringUTF("count")
			if err != nil {
				return err
			}
			if err := env.CallVoidMethod(bundle, putIntMid,
				jni.ObjectValue(&jKey.Object), jni.IntValue(42)); err != nil {
				return fmt.Errorf("putInt: %w", err)
			}
			jKey2, err := env.NewStringUTF("count")
			if err != nil {
				return err
			}
			gotInt, err := env.CallIntMethod(bundle, getIntMid,
				jni.ObjectValue(&jKey2.Object))
			if err != nil {
				return fmt.Errorf("getInt: %w", err)
			}
			if gotInt != 42 {
				return fmt.Errorf("getInt = %d, want 42", gotInt)
			}
		}

		// Put a bool and verify round-trip.
		{
			jKey, err := env.NewStringUTF("enabled")
			if err != nil {
				return err
			}
			if err := env.CallVoidMethod(bundle, putBoolMid,
				jni.ObjectValue(&jKey.Object), jni.BooleanValue(1)); err != nil {
				return fmt.Errorf("putBoolean: %w", err)
			}
			jKey2, err := env.NewStringUTF("enabled")
			if err != nil {
				return err
			}
			gotBool, err := env.CallBooleanMethod(bundle, getBoolMid,
				jni.ObjectValue(&jKey2.Object))
			if err != nil {
				return fmt.Errorf("getBoolean: %w", err)
			}
			if gotBool == 0 {
				return fmt.Errorf("getBoolean = false, want true")
			}
		}

		// Validate that ExtractBundle succeeds on the live object.
		_, err = app.ExtractBundle(env, bundle)
		if err != nil {
			return fmt.Errorf("ExtractBundle: %w", err)
		}

		return nil
	})
}

// testAlarmClockInfo creates an AlarmManager.AlarmClockInfo via JNI
// constructor(long triggerTime, PendingIntent showIntent), then extracts
// fields via alarm.ExtractalarmClockInfo and verifies the trigger time.
func testAlarmClockInfo(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		if err := alarm.Init(env); err != nil {
			return fmt.Errorf("alarm.Init: %w", err)
		}

		// Find AlarmClockInfo and its two-arg constructor.
		cls, err := env.FindClass("android/app/AlarmManager$AlarmClockInfo")
		if err != nil {
			return fmt.Errorf("FindClass AlarmClockInfo: %w", err)
		}
		initMid, err := env.GetMethodID(cls, "<init>", "(JLandroid/app/PendingIntent;)V")
		if err != nil {
			return fmt.Errorf("GetMethodID AlarmClockInfo.<init>: %w", err)
		}

		// Create with a known trigger time and null PendingIntent.
		var wantTrigger int64 = 1700000000000
		obj, err := env.NewObject(cls, initMid,
			jni.LongValue(wantTrigger), jni.ObjectValue(nil))
		if err != nil {
			return fmt.Errorf("NewObject AlarmClockInfo: %w", err)
		}

		// Extract via the generated extraction function.
		info, err := alarm.ExtractalarmClockInfo(env, obj)
		if err != nil {
			return fmt.Errorf("ExtractalarmClockInfo: %w", err)
		}
		if info.TriggerTime != wantTrigger {
			return fmt.Errorf("TriggerTime = %d, want %d", info.TriggerTime, wantTrigger)
		}

		return nil
	})
}

// testKeystoreKeyStore obtains a java.security.KeyStore instance for
// "AndroidKeyStore", loads it, and verifies containsAlias works via raw
// JNI calls (the generated wrapper types are unexported).
func testKeystoreKeyStore(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		// Get KeyStore.getInstance("AndroidKeyStore") via static method.
		ksCls, err := env.FindClass("java/security/KeyStore")
		if err != nil {
			return fmt.Errorf("FindClass KeyStore: %w", err)
		}
		getInstanceMid, err := env.GetStaticMethodID(ksCls, "getInstance",
			"(Ljava/lang/String;)Ljava/security/KeyStore;")
		if err != nil {
			return fmt.Errorf("GetStaticMethodID getInstance: %w", err)
		}
		provStr, err := env.NewStringUTF("AndroidKeyStore")
		if err != nil {
			return err
		}
		ksObj, err := env.CallStaticObjectMethod(ksCls, getInstanceMid,
			jni.ObjectValue(&provStr.Object))
		if err != nil {
			return fmt.Errorf("KeyStore.getInstance: %w", err)
		}
		if ksObj == nil {
			return fmt.Errorf("KeyStore.getInstance returned nil")
		}

		// Load the keystore with null parameter.
		loadMid, err := env.GetMethodID(ksCls, "load",
			"(Ljava/security/KeyStore$LoadStoreParameter;)V")
		if err != nil {
			return fmt.Errorf("GetMethodID load: %w", err)
		}
		if err := env.CallVoidMethod(ksObj, loadMid, jni.ObjectValue(nil)); err != nil {
			return fmt.Errorf("KeyStore.load: %w", err)
		}

		// Call containsAlias with a key that almost certainly does not exist.
		containsMid, err := env.GetMethodID(ksCls, "containsAlias",
			"(Ljava/lang/String;)Z")
		if err != nil {
			return fmt.Errorf("GetMethodID containsAlias: %w", err)
		}
		aliasStr, err := env.NewStringUTF("e2e_nonexistent_key_12345")
		if err != nil {
			return err
		}
		found, err := env.CallBooleanMethod(ksObj, containsMid,
			jni.ObjectValue(&aliasStr.Object))
		if err != nil {
			return fmt.Errorf("containsAlias: %w", err)
		}
		if found != 0 {
			return fmt.Errorf("containsAlias unexpectedly returned true for nonexistent key")
		}

		return nil
	})
}

// testNotificationChannel creates an android.app.NotificationChannel via
// its three-arg JNI constructor (id, name, importance), then reads back
// id, description, and importance via raw JNI to verify the round-trip.
func testNotificationChannel(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("android/app/NotificationChannel")
		if err != nil {
			return fmt.Errorf("FindClass NotificationChannel: %w", err)
		}

		// NotificationChannel(String id, CharSequence name, int importance)
		initMid, err := env.GetMethodID(cls, "<init>",
			"(Ljava/lang/String;Ljava/lang/CharSequence;I)V")
		if err != nil {
			return fmt.Errorf("GetMethodID NotificationChannel.<init>: %w", err)
		}

		jID, err := env.NewStringUTF("e2e_test_chan")
		if err != nil {
			return err
		}
		jName, err := env.NewStringUTF("E2E Test Channel")
		if err != nil {
			return err
		}
		const importanceDefault int32 = 3 // NotificationManager.IMPORTANCE_DEFAULT
		obj, err := env.NewObject(cls, initMid,
			jni.ObjectValue(&jID.Object),
			jni.ObjectValue(&jName.Object),
			jni.IntValue(importanceDefault))
		if err != nil {
			return fmt.Errorf("NewObject NotificationChannel: %w", err)
		}

		// Set description.
		setDescMid, err := env.GetMethodID(cls, "setDescription", "(Ljava/lang/String;)V")
		if err != nil {
			return fmt.Errorf("GetMethodID setDescription: %w", err)
		}
		jDesc, err := env.NewStringUTF("test description")
		if err != nil {
			return err
		}
		if err := env.CallVoidMethod(obj, setDescMid, jni.ObjectValue(&jDesc.Object)); err != nil {
			return fmt.Errorf("setDescription: %w", err)
		}

		// Verify getId.
		getIDMid, err := env.GetMethodID(cls, "getId", "()Ljava/lang/String;")
		if err != nil {
			return fmt.Errorf("GetMethodID getId: %w", err)
		}
		idObj, err := env.CallObjectMethod(obj, getIDMid)
		if err != nil {
			return fmt.Errorf("getId: %w", err)
		}
		gotID := env.GoString((*jni.String)(unsafe.Pointer(idObj)))
		if gotID != "e2e_test_chan" {
			return fmt.Errorf("getId = %q, want %q", gotID, "e2e_test_chan")
		}

		// Verify getDescription.
		getDescMid, err := env.GetMethodID(cls, "getDescription", "()Ljava/lang/String;")
		if err != nil {
			return fmt.Errorf("GetMethodID getDescription: %w", err)
		}
		descObj, err := env.CallObjectMethod(obj, getDescMid)
		if err != nil {
			return fmt.Errorf("getDescription: %w", err)
		}
		gotDesc := env.GoString((*jni.String)(unsafe.Pointer(descObj)))
		if gotDesc != "test description" {
			return fmt.Errorf("getDescription = %q, want %q", gotDesc, "test description")
		}

		// Verify getImportance.
		getImpMid, err := env.GetMethodID(cls, "getImportance", "()I")
		if err != nil {
			return fmt.Errorf("GetMethodID getImportance: %w", err)
		}
		gotImp, err := env.CallIntMethod(obj, getImpMid)
		if err != nil {
			return fmt.Errorf("getImportance: %w", err)
		}
		if gotImp != importanceDefault {
			return fmt.Errorf("getImportance = %d, want %d", gotImp, importanceDefault)
		}

		return nil
	})
}
