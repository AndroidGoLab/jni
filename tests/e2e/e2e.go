//go:build android

// E2E tests for the JNI Go bindings, compiled as a c-shared library
// and run on a real Android device/emulator via app_process.
package main

/*
#include <jni.h>
*/
import "C"
import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/accounts"
	"github.com/xaionaro-go/jni/app"
	"github.com/xaionaro-go/jni/app/alarm"
	"github.com/xaionaro-go/jni/app/download"
	"github.com/xaionaro-go/jni/app/job"
	"github.com/xaionaro-go/jni/app/notification"
	"github.com/xaionaro-go/jni/app/usage"
	"github.com/xaionaro-go/jni/bluetooth"
	"github.com/xaionaro-go/jni/capi"
	"github.com/xaionaro-go/jni/companion"
	"github.com/xaionaro-go/jni/content/clipboard"
	"github.com/xaionaro-go/jni/content/permission"
	"github.com/xaionaro-go/jni/content/pm"
	"github.com/xaionaro-go/jni/content/preferences"
	"github.com/xaionaro-go/jni/content/resolver"
	"github.com/xaionaro-go/jni/credentials"
	"github.com/xaionaro-go/jni/graphics/pdf"
	"github.com/xaionaro-go/jni/hardware/biometric"
	"github.com/xaionaro-go/jni/hardware/camera"
	"github.com/xaionaro-go/jni/hardware/ir"
	"github.com/xaionaro-go/jni/hardware/lights"
	"github.com/xaionaro-go/jni/hardware/usb"
	"github.com/xaionaro-go/jni/health/connect"
	"github.com/xaionaro-go/jni/location"
	"github.com/xaionaro-go/jni/media/audiomanager"
	"github.com/xaionaro-go/jni/media/player"
	"github.com/xaionaro-go/jni/media/projection"
	"github.com/xaionaro-go/jni/media/recorder"
	"github.com/xaionaro-go/jni/media/session"
	"github.com/xaionaro-go/jni/net"
	"github.com/xaionaro-go/jni/net/nsd"
	"github.com/xaionaro-go/jni/net/wifi"
	"github.com/xaionaro-go/jni/net/wifi/p2p"
	"github.com/xaionaro-go/jni/net/wifi/rtt"
	"github.com/xaionaro-go/jni/nfc"
	"github.com/xaionaro-go/jni/os/battery"
	"github.com/xaionaro-go/jni/os/build"
	"github.com/xaionaro-go/jni/os/environment"
	"github.com/xaionaro-go/jni/os/keyguard"
	"github.com/xaionaro-go/jni/os/power"
	"github.com/xaionaro-go/jni/os/storage"
	"github.com/xaionaro-go/jni/os/vibrator"
	"github.com/xaionaro-go/jni/print"
	"github.com/xaionaro-go/jni/provider/media"
	"github.com/xaionaro-go/jni/provider/documents"
	"github.com/xaionaro-go/jni/provider/settings"
	"github.com/xaionaro-go/jni/se/omapi"
	"github.com/xaionaro-go/jni/security/keystore"
	"github.com/xaionaro-go/jni/speech"
	"github.com/xaionaro-go/jni/telecom"
	"github.com/xaionaro-go/jni/telephony"
	"github.com/xaionaro-go/jni/view/display"
	"github.com/xaionaro-go/jni/view/inputmethod"
	"github.com/xaionaro-go/jni/widget/toast"
)

//export runE2ETests
func runE2ETests(cvm *C.JavaVM) {
	vm := jni.VMFromPtr(unsafe.Pointer(cvm))
	passed := 0
	failed := 0

	run := func(name string, f func(vm *jni.VM) error) {
		err := f(vm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: %s: %v\n", name, err)
			failed++
		} else {
			fmt.Fprintf(os.Stderr, "PASS: %s\n", name)
			passed++
		}
	}

	// xfail runs a test that is expected to fail (e.g. AndroidX classes
	// not available in the framework classpath). Failure counts as PASS.
	xfail := func(name string, f func(vm *jni.VM) error) {
		err := f(vm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "XFAIL: %s: %v\n", name, err)
			passed++
		} else {
			fmt.Fprintf(os.Stderr, "XPASS: %s (unexpected pass)\n", name)
			passed++
		}
	}

	fmt.Fprintln(os.Stderr, "=== Running E2E tests ===")

	// --- Core JNI surface ---
	fmt.Fprintln(os.Stderr, "--- Core JNI ---")
	run("FindClass", testFindClass)
	run("GetSuperclass", testGetSuperclass)
	run("IsAssignableFrom", testIsAssignableFrom)
	run("GetObjectClass", testGetObjectClass)
	run("NewStringUTF", testNewStringUTF)
	run("GoString", testGoString)
	run("GetStringLength", testGetStringLength)
	run("GetStringUTFLength", testGetStringUTFLength)
	run("NewObject", testNewObject)
	run("CallInstanceMethod", testCallInstanceMethod)
	run("CallStaticLongMethod", testCallStaticLongMethod)
	run("CallStaticIntMethod", testCallStaticIntMethod)
	run("CallStaticBooleanMethod", testCallStaticBooleanMethod)
	run("CallStaticObjectMethod", testCallStaticObjectMethod)
	run("CallStaticVoidMethod", testCallStaticVoidMethod)
	run("StaticFieldAccess", testStaticFieldAccess)
	run("InstanceFieldAccess", testInstanceFieldAccess)
	run("IntArray", testIntArray)
	run("ByteArray", testByteArray)
	run("ObjectArray", testObjectArray)
	run("ArrayLength", testArrayLength)
	run("GlobalRef", testGlobalRef)
	run("LocalRef", testLocalRef)
	run("WeakGlobalRef", testWeakGlobalRef)
	run("LocalFrame", testLocalFrame)
	run("ExceptionHandling", testExceptionHandling)
	run("ThrowNew", testThrowNew)
	run("IsInstanceOf", testIsInstanceOf)
	run("IsSameObject", testIsSameObject)
	run("GetObjectRefType", testGetObjectRefType)
	run("MonitorEnterExit", testMonitorEnterExit)
	run("DirectByteBuffer", testDirectByteBuffer)
	run("GetVersion", testGetVersion)
	run("ConcurrentDo", testConcurrentDo)
	run("NewProxy", testNewProxy)

	// --- Generated wrapper package Init (validates all FindClass + GetMethodID) ---
	fmt.Fprintln(os.Stderr, "--- Package Init (all 53) ---")
	run("Init/accounts", initTest(accounts.Init))
	run("Init/alarm", initTest(alarm.Init))
	run("Init/app", initTest(app.Init))
	run("Init/audiomanager", initTest(audiomanager.Init))
	run("Init/battery", initTest(battery.Init))
	run("Init/biometric", initTest(biometric.Init))
	run("Init/bluetooth", initTest(bluetooth.Init))
	run("Init/build", initTest(build.Init))
	run("Init/camera", initTest(camera.Init))
	run("Init/clipboard", initTest(clipboard.Init))
	run("Init/companion", initTest(companion.Init))
	xfail("Init/credentials", initTest(credentials.Init)) // AndroidX: not in framework classpath
	run("Init/display", initTest(display.Init))
	run("Init/documents", initTest(documents.Init))
	run("Init/download", initTest(download.Init))
	run("Init/environment", initTest(environment.Init))
	xfail("Init/health", initTest(connect.Init)) // AndroidX: not in framework classpath
	run("Init/inputmethod", initTest(inputmethod.Init))
	run("Init/ir", initTest(ir.Init))
	run("Init/job", initTest(job.Init))
	run("Init/keyguard", initTest(keyguard.Init))
	run("Init/keystore", initTest(keystore.Init))
	run("Init/lights", initTest(lights.Init))
	run("Init/location", initTest(location.Init))
	run("Init/mediastore", initTest(media.Init))
	run("Init/net", initTest(net.Init))
	run("Init/nfc", initTest(nfc.Init))
	run("Init/notification", initTest(notification.Init))
	run("Init/nsd", initTest(nsd.Init))
	run("Init/omapi", initTest(omapi.Init))
	run("Init/pdf", initTest(pdf.Init))
	xfail("Init/permission", initTest(permission.Init)) // AndroidX: not in framework classpath
	run("Init/player", initTest(player.Init))
	run("Init/pm", initTest(pm.Init))
	run("Init/power", initTest(power.Init))
	run("Init/preferences", initTest(preferences.Init))
	run("Init/print", initTest(print.Init))
	run("Init/projection", initTest(projection.Init))
	run("Init/recorder", initTest(recorder.Init))
	run("Init/resolver", initTest(resolver.Init))
	run("Init/session", initTest(session.Init))
	run("Init/settings", initTest(settings.Init))
	run("Init/speech", initTest(speech.Init))
	run("Init/storage", initTest(storage.Init))
	run("Init/telecom", initTest(telecom.Init))
	run("Init/telephony", initTest(telephony.Init))
	run("Init/toast", initTest(toast.Init))
	run("Init/usage", initTest(usage.Init))
	run("Init/usb", initTest(usb.Init))
	run("Init/vibrator", initTest(vibrator.Init))
	run("Init/wifi", initTest(wifi.Init))
	run("Init/wifi_p2p", initTest(p2p.Init))
	run("Init/wifi_rtt", initTest(rtt.Init))

	// --- Tier 1: packages testable without Context ---
	fmt.Fprintln(os.Stderr, "--- Tier 1: no-context functional tests ---")
	run("app/Intent", testAppIntent)
	run("app/Bundle", testAppBundle)
	run("alarm/AlarmClockInfo", testAlarmClockInfo)
	run("build/BuildInfo", testBuildWrapper)
	run("environment/Paths", testEnvironmentWrapper)
	xfail("keystore/KeyStore", testKeystoreKeyStore) // AndroidKeyStore not registered in app_process
	run("notification/Channel", testNotificationChannel)
	run("location/NilContext", testLocationNilContext)
	run("location/GetLocation", testLocationGetLocation)

	// --- Tier 2: system service tests (first batch) ---
	fmt.Fprintln(os.Stderr, "--- Tier 2a: system service tests ---")
	run("bluetooth/Adapter", testBluetoothWrapper)
	run("wifi/Manager", testWifiWrapper)
	run("telephony/Manager", testTelephonyWrapper)
	run("pm/Manager", testPmWrapper)
	run("power/Manager", testPowerWrapper)
	run("vibrator/Vibrator", testVibratorWrapper)
	run("clipboard/Manager", testClipboardWrapper)
	run("keyguard/Manager", testKeyguardWrapper)
	run("display/WindowManager", testDisplayWrapper)
	run("storage/Manager", testStorageWrapper)

	// --- Tier 2: system service tests (second batch) ---
	fmt.Fprintln(os.Stderr, "--- Tier 2b: system service tests ---")
	run("audiomanager/Wrapper", testAudioManagerWrapper)
	run("biometric/Wrapper", testBiometricWrapper)
	run("camera/Wrapper", testCameraWrapper)
	xfail("companion/Wrapper", testCompanionWrapper) // companion_device service not available on emulator
	run("download/Wrapper", testDownloadWrapper)
	run("inputmethod/Wrapper", testInputMethodWrapper)
	run("ir/Wrapper", testIrWrapper)
	run("job/Wrapper", testJobSchedulerWrapper)
	run("lights/Wrapper", testLightsWrapper)
	run("net/Wrapper", testNetWrapper)
	run("notification/Wrapper", testNotificationWrapper)
	run("nsd/Wrapper", testNsdWrapper)
	run("omapi/Wrapper", testOmapiWrapper)
	run("player/Wrapper", testPlayerWrapper)
	run("print/Wrapper", testPrintWrapper)
	run("projection/Wrapper", testProjectionWrapper)
	run("recorder/Wrapper", testRecorderWrapper)
	xfail("session/Wrapper", testSessionWrapper) // MediaServiceManager null in app_process
	run("speech/TTS", testSpeechTTSWrapper)
	run("telecom/Wrapper", testTelecomWrapper)
	run("usage/Wrapper", testUsageWrapper)
	run("usb/Wrapper", testUsbWrapper)
	run("wifi_p2p/Wrapper", testWifiP2pWrapper)
	run("wifi_rtt/Wrapper", testWifiRttWrapper)
	run("accounts/Wrapper", testAccountsWrapper)
	run("nfc/Adapter", testNfcAdapterWrapper)
	run("preferences/Wrapper", testPreferencesWrapper)
	run("resolver/Wrapper", testResolverWrapper)
	run("documents/Init", testDocumentsInitWrapper)
	run("mediastore/Init", testMediastoreInitWrapper)
	run("pdf/Init", testPdfInitWrapper)
	run("toast/Init", testToastInitWrapper)

	fmt.Fprintf(os.Stderr, "\n=== Results: %d passed, %d failed ===\n", passed, failed)
	if failed > 0 {
		fmt.Fprintln(os.Stderr, "E2E_FAIL")
	} else {
		fmt.Fprintln(os.Stderr, "E2E_PASS")
	}

	_ = capi.JNI_OK
}

// initTest wraps a package Init function as a test.
func initTest(initFn func(env *jni.Env) error) func(vm *jni.VM) error {
	return func(vm *jni.VM) error {
		return vm.Do(func(env *jni.Env) error {
			return initFn(env)
		})
	}
}

// =====================================================================
// Core JNI tests
// =====================================================================

func testFindClass(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/String")
		if err != nil {
			return fmt.Errorf("FindClass java/lang/String: %w", err)
		}
		if cls == nil {
			return fmt.Errorf("FindClass returned nil")
		}
		return nil
	})
}

func testGetSuperclass(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/Integer")
		if err != nil {
			return fmt.Errorf("FindClass Integer: %w", err)
		}
		super := env.GetSuperclass(cls)
		if super == nil {
			return fmt.Errorf("GetSuperclass returned nil")
		}
		numberCls, err := env.FindClass("java/lang/Number")
		if err != nil {
			return fmt.Errorf("FindClass Number: %w", err)
		}
		if !env.IsSameObject((*jni.Object)(unsafe.Pointer(super)), (*jni.Object)(unsafe.Pointer(numberCls))) {
			return fmt.Errorf("Integer superclass is not Number")
		}
		return nil
	})
}

func testIsAssignableFrom(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		intCls, err := env.FindClass("java/lang/Integer")
		if err != nil {
			return fmt.Errorf("FindClass Integer: %w", err)
		}
		numCls, err := env.FindClass("java/lang/Number")
		if err != nil {
			return fmt.Errorf("FindClass Number: %w", err)
		}
		if !env.IsAssignableFrom(intCls, numCls) {
			return fmt.Errorf("Integer should be assignable to Number")
		}
		if env.IsAssignableFrom(numCls, intCls) {
			return fmt.Errorf("Number should NOT be assignable to Integer")
		}
		return nil
	})
}

func testGetObjectClass(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		str, err := env.NewStringUTF("hello")
		if err != nil {
			return fmt.Errorf("NewStringUTF: %w", err)
		}
		cls := env.GetObjectClass(&str.Object)
		if cls == nil {
			return fmt.Errorf("GetObjectClass returned nil")
		}
		strCls, err := env.FindClass("java/lang/String")
		if err != nil {
			return fmt.Errorf("FindClass String: %w", err)
		}
		if !env.IsSameObject((*jni.Object)(unsafe.Pointer(cls)), (*jni.Object)(unsafe.Pointer(strCls))) {
			return fmt.Errorf("object class is not String")
		}
		return nil
	})
}

func testNewStringUTF(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		str, err := env.NewStringUTF("hello from Go on Android")
		if err != nil {
			return fmt.Errorf("NewStringUTF: %w", err)
		}
		if str == nil {
			return fmt.Errorf("NewStringUTF returned nil")
		}
		return nil
	})
}

func testGoString(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		want := "round-trip test 日本語"
		str, err := env.NewStringUTF(want)
		if err != nil {
			return fmt.Errorf("NewStringUTF: %w", err)
		}
		got := env.GoString(str)
		if got != want {
			return fmt.Errorf("GoString = %q, want %q", got, want)
		}
		return nil
	})
}

func testGetStringLength(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		str, err := env.NewStringUTF("hello")
		if err != nil {
			return err
		}
		if length := env.GetStringLength(str); length != 5 {
			return fmt.Errorf("GetStringLength = %d, want 5", length)
		}
		return nil
	})
}

func testGetStringUTFLength(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		str, err := env.NewStringUTF("日")
		if err != nil {
			return err
		}
		if n := env.GetStringUTFLength(str); n != 3 {
			return fmt.Errorf("GetStringUTFLength('日') = %d, want 3", n)
		}
		if n := env.GetStringLength(str); n != 1 {
			return fmt.Errorf("GetStringLength('日') = %d, want 1", n)
		}
		return nil
	})
}

func testNewObject(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/StringBuilder")
		if err != nil {
			return err
		}
		mid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return err
		}
		obj, err := env.NewObject(cls, mid)
		if err != nil {
			return fmt.Errorf("NewObject: %w", err)
		}
		if obj == nil {
			return fmt.Errorf("NewObject returned nil")
		}
		return nil
	})
}

func testCallInstanceMethod(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/StringBuilder")
		if err != nil {
			return err
		}
		initMid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return err
		}
		sb, err := env.NewObject(cls, initMid)
		if err != nil {
			return err
		}
		appendMid, err := env.GetMethodID(cls, "append", "(Ljava/lang/String;)Ljava/lang/StringBuilder;")
		if err != nil {
			return err
		}
		for _, s := range []string{"Hello", " World"} {
			jstr, err := env.NewStringUTF(s)
			if err != nil {
				return err
			}
			_, err = env.CallObjectMethod(sb, appendMid, jni.ObjectValue(&jstr.Object))
			if err != nil {
				return err
			}
		}
		toStrMid, err := env.GetMethodID(cls, "toString", "()Ljava/lang/String;")
		if err != nil {
			return err
		}
		resultObj, err := env.CallObjectMethod(sb, toStrMid)
		if err != nil {
			return err
		}
		got := env.GoString((*jni.String)(unsafe.Pointer(resultObj)))
		if got != "Hello World" {
			return fmt.Errorf("got %q, want %q", got, "Hello World")
		}
		lenMid, err := env.GetMethodID(cls, "length", "()I")
		if err != nil {
			return err
		}
		length, err := env.CallIntMethod(sb, lenMid)
		if err != nil {
			return err
		}
		if length != 11 {
			return fmt.Errorf("length = %d, want 11", length)
		}
		return nil
	})
}

func testCallStaticLongMethod(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/System")
		if err != nil {
			return err
		}
		mid, err := env.GetStaticMethodID(cls, "currentTimeMillis", "()J")
		if err != nil {
			return err
		}
		result, err := env.CallStaticLongMethod(cls, mid)
		if err != nil {
			return err
		}
		if result <= 0 {
			return fmt.Errorf("currentTimeMillis = %d", result)
		}
		return nil
	})
}

func testCallStaticIntMethod(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/Integer")
		if err != nil {
			return err
		}
		mid, err := env.GetStaticMethodID(cls, "parseInt", "(Ljava/lang/String;)I")
		if err != nil {
			return err
		}
		jstr, err := env.NewStringUTF("42")
		if err != nil {
			return err
		}
		result, err := env.CallStaticIntMethod(cls, mid, jni.ObjectValue(&jstr.Object))
		if err != nil {
			return err
		}
		if result != 42 {
			return fmt.Errorf("parseInt('42') = %d", result)
		}
		return nil
	})
}

func testCallStaticBooleanMethod(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/Boolean")
		if err != nil {
			return err
		}
		mid, err := env.GetStaticMethodID(cls, "parseBoolean", "(Ljava/lang/String;)Z")
		if err != nil {
			return err
		}
		jstr, err := env.NewStringUTF("true")
		if err != nil {
			return err
		}
		result, err := env.CallStaticBooleanMethod(cls, mid, jni.ObjectValue(&jstr.Object))
		if err != nil {
			return err
		}
		if result == 0 {
			return fmt.Errorf("parseBoolean('true') = false")
		}
		return nil
	})
}

func testCallStaticObjectMethod(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/Integer")
		if err != nil {
			return err
		}
		mid, err := env.GetStaticMethodID(cls, "valueOf", "(I)Ljava/lang/Integer;")
		if err != nil {
			return err
		}
		obj, err := env.CallStaticObjectMethod(cls, mid, jni.IntValue(99))
		if err != nil {
			return err
		}
		if obj == nil {
			return fmt.Errorf("valueOf(99) returned nil")
		}
		intMid, err := env.GetMethodID(cls, "intValue", "()I")
		if err != nil {
			return err
		}
		val, err := env.CallIntMethod(obj, intMid)
		if err != nil {
			return err
		}
		if val != 99 {
			return fmt.Errorf("intValue = %d", val)
		}
		return nil
	})
}

func testCallStaticVoidMethod(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/System")
		if err != nil {
			return err
		}
		mid, err := env.GetStaticMethodID(cls, "gc", "()V")
		if err != nil {
			return err
		}
		return env.CallStaticVoidMethod(cls, mid)
	})
}

func testStaticFieldAccess(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/Integer")
		if err != nil {
			return err
		}
		fid, err := env.GetStaticFieldID(cls, "MAX_VALUE", "I")
		if err != nil {
			return err
		}
		if val := env.GetStaticIntField(cls, fid); val != 2147483647 {
			return fmt.Errorf("Integer.MAX_VALUE = %d", val)
		}
		return nil
	})
}

func testInstanceFieldAccess(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("android/graphics/Point")
		if err != nil {
			return err
		}
		initMid, err := env.GetMethodID(cls, "<init>", "(II)V")
		if err != nil {
			return err
		}
		pt, err := env.NewObject(cls, initMid, jni.IntValue(10), jni.IntValue(20))
		if err != nil {
			return err
		}
		xFid, err := env.GetFieldID(cls, "x", "I")
		if err != nil {
			return err
		}
		if x := env.GetIntField(pt, xFid); x != 10 {
			return fmt.Errorf("Point.x = %d, want 10", x)
		}
		env.SetIntField(pt, xFid, 42)
		if x := env.GetIntField(pt, xFid); x != 42 {
			return fmt.Errorf("Point.x after set = %d, want 42", x)
		}
		return nil
	})
}

func testIntArray(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		arr := env.NewIntArray(5)
		data := [5]int32{10, 20, 30, 40, 50}
		env.SetIntArrayRegion(arr, 0, 5, unsafe.Pointer(&data[0]))
		var buf [5]int32
		env.GetIntArrayRegion(arr, 0, 5, unsafe.Pointer(&buf[0]))
		for i := range 5 {
			if buf[i] != data[i] {
				return fmt.Errorf("[%d] = %d, want %d", i, buf[i], data[i])
			}
		}
		return nil
	})
}

func testByteArray(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		arr := env.NewByteArray(4)
		data := [4]int8{1, 2, 3, 4}
		env.SetByteArrayRegion(arr, 0, 4, unsafe.Pointer(&data[0]))
		var buf [4]int8
		env.GetByteArrayRegion(arr, 0, 4, unsafe.Pointer(&buf[0]))
		for i := range 4 {
			if buf[i] != data[i] {
				return fmt.Errorf("[%d] = %d, want %d", i, buf[i], data[i])
			}
		}
		return nil
	})
}

func testObjectArray(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		strCls, err := env.FindClass("java/lang/String")
		if err != nil {
			return err
		}
		arr, err := env.NewObjectArray(3, strCls, nil)
		if err != nil {
			return err
		}
		strs := []string{"alpha", "beta", "gamma"}
		for i, s := range strs {
			jstr, err := env.NewStringUTF(s)
			if err != nil {
				return err
			}
			if err := env.SetObjectArrayElement(arr, int32(i), &jstr.Object); err != nil {
				return err
			}
		}
		for i, want := range strs {
			elem, err := env.GetObjectArrayElement(arr, int32(i))
			if err != nil {
				return err
			}
			if got := env.GoString((*jni.String)(unsafe.Pointer(elem))); got != want {
				return fmt.Errorf("[%d] = %q, want %q", i, got, want)
			}
		}
		return nil
	})
}

func testArrayLength(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		arr := env.NewIntArray(7)
		if n := env.GetArrayLength(&arr.Array); n != 7 {
			return fmt.Errorf("GetArrayLength = %d, want 7", n)
		}
		return nil
	})
}

func testGlobalRef(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		str, err := env.NewStringUTF("global ref test")
		if err != nil {
			return err
		}
		global := env.NewGlobalRef(&str.Object)
		if global == nil {
			return fmt.Errorf("NewGlobalRef returned nil")
		}
		got := env.GoString((*jni.String)(unsafe.Pointer(global)))
		env.DeleteGlobalRef(global)
		if got != "global ref test" {
			return fmt.Errorf("got %q", got)
		}
		return nil
	})
}

func testLocalRef(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		str, err := env.NewStringUTF("local ref test")
		if err != nil {
			return err
		}
		local := env.NewLocalRef(&str.Object)
		if local == nil {
			return fmt.Errorf("NewLocalRef returned nil")
		}
		got := env.GoString((*jni.String)(unsafe.Pointer(local)))
		env.DeleteLocalRef(local)
		if got != "local ref test" {
			return fmt.Errorf("got %q", got)
		}
		return nil
	})
}

func testWeakGlobalRef(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		str, err := env.NewStringUTF("weak ref test")
		if err != nil {
			return err
		}
		weak := env.NewWeakGlobalRef(&str.Object)
		if weak == nil {
			return fmt.Errorf("NewWeakGlobalRef returned nil")
		}
		if env.IsSameObject((*jni.Object)(unsafe.Pointer(weak)), nil) {
			return fmt.Errorf("weak ref was collected")
		}
		env.DeleteWeakGlobalRef(weak)
		return nil
	})
}

func testLocalFrame(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		if err := env.PushLocalFrame(16); err != nil {
			return err
		}
		str, err := env.NewStringUTF("in local frame")
		if err != nil {
			env.PopLocalFrame(nil)
			return err
		}
		result := env.PopLocalFrame(&str.Object)
		if result == nil {
			return fmt.Errorf("PopLocalFrame returned nil")
		}
		if got := env.GoString((*jni.String)(unsafe.Pointer(result))); got != "in local frame" {
			return fmt.Errorf("got %q", got)
		}
		return nil
	})
}

func testExceptionHandling(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		if env.ExceptionCheck() {
			return fmt.Errorf("unexpected pending exception")
		}
		_, err := env.FindClass("com/nonexistent/Class")
		if err == nil {
			return fmt.Errorf("FindClass should have failed")
		}
		if env.ExceptionCheck() {
			env.ExceptionClear()
			return fmt.Errorf("exception not cleared by Go wrapper")
		}
		return nil
	})
}

func testThrowNew(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("java/lang/RuntimeException")
		if err != nil {
			return err
		}
		if err := env.ThrowNew(cls, "test exception from Go"); err != nil {
			return err
		}
		if !env.ExceptionCheck() {
			return fmt.Errorf("no exception after ThrowNew")
		}
		throwable := env.ExceptionOccurred()
		env.ExceptionClear()
		if throwable == nil {
			return fmt.Errorf("ExceptionOccurred nil")
		}
		thCls := env.GetObjectClass((*jni.Object)(unsafe.Pointer(throwable)))
		getMsgMid, err := env.GetMethodID(thCls, "getMessage", "()Ljava/lang/String;")
		if err != nil {
			return err
		}
		msgObj, err := env.CallObjectMethod((*jni.Object)(unsafe.Pointer(throwable)), getMsgMid)
		if err != nil {
			return err
		}
		if msg := env.GoString((*jni.String)(unsafe.Pointer(msgObj))); msg != "test exception from Go" {
			return fmt.Errorf("message = %q", msg)
		}
		return nil
	})
}

func testIsInstanceOf(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		str, _ := env.NewStringUTF("test")
		strCls, _ := env.FindClass("java/lang/String")
		objCls, _ := env.FindClass("java/lang/Object")
		numCls, _ := env.FindClass("java/lang/Number")
		if !env.IsInstanceOf(&str.Object, strCls) {
			return fmt.Errorf("not instance of String")
		}
		if !env.IsInstanceOf(&str.Object, objCls) {
			return fmt.Errorf("not instance of Object")
		}
		if env.IsInstanceOf(&str.Object, numCls) {
			return fmt.Errorf("should not be Number")
		}
		return nil
	})
}

func testIsSameObject(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		s1, _ := env.NewStringUTF("a")
		s2, _ := env.NewStringUTF("b")
		if !env.IsSameObject(&s1.Object, &s1.Object) {
			return fmt.Errorf("same ref not same")
		}
		if env.IsSameObject(&s1.Object, &s2.Object) {
			return fmt.Errorf("different refs same")
		}
		if !env.IsSameObject(nil, nil) {
			return fmt.Errorf("nil != nil")
		}
		return nil
	})
}

func testGetObjectRefType(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		str, _ := env.NewStringUTF("ref type test")
		if t := env.GetObjectRefType(&str.Object); t != capi.JNILocalRefType {
			return fmt.Errorf("local ref type = %d", t)
		}
		global := env.NewGlobalRef(&str.Object)
		t := env.GetObjectRefType(global)
		env.DeleteGlobalRef(global)
		if t != capi.JNIGlobalRefType {
			return fmt.Errorf("global ref type = %d", t)
		}
		return nil
	})
}

func testMonitorEnterExit(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		str, _ := env.NewStringUTF("monitor")
		if err := env.MonitorEnter(&str.Object); err != nil {
			return err
		}
		return env.MonitorExit(&str.Object)
	})
}

func testDirectByteBuffer(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		data := make([]byte, 64)
		for i := range data {
			data[i] = byte(i)
		}
		buf := env.NewDirectByteBuffer(unsafe.Pointer(&data[0]), 64)
		if buf == nil {
			return fmt.Errorf("nil")
		}
		if c := env.GetDirectBufferCapacity(buf); c != 64 {
			return fmt.Errorf("capacity = %d", c)
		}
		addr := env.GetDirectBufferAddress(buf)
		slice := unsafe.Slice((*byte)(addr), 64)
		for i := range 64 {
			if slice[i] != byte(i) {
				return fmt.Errorf("[%d] = %d", i, slice[i])
			}
		}
		return nil
	})
}

func testGetVersion(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		if v := env.GetVersion(); v < int32(capi.JNI_VERSION_1_6) {
			return fmt.Errorf("version = 0x%x", v)
		}
		return nil
	})
}

func testConcurrentDo(vm *jni.VM) error {
	const n = 10
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := vm.Do(func(env *jni.Env) error {
				s := fmt.Sprintf("goroutine-%d", idx)
				str, err := env.NewStringUTF(s)
				if err != nil {
					return err
				}
				if got := env.GoString(str); got != s {
					return fmt.Errorf("got %q, want %q", got, s)
				}
				return nil
			}); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		return err
	}
	return nil
}

func testNewProxy(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		runnableCls, err := env.FindClass("java/lang/Runnable")
		if err != nil {
			return err
		}
		proxy, cleanup, err := env.NewProxy(
			[]*jni.Class{runnableCls},
			func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
				return nil, nil
			},
		)
		if err != nil {
			return fmt.Errorf("NewProxy: %w", err)
		}
		defer cleanup()
		if proxy == nil {
			return fmt.Errorf("proxy is nil")
		}
		if !env.IsInstanceOf(proxy, (*jni.Class)(unsafe.Pointer(runnableCls))) {
			return fmt.Errorf("proxy not Runnable")
		}
		runMid, err := env.GetMethodID(runnableCls, "run", "()V")
		if err != nil {
			return err
		}
		return env.CallVoidMethod(proxy, runMid)
	})
}

// =====================================================================
// Generated wrapper functional tests
// =====================================================================

func testAppIntent(vm *jni.VM) error {
	intent, err := app.NewIntent(vm)
	if err != nil {
		return fmt.Errorf("NewIntent: %w", err)
	}

	// SetAction + GetAction round-trip
	intent.SetAction("android.intent.action.VIEW")
	if action := intent.GetAction(); action != "android.intent.action.VIEW" {
		return fmt.Errorf("GetAction = %q, want android.intent.action.VIEW", action)
	}

	// PutStringExtra + GetStringExtra round-trip
	intent.PutStringExtra("key1", "value1")
	if val := intent.GetStringExtra("key1"); val != "value1" {
		return fmt.Errorf("GetStringExtra = %q, want value1", val)
	}

	// PutIntExtra + GetIntExtra round-trip
	intent.PutIntExtra("num", 42)
	if n := intent.GetIntExtra("num", 0); n != 42 {
		return fmt.Errorf("GetIntExtra = %d, want 42", n)
	}

	// PutBoolExtra + GetBoolExtra round-trip
	intent.PutBoolExtra("flag", true)
	if b := intent.GetBoolExtra("flag", false); !b {
		return fmt.Errorf("GetBoolExtra = false, want true")
	}

	return nil
}

func testBuildWrapper(vm *jni.VM) error {
	// GetBuildInfo triggers ensureInit and resolves android.os.Build classes
	_, err := build.GetBuildInfo(vm)
	if err != nil {
		return fmt.Errorf("GetBuildInfo: %w", err)
	}
	_, err = build.GetVersionInfo(vm)
	if err != nil {
		return fmt.Errorf("GetVersionInfo: %w", err)
	}
	return nil
}

func testEnvironmentWrapper(vm *jni.VM) error {
	return vm.Do(func(env *jni.Env) error {
		if err := environment.Init(env); err != nil {
			return fmt.Errorf("Init: %w", err)
		}
		// environment package has static methods that don't need a Context
		// Init already validates all FindClass + GetMethodID succeed
		return nil
	})
}

func testLocationNilContext(vm *jni.VM) error {
	_, err := location.NewManager(nil)
	if err == nil {
		return fmt.Errorf("NewManager(nil) should have returned error")
	}
	return nil
}

// getSystemContext obtains an Android Context via ActivityThread in app_process.
func getSystemContext(vm *jni.VM) (*app.Context, error) {
	var ctx app.Context
	ctx.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}

		atClass, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread: %w", err)
		}

		// Try currentActivityThread() first (returns existing thread or null).
		currentMid, err := env.GetStaticMethodID(atClass, "currentActivityThread", "()Landroid/app/ActivityThread;")
		if err != nil {
			return fmt.Errorf("get currentActivityThread: %w", err)
		}
		atObj, _ := env.CallStaticObjectMethod(atClass, currentMid)

		if atObj == nil || atObj.Ref() == 0 {
			// Prepare Looper before creating ActivityThread.
			looperClass, err := env.FindClass("android/os/Looper")
			if err != nil {
				return fmt.Errorf("find Looper: %w", err)
			}
			prepMid, err := env.GetStaticMethodID(looperClass, "prepareMainLooper", "()V")
			if err != nil {
				return fmt.Errorf("get prepareMainLooper: %w", err)
			}
			// Ignore error — may already be prepared.
			_ = env.CallStaticVoidMethod(looperClass, prepMid)

			// Create ActivityThread via systemMain().
			sysMid, err := env.GetStaticMethodID(atClass, "systemMain", "()Landroid/app/ActivityThread;")
			if err != nil {
				return fmt.Errorf("get systemMain: %w", err)
			}
			atObj, err = env.CallStaticObjectMethod(atClass, sysMid)
			if err != nil {
				return fmt.Errorf("call systemMain: %w", err)
			}
		}

		getCtxMid, err := env.GetMethodID(atClass, "getSystemContext", "()Landroid/app/ContextImpl;")
		if err != nil {
			return fmt.Errorf("get getSystemContext: %w", err)
		}
		sysCtxObj, err := env.CallObjectMethod(atObj, getCtxMid)
		if err != nil {
			return fmt.Errorf("call getSystemContext: %w", err)
		}

		// Create a package context for com.android.shell (matches shell uid 2000
		// and has location/network permissions).
		ctxClass, err := env.FindClass("android/content/Context")
		if err != nil {
			return fmt.Errorf("find Context class: %w", err)
		}
		createPkgCtxMid, err := env.GetMethodID(ctxClass, "createPackageContext", "(Ljava/lang/String;I)Landroid/content/Context;")
		if err != nil {
			return fmt.Errorf("get createPackageContext: %w", err)
		}
		pkgName, err := env.NewStringUTF("com.android.shell")
		if err != nil {
			return fmt.Errorf("new string: %w", err)
		}
		shellCtxObj, err := env.CallObjectMethod(sysCtxObj, createPkgCtxMid,
			jni.ObjectValue(&pkgName.Object), jni.IntValue(0))
		if err != nil {
			// Fall back to system context if createPackageContext fails.
			ctx.Obj = env.NewGlobalRef(sysCtxObj)
			return nil
		}

		ctx.Obj = env.NewGlobalRef(shellCtxObj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}

func testLocationGetLocation(vm *jni.VM) error {
	ctx, err := getSystemContext(vm)
	if err != nil {
		return fmt.Errorf("get system context: %w", err)
	}
	defer ctx.Close()

	mgr, err := location.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("new location manager: %w", err)
	}
	defer mgr.Close()

	gpsEnabled, err := mgr.IsProviderEnabled("gps")
	if err != nil {
		return fmt.Errorf("isProviderEnabled: %w", err)
	}
	if !gpsEnabled {
		return fmt.Errorf("GPS provider not enabled on emulator")
	}

	// Try all providers: gps first, then fused, then network.
	var locObj *jni.Object
	for _, provider := range []string{"gps", "fused", "network"} {
		enabled, _ := mgr.IsProviderEnabled(provider)
		if !enabled {
			continue
		}
		locObj, err = mgr.GetLastKnownLocation(provider)
		if err != nil {
			continue
		}
		if locObj != nil && locObj.Ref() != 0 {
			break
		}
	}

	if locObj == nil || locObj.Ref() == 0 {
		// No cached location from any provider. Create a Location object
		// manually via JNI to validate the extraction pipeline.
		err = vm.Do(func(env *jni.Env) error {
			locClass, err := env.FindClass("android/location/Location")
			if err != nil {
				return fmt.Errorf("find Location class: %w", err)
			}
			initMid, err := env.GetMethodID(locClass, "<init>", "(Ljava/lang/String;)V")
			if err != nil {
				return fmt.Errorf("get Location.<init>: %w", err)
			}
			provStr, err := env.NewStringUTF("test")
			if err != nil {
				return err
			}
			obj, err := env.NewObject(locClass, initMid, jni.ObjectValue(&provStr.Object))
			if err != nil {
				return fmt.Errorf("new Location: %w", err)
			}

			// Set latitude/longitude via setLatitude/setLongitude.
			setLatMid, err := env.GetMethodID(locClass, "setLatitude", "(D)V")
			if err != nil {
				return err
			}
			setLonMid, err := env.GetMethodID(locClass, "setLongitude", "(D)V")
			if err != nil {
				return err
			}
			setTimeMid, err := env.GetMethodID(locClass, "setTime", "(J)V")
			if err != nil {
				return err
			}
			if err := env.CallVoidMethod(obj, setLatMid, jni.DoubleValue(37.4220)); err != nil {
				return err
			}
			if err := env.CallVoidMethod(obj, setLonMid, jni.DoubleValue(-122.0840)); err != nil {
				return err
			}
			if err := env.CallVoidMethod(obj, setTimeMid, jni.LongValue(1709000000000)); err != nil {
				return err
			}
			locObj = obj
			return nil
		})
		if err != nil {
			return fmt.Errorf("create test location: %w", err)
		}
		fmt.Fprintf(os.Stderr, "  (no cached location; using synthetic test location)\n")
	}

	// Extract the location struct with lat/lng.
	var loc *location.Location
	err = vm.Do(func(env *jni.Env) error {
		var extractErr error
		loc, extractErr = location.ExtractLocation(env, locObj)
		return extractErr
	})
	if err != nil {
		return fmt.Errorf("extract location: %w", err)
	}

	if loc.Latitude == 0 && loc.Longitude == 0 {
		return fmt.Errorf("location is (0, 0)")
	}

	fmt.Fprintf(os.Stderr, "  location: lat=%.4f lon=%.4f provider=%s\n",
		loc.Latitude, loc.Longitude, loc.Provider)
	return nil
}

func main() {}
