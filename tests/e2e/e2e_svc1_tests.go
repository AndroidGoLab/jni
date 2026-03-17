//go:build android

package main

import (
	"fmt"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/bluetooth"
	"github.com/AndroidGoLab/jni/content/clipboard"
	"github.com/AndroidGoLab/jni/content/pm"
	"github.com/AndroidGoLab/jni/net/wifi"
	"github.com/AndroidGoLab/jni/os/keyguard"
	"github.com/AndroidGoLab/jni/os/power"
	"github.com/AndroidGoLab/jni/os/storage"
	"github.com/AndroidGoLab/jni/os/vibrator"
	"github.com/AndroidGoLab/jni/telephony"
	"github.com/AndroidGoLab/jni/view/display"
)

func testBluetoothWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	adapter, err := bluetooth.NewAdapter(ctx)
	if err != nil {
		return fmt.Errorf("new adapter: %w", err)
	}
	defer adapter.Close()

	_, err = adapter.IsEnabled()
	if err != nil {
		return fmt.Errorf("isEnabled: %w", err)
	}
	return nil
}

func testWifiWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := wifi.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new manager: %w", err)
	}
	defer mgr.Close()

	_, err = mgr.IsWifiEnabled()
	if err != nil {
		return fmt.Errorf("isWifiEnabled: %w", err)
	}
	return nil
}

func testTelephonyWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// All TelephonyManager methods are unexported, so we can only verify
	// that the system service is obtained successfully.
	mgr, err := telephony.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new manager: %w", err)
	}
	defer deleteGlobalRef(vm, mgr.Obj)
	return nil
}

func testPmWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// PackageManager is obtained via Context.PackageManager(), not a NewManager.
	pmObj, err := ctx.PackageManager()
	if err != nil {
		return fmt.Errorf("get package manager: %w", err)
	}

	var pmGlobal *jni.GlobalRef
	err = vm.Do(func(env *jni.Env) error {
		pmGlobal = env.NewGlobalRef(pmObj)
		return nil
	})
	if err != nil {
		return fmt.Errorf("new global ref: %w", err)
	}
	defer vm.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(pmGlobal)
		return nil
	})

	mgr := &pm.Manager{
		VM:  vm,
		Obj: pmGlobal,
	}

	_, err = mgr.HasSystemFeature("android.hardware.touchscreen")
	if err != nil {
		return fmt.Errorf("hasSystemFeature: %w", err)
	}
	return nil
}

func testPowerWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := power.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new manager: %w", err)
	}
	defer deleteGlobalRef(vm, mgr.Obj)

	_, err = mgr.IsInteractive()
	if err != nil {
		return fmt.Errorf("isInteractive: %w", err)
	}

	_, err = mgr.IsPowerSaveMode()
	if err != nil {
		return fmt.Errorf("isPowerSaveMode: %w", err)
	}
	return nil
}

func testVibratorWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	vib, err := vibrator.NewVibrator(ctx)
	if err != nil {
		return fmt.Errorf("new vibrator: %w", err)
	}
	defer deleteGlobalRef(vm, vib.Obj)

	_, err = vib.HasVibrator()
	if err != nil {
		return fmt.Errorf("hasVibrator: %w", err)
	}
	return nil
}

func testClipboardWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// All ClipboardManager query methods are unexported, so we can only
	// verify that the system service is obtained successfully.
	mgr, err := clipboard.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new manager: %w", err)
	}
	defer mgr.Close()
	return nil
}

func testKeyguardWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	mgr, err := keyguard.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new manager: %w", err)
	}
	defer mgr.Close()

	_, err = mgr.IsKeyguardLocked()
	if err != nil {
		return fmt.Errorf("isKeyguardLocked: %w", err)
	}

	_, err = mgr.IsDeviceSecure()
	if err != nil {
		return fmt.Errorf("isDeviceSecure: %w", err)
	}
	return nil
}

func testDisplayWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// The windowManager type and its methods are unexported, so we can only
	// verify that the system service is obtained successfully.
	mgr, err := display.NewwindowManager(ctx)
	if err != nil {
		return fmt.Errorf("new window manager: %w", err)
	}
	defer deleteGlobalRef(vm, mgr.Obj)
	return nil
}

func testStorageWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	defer ctx.Close()

	// All StorageManager methods are unexported, so we can only verify
	// that the system service is obtained successfully.
	mgr, err := storage.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new manager: %w", err)
	}
	defer deleteGlobalRef(vm, mgr.Obj)
	return nil
}
