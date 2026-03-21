//go:build android

package main

import (
	"fmt"
	"os"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/accounts"
	"github.com/AndroidGoLab/jni/app/download"
	"github.com/AndroidGoLab/jni/app/job"
	"github.com/AndroidGoLab/jni/app/notification"
	"github.com/AndroidGoLab/jni/app/usage"
	"github.com/AndroidGoLab/jni/companion"
	"github.com/AndroidGoLab/jni/content/preferences"
	"github.com/AndroidGoLab/jni/content/resolver"
	"github.com/AndroidGoLab/jni/graphics/pdf"
	"github.com/AndroidGoLab/jni/hardware/biometric"
	"github.com/AndroidGoLab/jni/hardware/camera"
	"github.com/AndroidGoLab/jni/hardware/ir"
	"github.com/AndroidGoLab/jni/hardware/lights"
	"github.com/AndroidGoLab/jni/hardware/usb"
	"github.com/AndroidGoLab/jni/media/audiomanager"
	"github.com/AndroidGoLab/jni/media/player"
	"github.com/AndroidGoLab/jni/media/projection"
	"github.com/AndroidGoLab/jni/media/recorder"
	"github.com/AndroidGoLab/jni/media/ringtone"
	"github.com/AndroidGoLab/jni/media/session"
	"github.com/AndroidGoLab/jni/net"
	"github.com/AndroidGoLab/jni/net/nsd"
	"github.com/AndroidGoLab/jni/net/wifi/p2p"
	"github.com/AndroidGoLab/jni/net/wifi/rtt"
	"github.com/AndroidGoLab/jni/nfc"
	"github.com/AndroidGoLab/jni/print"
	"github.com/AndroidGoLab/jni/provider/documents"
	"github.com/AndroidGoLab/jni/provider/media"
	"github.com/AndroidGoLab/jni/provider/settings"
	"github.com/AndroidGoLab/jni/se/omapi"
	"github.com/AndroidGoLab/jni/speech"
	"github.com/AndroidGoLab/jni/telecom"
	"github.com/AndroidGoLab/jni/view/inputmethod"
	"github.com/AndroidGoLab/jni/widget/toast"
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

	mgr, err := audiomanager.NewAudioManager(ctx)
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

	mgr, err := companion.NewDeviceManager(ctx)
	if err != nil {
		return fmt.Errorf("new companion manager: %w", err)
	}
	defer mgr.Close()

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

	mgr, err := inputmethod.NewInputMethodManager(ctx)
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

	mgr, err := ir.NewConsumerIrManager(ctx)
	if err != nil {
		// IR service may not exist on emulator -- non-fatal.
		fmt.Fprintf(os.Stderr, "  IR manager (non-fatal): %v\n", err)
		return nil
	}
	defer mgr.Close()

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

	// LightsManager has no generated constructor; obtain via GetSystemService.
	var mgr lights.Manager
	mgr.VM = vm
	err = vm.Do(func(env *jni.Env) error {
		if err := lights.Init(env); err != nil {
			return err
		}
		svc, err := ctx.GetSystemService("lights")
		if err != nil {
			return err
		}
		if svc == nil || svc.Ref() == 0 {
			return fmt.Errorf("lights service not available")
		}
		// GetSystemService already returns a GlobalRef, so use it directly
		// instead of wrapping again (which would leak the original).
		mgr.Obj = svc
		return nil
	})
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

	mgr, err := net.NewConnectivityManager(ctx)
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

	connected, _ := svc.IsConnected()
	fmt.Fprintf(os.Stderr, "  omapi isConnected = %v\n", connected)
	return nil
}

func testPlayerWrapper(vm *jni.VM) error {
	// MediaPlayer has no generated constructor; create via JNI new MediaPlayer().
	var p player.MediaPlayer
	p.VM = vm
	err := vm.Do(func(env *jni.Env) error {
		if err := player.Init(env); err != nil {
			return err
		}
		cls, err := env.FindClass("android/media/MediaPlayer")
		if err != nil {
			return fmt.Errorf("find MediaPlayer: %w", err)
		}
		initMid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("get MediaPlayer.<init>: %w", err)
		}
		obj, err := env.NewObject(cls, initMid)
		if err != nil {
			return fmt.Errorf("new MediaPlayer: %w", err)
		}
		p.Obj = env.NewGlobalRef(obj)
		return nil
	})
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

	mgr, err := projection.NewMediaProjectionManager(ctx)
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

	mgr, err := session.NewMediaSessionManager(ctx)
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

	mgr, err := usage.NewStatsManager(ctx)
	if err != nil {
		return fmt.Errorf("new usage manager: %w", err)
	}
	defer mgr.Close()

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

	mgr, err := p2p.NewWifiP2pManager(ctx)
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

	mgr, err := rtt.NewWifiRttManager(ctx)
	if err != nil {
		// Wi-Fi RTT may not be available on emulator -- non-fatal.
		fmt.Fprintf(os.Stderr, "  wifi_rtt manager (non-fatal): %v\n", err)
		return nil
	}
	defer mgr.Close()

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
	var mgr accounts.AccountManager
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
	defer deleteGlobalRef(vm, adapter.Obj)

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
	var prefs preferences.SharedPreferences
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
	defer deleteGlobalRef(vm, prefs.Obj)

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
	var res resolver.ContentResolver
	res.VM = vm

	resolverObj, err := ctx.GetContentResolver()
	if err != nil {
		return fmt.Errorf("ContentResolver: %w", err)
	}
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

func testDocumentsWrapper(vm *jni.VM) error {
	// DocumentsContract has static URI-builder methods.
	// Build a document URI and verify it is not nil.
	var dc documents.Contract
	dc.VM = vm

	uri, err := dc.BuildDocumentUri("com.example.provider", "doc:1234")
	if err != nil {
		return fmt.Errorf("buildDocumentUri: %w", err)
	}
	if uri == nil || uri.Ref() == 0 {
		return fmt.Errorf("buildDocumentUri returned nil")
	}
	defer deleteGlobalRef(vm, uri)

	rootUri, err := dc.BuildRootUri("com.example.provider", "root:0")
	if err != nil {
		return fmt.Errorf("buildRootUri: %w", err)
	}
	if rootUri == nil || rootUri.Ref() == 0 {
		return fmt.Errorf("buildRootUri returned nil")
	}
	defer deleteGlobalRef(vm, rootUri)

	return nil
}

func testMediastoreWrapper(vm *jni.VM) error {
	// MediaStore has static methods. Call GetMediaScannerUri and
	// GetPickImagesMaxLimit to verify they work.
	var ms media.MediaStore
	ms.VM = vm

	uri, err := ms.GetMediaScannerUri()
	if err != nil {
		return fmt.Errorf("getMediaScannerUri: %w", err)
	}
	if uri == nil || uri.Ref() == 0 {
		return fmt.Errorf("getMediaScannerUri returned nil")
	}
	defer deleteGlobalRef(vm, uri)

	limit, err := ms.GetPickImagesMaxLimit()
	if err != nil {
		// May not be available on older API levels -- non-fatal.
		fmt.Fprintf(os.Stderr, "  getPickImagesMaxLimit (non-fatal): %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "  pickImagesMaxLimit = %d\n", limit)
	}

	return nil
}

func testPdfWrapper(vm *jni.VM) error {
	// Create a small Bitmap via Bitmap.createBitmap(width, height, config),
	// then verify its dimensions with GetWidth/GetHeight.
	config, err := pdf.ARGB8888(vm)
	if err != nil {
		return fmt.Errorf("ARGB8888: %w", err)
	}
	defer deleteGlobalRef(vm, config)

	var bmp pdf.Bitmap
	bmp.VM = vm

	bmpObj, err := bmp.CreateBitmap3_10(16, 32, config)
	if err != nil {
		return fmt.Errorf("createBitmap: %w", err)
	}
	if bmpObj == nil || bmpObj.Ref() == 0 {
		return fmt.Errorf("createBitmap returned nil")
	}

	bmp.Obj = bmpObj
	defer deleteGlobalRef(vm, bmp.Obj)

	w, err := bmp.GetWidth()
	if err != nil {
		return fmt.Errorf("getWidth: %w", err)
	}
	if w != 16 {
		return fmt.Errorf("getWidth = %d, want 16", w)
	}

	h, err := bmp.GetHeight()
	if err != nil {
		return fmt.Errorf("getHeight: %w", err)
	}
	if h != 32 {
		return fmt.Errorf("getHeight = %d, want 32", h)
	}

	return nil
}

func testRingtoneWrapper(vm *jni.VM) error {
	// Call the static RingtoneManager.getDefaultUri(TYPE_RINGTONE)
	// to get the default ringtone URI and verify it is not nil.
	var mgr ringtone.Manager
	mgr.VM = vm

	uri, err := mgr.GetDefaultUri(int32(ringtone.TypeRingtone))
	if err != nil {
		return fmt.Errorf("getDefaultUri: %w", err)
	}
	if uri == nil || uri.Ref() == 0 {
		return fmt.Errorf("getDefaultUri returned nil")
	}
	defer deleteGlobalRef(vm, uri)

	// Verify IsDefault returns true for the default URI.
	isDef, err := mgr.IsDefault(uri)
	if err != nil {
		return fmt.Errorf("isDefault: %w", err)
	}
	if !isDef {
		return fmt.Errorf("isDefault(defaultUri) = false, want true")
	}

	return nil
}

func testSettingsWrapper(vm *jni.VM) error {
	// Read a system setting via Settings.System.getString using raw JNI.
	// The settings package only resolves class references; reading values
	// requires calling the static getString/getInt methods directly.
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	resolverObj, err := ctx.GetContentResolver()
	if err != nil {
		return fmt.Errorf("getContentResolver: %w", err)
	}
	if resolverObj == nil || resolverObj.Ref() == 0 {
		return fmt.Errorf("getContentResolver returned nil")
	}
	defer deleteGlobalRef(vm, resolverObj)

	return vm.Do(func(env *jni.Env) error {
		if err := settings.Init(env); err != nil {
			return fmt.Errorf("settings.Init: %w", err)
		}

		// Read Settings.Global.getString(resolver, "device_name").
		cls, err := env.FindClass("android/provider/Settings$Global")
		if err != nil {
			return fmt.Errorf("find Settings$Global: %w", err)
		}
		mid, err := env.GetStaticMethodID(cls, "getString",
			"(Landroid/content/ContentResolver;Ljava/lang/String;)Ljava/lang/String;")
		if err != nil {
			return fmt.Errorf("get getString: %w", err)
		}
		jName, err := env.NewStringUTF("device_name")
		if err != nil {
			return fmt.Errorf("NewStringUTF: %w", err)
		}
		_, err = env.CallStaticObjectMethod(cls, mid,
			jni.ObjectValue(resolverObj),
			jni.ObjectValue(&jName.Object))
		if err != nil {
			return fmt.Errorf("getString(device_name): %w", err)
		}
		// device_name may be null on some devices, so we don't check the value.
		// The fact that the call succeeded without error is sufficient.
		return nil
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

