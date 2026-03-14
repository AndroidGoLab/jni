//go:build android

package main

import (
	"fmt"
	"os"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/accounts"
	"github.com/xaionaro-go/jni/app/download"
	"github.com/xaionaro-go/jni/app/job"
	"github.com/xaionaro-go/jni/app/notification"
	"github.com/xaionaro-go/jni/app/usage"
	"github.com/xaionaro-go/jni/companion"
	"github.com/xaionaro-go/jni/content/preferences"
	"github.com/xaionaro-go/jni/content/resolver"
	"github.com/xaionaro-go/jni/graphics/pdf"
	"github.com/xaionaro-go/jni/hardware/biometric"
	"github.com/xaionaro-go/jni/hardware/camera"
	"github.com/xaionaro-go/jni/hardware/ir"
	"github.com/xaionaro-go/jni/hardware/lights"
	"github.com/xaionaro-go/jni/hardware/usb"
	"github.com/xaionaro-go/jni/media/audiomanager"
	"github.com/xaionaro-go/jni/media/player"
	"github.com/xaionaro-go/jni/media/projection"
	"github.com/xaionaro-go/jni/media/recorder"
	"github.com/xaionaro-go/jni/media/session"
	"github.com/xaionaro-go/jni/net"
	"github.com/xaionaro-go/jni/net/nsd"
	"github.com/xaionaro-go/jni/net/wifi/p2p"
	"github.com/xaionaro-go/jni/net/wifi/rtt"
	"github.com/xaionaro-go/jni/nfc"
	"github.com/xaionaro-go/jni/print"
	"github.com/xaionaro-go/jni/provider/documents"
	"github.com/xaionaro-go/jni/provider/media"
	"github.com/xaionaro-go/jni/se/omapi"
	"github.com/xaionaro-go/jni/speech"
	"github.com/xaionaro-go/jni/telecom"
	"github.com/xaionaro-go/jni/view/inputmethod"
	"github.com/xaionaro-go/jni/widget/toast"
)

// =====================================================================
// Service wrapper functional tests (packages requiring Android Context
// or constructors). Each test obtains the manager/service, optionally
// calls an exported read-only method, then cleans up.
// =====================================================================

// deleteGlobalRef releases a JNI global reference. Used for types that
// lack a Close() method.
func deleteGlobalRef(vm *jni.VM, ref *jni.GlobalRef) {
	if ref != nil {
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(ref)
			return nil
		})
	}
}

func testAudioManagerWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := audiomanager.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new audiomanager: %w", err)
	}
	defer mgr.Close()

	// All query methods (getStreamVolume etc.) are unexported.
	// Verify the manager is obtained without error.
	return nil
}

func testBiometricWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := biometric.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new biometric manager: %w", err)
	}
	defer mgr.Close()

	return nil
}

func testCameraWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := camera.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new camera manager: %w", err)
	}
	defer mgr.Close()

	return nil
}

func testCompanionWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := companion.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new companion manager: %w", err)
	}
	defer deleteGlobalRef(vm, mgr.Obj)

	return nil
}

func testDownloadWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := download.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new download manager: %w", err)
	}
	defer mgr.Close()

	return nil
}

func testInputMethodWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := inputmethod.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new inputmethod manager: %w", err)
	}
	defer mgr.Close()

	return nil
}

func testIrWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := ir.NewManager(ctx)
	if err != nil {
		// IR service may not exist on emulator -- non-fatal.
		fmt.Fprintf(os.Stderr, "  IR manager (non-fatal): %v\n", err)
		return nil
	}
	defer deleteGlobalRef(vm, mgr.Obj)

	hasIR, err := mgr.HasIrEmitter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  hasIrEmitter (non-fatal): %v\n", err)
		return nil
	}
	fmt.Fprintf(os.Stderr, "  hasIrEmitter = %v\n", hasIR)
	return nil
}

func testJobSchedulerWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	sched, err := job.NewScheduler(ctx)
	if err != nil {
		return fmt.Errorf("new job scheduler: %w", err)
	}
	defer sched.Close()

	return nil
}

func testLightsWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := lights.NewManager(ctx)
	if err != nil {
		// LightsManager may not be available on older API levels.
		fmt.Fprintf(os.Stderr, "  lights manager (non-fatal): %v\n", err)
		return nil
	}
	defer deleteGlobalRef(vm, mgr.Obj)

	return nil
}

func testNetWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := net.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new connectivity manager: %w", err)
	}
	defer mgr.Close()

	return nil
}

func testNotificationWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := notification.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new notification manager: %w", err)
	}
	defer mgr.Close()

	return nil
}

func testNsdWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := nsd.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new nsd manager: %w", err)
	}
	defer mgr.Close()

	return nil
}

func testOmapiWrapper(vm *jni.VM) error {
	svc, err := omapi.NewService(vm)
	if err != nil {
		// SEService no-arg constructor may fail on emulator -- non-fatal.
		fmt.Fprintf(os.Stderr, "  omapi service (non-fatal): %v\n", err)
		return nil
	}
	defer svc.Close()

	connected := svc.IsConnected()
	fmt.Fprintf(os.Stderr, "  omapi isConnected = %v\n", connected)
	return nil
}

func testPlayerWrapper(vm *jni.VM) error {
	p, err := player.NewPlayer(vm)
	if err != nil {
		return fmt.Errorf("new player: %w", err)
	}
	defer deleteGlobalRef(vm, p.Obj)

	// Player created successfully. All query methods are unexported.
	return nil
}

func testPrintWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := print.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new print manager: %w", err)
	}
	defer deleteGlobalRef(vm, mgr.Obj)

	return nil
}

func testProjectionWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := projection.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new projection manager: %w", err)
	}
	defer mgr.Close()

	return nil
}

func testRecorderWrapper(vm *jni.VM) error {
	// MediaRecorder() constructor crashes JVM in app_process context
	// (native media libraries not loaded). Verify Init only.
	return vm.Do(func(env *jni.Env) error {
		return recorder.Init(env)
	})
}

func testSessionWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := session.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new session manager: %w", err)
	}
	defer mgr.Close()

	return nil
}

func testSpeechTTSWrapper(vm *jni.VM) error {
	tts, err := speech.NewTTS(vm)
	if err != nil {
		// TTS no-arg constructor may fail on emulator -- non-fatal.
		fmt.Fprintf(os.Stderr, "  TTS (non-fatal): %v\n", err)
		return nil
	}
	defer tts.Close()

	return nil
}

func testTelecomWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := telecom.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new telecom manager: %w", err)
	}
	defer deleteGlobalRef(vm, mgr.Obj)

	return nil
}

func testUsageWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := usage.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new usage manager: %w", err)
	}
	defer deleteGlobalRef(vm, mgr.Obj)

	// IsAppInactive is an exported read-only query.
	inactive, err := mgr.IsAppInactive("com.android.shell")
	if err != nil {
		// May require PACKAGE_USAGE_STATS -- non-fatal.
		fmt.Fprintf(os.Stderr, "  isAppInactive (non-fatal): %v\n", err)
		return nil
	}
	fmt.Fprintf(os.Stderr, "  isAppInactive(com.android.shell) = %v\n", inactive)
	return nil
}

func testUsbWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := usb.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new usb manager: %w", err)
	}
	defer mgr.Close()

	return nil
}

func testWifiP2pWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := p2p.NewManager(ctx)
	if err != nil {
		// Wi-Fi P2P may not be available on emulator -- non-fatal.
		fmt.Fprintf(os.Stderr, "  wifi_p2p manager (non-fatal): %v\n", err)
		return nil
	}
	defer mgr.Close()

	return nil
}

func testWifiRttWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := rtt.NewManager(ctx)
	if err != nil {
		// Wi-Fi RTT may not be available on emulator -- non-fatal.
		fmt.Fprintf(os.Stderr, "  wifi_rtt manager (non-fatal): %v\n", err)
		return nil
	}
	defer deleteGlobalRef(vm, mgr.Obj)

	return nil
}

func testAccountsWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	// AccountManager is obtained via static factory AccountManager.get(ctx).
	// The wrapper's getManagerRaw is unexported, so we call the static method
	// via JNI and populate the exported Manager struct fields.
	var mgr accounts.Manager
	mgr.VM = vm

	err = vm.Do(func(env *jni.Env) error {
		if err := accounts.Init(env); err != nil {
			return err
		}
		cls, err := env.FindClass("android/accounts/AccountManager")
		if err != nil {
			return fmt.Errorf("find AccountManager: %w", err)
		}
		getMid, err := env.GetStaticMethodID(cls, "get",
			"(Landroid/content/Context;)Landroid/accounts/AccountManager;")
		if err != nil {
			return fmt.Errorf("get AccountManager.get: %w", err)
		}
		obj, err := env.CallStaticObjectMethod(cls, getMid,
			jni.ObjectValue(ctx.Obj))
		if err != nil {
			return fmt.Errorf("call AccountManager.get: %w", err)
		}
		if obj == nil || obj.Ref() == 0 {
			return fmt.Errorf("AccountManager.get returned null")
		}
		mgr.Obj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		return err
	}
	defer func() {
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(mgr.Obj)
			return nil
		})
	}()

	return nil
}

func testNfcAdapterWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	// NfcAdapter is obtained via NfcAdapter.getDefaultAdapter(ctx).
	// No exported factory in the wrapper, so use JNI directly.
	var adapter nfc.Adapter
	adapter.VM = vm

	err = vm.Do(func(env *jni.Env) error {
		if err := nfc.Init(env); err != nil {
			return err
		}
		cls, err := env.FindClass("android/nfc/NfcAdapter")
		if err != nil {
			return fmt.Errorf("find NfcAdapter: %w", err)
		}
		getMid, err := env.GetStaticMethodID(cls, "getDefaultAdapter",
			"(Landroid/content/Context;)Landroid/nfc/NfcAdapter;")
		if err != nil {
			return fmt.Errorf("get getDefaultAdapter: %w", err)
		}
		obj, err := env.CallStaticObjectMethod(cls, getMid,
			jni.ObjectValue(ctx.Obj))
		if err != nil {
			return fmt.Errorf("call getDefaultAdapter: %w", err)
		}
		if obj == nil || obj.Ref() == 0 {
			// NFC not available on this device/emulator.
			return nil
		}
		adapter.Obj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		// NFC may not be available -- non-fatal.
		fmt.Fprintf(os.Stderr, "  NFC adapter (non-fatal): %v\n", err)
		return nil
	}
	if adapter.Obj == nil {
		fmt.Fprintf(os.Stderr, "  NFC adapter not available on this device\n")
		return nil
	}
	defer adapter.Close()

	return nil
}

func testPreferencesWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	// SharedPreferences is obtained via Context.getSharedPreferences(name, mode).
	// No exported factory in the wrapper, so use JNI directly.
	var prefs preferences.Preferences
	prefs.VM = vm

	err = vm.Do(func(env *jni.Env) error {
		if err := preferences.Init(env); err != nil {
			return err
		}
		ctxCls, err := env.FindClass("android/content/Context")
		if err != nil {
			return fmt.Errorf("find Context: %w", err)
		}
		getPrefsMid, err := env.GetMethodID(ctxCls, "getSharedPreferences",
			"(Ljava/lang/String;I)Landroid/content/SharedPreferences;")
		if err != nil {
			return fmt.Errorf("get getSharedPreferences: %w", err)
		}
		name, err := env.NewStringUTF("e2e_test_prefs")
		if err != nil {
			return err
		}
		obj, err := env.CallObjectMethod(ctx.Obj, getPrefsMid,
			jni.ObjectValue(&name.Object), jni.IntValue(0))
		if err != nil {
			return fmt.Errorf("call getSharedPreferences: %w", err)
		}
		prefs.Obj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		return err
	}
	defer prefs.Close()

	// GetString with a non-existent key should return the default value.
	val, err := prefs.GetString("nonexistent_key", "default_val")
	if err != nil {
		return fmt.Errorf("getString: %w", err)
	}
	if val != "default_val" {
		return fmt.Errorf("getString = %q, want default_val", val)
	}

	// Contains should return false for a non-existent key.
	has, err := prefs.Contains("nonexistent_key")
	if err != nil {
		return fmt.Errorf("contains: %w", err)
	}
	if has {
		return fmt.Errorf("contains(nonexistent_key) = true")
	}

	return nil
}

func testResolverWrapper(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	// ContentResolver is obtained via Context.getContentResolver().
	var res resolver.Resolver
	res.VM = vm

	resolverObj := ctx.ContentResolver()
	if resolverObj == nil || resolverObj.Ref() == 0 {
		return fmt.Errorf("ContentResolver() returned nil")
	}

	err = vm.Do(func(env *jni.Env) error {
		if err := resolver.Init(env); err != nil {
			return err
		}
		res.Obj = env.NewGlobalRef(resolverObj)
		return nil
	})
	if err != nil {
		return err
	}
	defer deleteGlobalRef(vm, res.Obj)

	return nil
}

func testDocumentsInitWrapper(vm *jni.VM) error {
	// DocumentsContract has only static methods on an unexported type.
	// Verify Init resolves all classes and methods.
	return vm.Do(func(env *jni.Env) error {
		return documents.Init(env)
	})
}

func testMediastoreInitWrapper(vm *jni.VM) error {
	// MediaStore has only static methods on an unexported type.
	// Verify Init resolves all classes and methods.
	return vm.Do(func(env *jni.Env) error {
		return media.Init(env)
	})
}

func testPdfInitWrapper(vm *jni.VM) error {
	// PdfRenderer requires a ParcelFileDescriptor to construct.
	// Verify Init resolves all classes and methods.
	return vm.Do(func(env *jni.Env) error {
		return pdf.Init(env)
	})
}

func testToastInitWrapper(vm *jni.VM) error {
	// Toast has an unexported type. Verify Init resolves the class and method.
	// Then create a Toast via the static factory makeText and verify it works.
	return vm.Do(func(env *jni.Env) error {
		if err := toast.Init(env); err != nil {
			return err
		}
		// Additionally, call Toast.makeText as a smoke test.
		cls, err := env.FindClass("android/widget/Toast")
		if err != nil {
			return fmt.Errorf("find Toast: %w", err)
		}
		makeTextMid, err := env.GetStaticMethodID(cls, "makeText",
			"(Landroid/content/Context;Ljava/lang/CharSequence;I)Landroid/widget/Toast;")
		if err != nil {
			return fmt.Errorf("get makeText: %w", err)
		}
		_ = makeTextMid
		// We don't actually call makeText because we'd need a valid
		// UI-thread Context. Init verification is sufficient.
		return nil
	})
}

