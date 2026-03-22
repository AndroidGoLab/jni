//go:build android

// Command camera_record demonstrates recording a short video from the
// camera using the Camera2 API + MediaRecorder — all via JNI from Go,
// with no NDK camera code. Uses GoAbstractDispatch proxy adapters for
// Camera2's abstract class callbacks.
package main

/*
#include <android/native_activity.h>

extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static void _setCallbacks(ANativeActivity* a) {
    a->callbacks->onResume = _onResume;
    a->callbacks->onNativeWindowCreated = _onWindowCreated;
}
*/
import "C"
import (
	"bytes"
	"fmt"
	"os"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/media/recorder"
)

const templateRecord = 3 // CameraDevice.TEMPLATE_RECORD

// Cached class references and main-thread Handler for Camera2 callbacks.
var (
	clsDeviceCallback  *jni.GlobalRef
	clsSessionCallback *jni.GlobalRef
	mainHandler        *jni.GlobalRef
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	vm := jni.VMFromPtr(unsafe.Pointer(activity.vm))
	actObj := jni.ObjectFromPtr(unsafe.Pointer(activity.clazz))
	ui.OnCreate(vm, actObj)

	// Set ClassLoader and cache adapter classes on the main thread
	// (FindClass from goroutines can't see APK classes).
	vm.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(actObj)
		mid, err := env.GetMethodID(cls, "getClassLoader", "()Ljava/lang/ClassLoader;")
		if err != nil {
			return err
		}
		cl, err := env.CallObjectMethod(actObj, mid)
		if err != nil {
			return err
		}
		globalCL := env.NewGlobalRef(cl)
		jni.SetProxyClassLoader(globalCL)

		// Load adapter classes via ClassLoader.loadClass() since
		// FindClass uses the boot ClassLoader in NativeActivity context.
		clsCL := env.GetObjectClass(cl)
		loadClassMid, err := env.GetMethodID(clsCL, "loadClass", "(Ljava/lang/String;)Ljava/lang/Class;")
		if err != nil {
			println("get loadClass:", err.Error())
			return nil
		}

		loadClass := func(name string) *jni.GlobalRef {
			jName, err := env.NewStringUTF(name)
			if err != nil {
				println("newStringUTF:", err.Error())
				return nil
			}
			defer env.DeleteLocalRef(&jName.Object)
			cls, err := env.CallObjectMethod(cl, loadClassMid, jni.ObjectValue(&jName.Object))
			if err != nil {
				println("loadClass "+name+":", err.Error())
				return nil
			}
			return env.NewGlobalRef(cls)
		}

		clsDeviceCallback = loadClass("center.dx.jni.generated.CameraDeviceCallback")
		clsSessionCallback = loadClass("center.dx.jni.generated.CaptureSessionCallback")

		// Create a Handler on the main Looper for Camera2 callbacks.
		looperCls, _ := env.FindClass("android/os/Looper")
		getMainMid, _ := env.GetStaticMethodID(looperCls, "getMainLooper", "()Landroid/os/Looper;")
		mainLooper, _ := env.CallStaticObjectMethod(looperCls, getMainMid)
		handlerCls, _ := env.FindClass("android/os/Handler")
		handlerInit, _ := env.GetMethodID(handlerCls, "<init>", "(Landroid/os/Looper;)V")
		handler, _ := env.NewObject(handlerCls, handlerInit, jni.ObjectValue(mainLooper))
		mainHandler = env.NewGlobalRef(handler)

		jni.EnsureProxyInit(env)
		return nil
	})
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)))
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
}

// cameraState holds JNI global references for camera lifecycle.
type cameraState struct {
	vm           *jni.VM
	device       *jni.GlobalRef
	session      *jni.GlobalRef
	devHandlerID int64
	sesHandlerID int64
}

func (cs *cameraState) close() {
	// Stop repeating and close session.
	if cs.session != nil {
		cs.vm.Do(func(env *jni.Env) error {
			cls := env.GetObjectClass(cs.session)
			if mid, err := env.GetMethodID(cls, "stopRepeating", "()V"); err == nil {
				env.CallVoidMethod(cs.session, mid)
			}
			if mid, err := env.GetMethodID(cls, "close", "()V"); err == nil {
				env.CallVoidMethod(cs.session, mid)
			}
			return nil
		})
		// Small delay to let in-flight frames drain.
		time.Sleep(100 * time.Millisecond)
		cs.vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(cs.session)
			return nil
		})
	}

	// Close camera device.
	if cs.device != nil {
		cs.vm.Do(func(env *jni.Env) error {
			cls := env.GetObjectClass(cs.device)
			if mid, err := env.GetMethodID(cls, "close", "()V"); err == nil {
				env.CallVoidMethod(cs.device, mid)
			}
			env.DeleteGlobalRef(cs.device)
			return nil
		})
	}

	jni.UnregisterProxyHandler(cs.devHandlerID)
	jni.UnregisterProxyHandler(cs.sesHandlerID)
}

// startCamera opens the first back camera via the Java Camera2 API and
// starts a repeating RECORD request targeting recSurface.
func startCamera(
	vm *jni.VM,
	activity *jni.Object,
	recSurface *jni.Object,
) (*cameraState, error) {
	cs := &cameraState{vm: vm}

	// Get camera ID.
	var cameraID string
	err := vm.Do(func(env *jni.Env) error {
		mgrCls, err := env.FindClass("android/hardware/camera2/CameraManager")
		if err != nil {
			return fmt.Errorf("find CameraManager: %w", err)
		}

		// Get CameraManager from context.
		ctxCls := env.GetObjectClass(activity)
		getSysMid, err := env.GetMethodID(ctxCls, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;")
		if err != nil {
			return fmt.Errorf("get getSystemService: %w", err)
		}
		svcName, err := env.NewStringUTF("camera")
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&svcName.Object)
		mgrObj, err := env.CallObjectMethod(activity, getSysMid, jni.ObjectValue(&svcName.Object))
		if err != nil {
			return fmt.Errorf("getSystemService(camera): %w", err)
		}

		// getCameraIdList() -> String[]
		getIdsMid, err := env.GetMethodID(mgrCls, "getCameraIdList", "()[Ljava/lang/String;")
		if err != nil {
			return fmt.Errorf("get getCameraIdList: %w", err)
		}
		idsObj, err := env.CallObjectMethod(mgrObj, getIdsMid)
		if err != nil {
			return fmt.Errorf("getCameraIdList: %w", err)
		}
		if idsObj == nil {
			return fmt.Errorf("getCameraIdList returned null")
		}
		length := env.GetArrayLength((*jni.Array)(unsafe.Pointer(idsObj)))
		if length == 0 {
			return fmt.Errorf("no cameras available")
		}
		firstID, err := env.GetObjectArrayElement((*jni.ObjectArray)(unsafe.Pointer(idsObj)), 0)
		if err != nil {
			return fmt.Errorf("get camera ID: %w", err)
		}
		cameraID = env.GoString((*jni.String)(unsafe.Pointer(firstID)))
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Register device callback and open camera.
	type deviceResult struct {
		device *jni.Object
		err    error
	}
	deviceCh := make(chan deviceResult, 1)
	cs.devHandlerID = jni.RegisterProxyHandler(func(env *jni.Env, method string, args []*jni.Object) (*jni.Object, error) {
		switch method {
		case "onOpened":
			deviceCh <- deviceResult{device: env.NewGlobalRef(args[0])}
		case "onError":
			deviceCh <- deviceResult{err: fmt.Errorf("camera error")}
		case "onDisconnected":
			deviceCh <- deviceResult{err: fmt.Errorf("camera disconnected")}
		}
		return nil, nil
	})

	err = vm.Do(func(env *jni.Env) error {
		// Create CameraDeviceCallback adapter using cached class ref.
		cbCls := (*jni.Class)(unsafe.Pointer(clsDeviceCallback))
		cbInit, err := env.GetMethodID(cbCls, "<init>", "(J)V")
		if err != nil {
			return fmt.Errorf("get CameraDeviceCallback.<init>: %w", err)
		}
		cbObj, err := env.NewObject(cbCls, cbInit, jni.LongValue(cs.devHandlerID))
		if err != nil {
			return fmt.Errorf("new CameraDeviceCallback: %w", err)
		}

		// Call CameraManager.openCamera(String, StateCallback, Handler).
		mgrCls, _ := env.FindClass("android/hardware/camera2/CameraManager")
		ctxCls := env.GetObjectClass(activity)
		getSysMid, _ := env.GetMethodID(ctxCls, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;")
		svcName, _ := env.NewStringUTF("camera")
		defer env.DeleteLocalRef(&svcName.Object)
		mgrObj, _ := env.CallObjectMethod(activity, getSysMid, jni.ObjectValue(&svcName.Object))

		openMid, err := env.GetMethodID(mgrCls, "openCamera",
			"(Ljava/lang/String;Landroid/hardware/camera2/CameraDevice$StateCallback;Landroid/os/Handler;)V")
		if err != nil {
			return fmt.Errorf("get openCamera: %w", err)
		}
		jCamID, _ := env.NewStringUTF(cameraID)
		defer env.DeleteLocalRef(&jCamID.Object)
		return env.CallVoidMethod(mgrObj, openMid,
			jni.ObjectValue(&jCamID.Object),
			jni.ObjectValue(cbObj),
			jni.ObjectValue((*jni.Object)(unsafe.Pointer(mainHandler))))
	})
	if err != nil {
		jni.UnregisterProxyHandler(cs.devHandlerID)
		return nil, fmt.Errorf("openCamera: %w", err)
	}

	// Wait for camera to open.
	dr := <-deviceCh
	if dr.err != nil {
		jni.UnregisterProxyHandler(cs.devHandlerID)
		return nil, fmt.Errorf("camera open: %w", dr.err)
	}
	cs.device = (*jni.GlobalRef)(unsafe.Pointer(dr.device))

	// Register session callback and create capture session.
	type sessionResult struct {
		session *jni.Object
		err     error
	}
	sessionCh := make(chan sessionResult, 1)
	cs.sesHandlerID = jni.RegisterProxyHandler(func(env *jni.Env, method string, args []*jni.Object) (*jni.Object, error) {
		switch method {
		case "onConfigured":
			sessionCh <- sessionResult{session: env.NewGlobalRef(args[0])}
		case "onConfigureFailed":
			sessionCh <- sessionResult{err: fmt.Errorf("session configure failed")}
		}
		return nil, nil
	})

	err = vm.Do(func(env *jni.Env) error {
		// Create CaptureSessionCallback adapter using cached class ref.
		scCls := (*jni.Class)(unsafe.Pointer(clsSessionCallback))
		scInit, err := env.GetMethodID(scCls, "<init>", "(J)V")
		if err != nil {
			return fmt.Errorf("get CaptureSessionCallback.<init>: %w", err)
		}
		scObj, err := env.NewObject(scCls, scInit, jni.LongValue(cs.sesHandlerID))
		if err != nil {
			return fmt.Errorf("new CaptureSessionCallback: %w", err)
		}

		// Build List<Surface> for createCaptureSession.
		listCls, err := env.FindClass("java/util/ArrayList")
		if err != nil {
			return fmt.Errorf("find ArrayList: %w", err)
		}
		listInit, err := env.GetMethodID(listCls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("get ArrayList.<init>: %w", err)
		}
		listObj, err := env.NewObject(listCls, listInit)
		if err != nil {
			return fmt.Errorf("new ArrayList: %w", err)
		}
		addMid, err := env.GetMethodID(listCls, "add", "(Ljava/lang/Object;)Z")
		if err != nil {
			return fmt.Errorf("get ArrayList.add: %w", err)
		}
		env.CallBooleanMethod(listObj, addMid, jni.ObjectValue(recSurface))

		// CameraDevice.createCaptureSession(List, StateCallback, Handler)
		devCls := env.GetObjectClass(cs.device)
		createSessMid, err := env.GetMethodID(devCls, "createCaptureSession",
			"(Ljava/util/List;Landroid/hardware/camera2/CameraCaptureSession$StateCallback;Landroid/os/Handler;)V")
		if err != nil {
			return fmt.Errorf("get createCaptureSession: %w", err)
		}
		return env.CallVoidMethod(cs.device, createSessMid,
			jni.ObjectValue(listObj),
			jni.ObjectValue(scObj),
			jni.ObjectValue((*jni.Object)(unsafe.Pointer(mainHandler))))
	})
	if err != nil {
		cs.close()
		return nil, fmt.Errorf("createCaptureSession: %w", err)
	}

	// Wait for session.
	sr := <-sessionCh
	if sr.err != nil {
		cs.close()
		return nil, fmt.Errorf("capture session: %w", sr.err)
	}
	cs.session = (*jni.GlobalRef)(unsafe.Pointer(sr.session))

	// Create capture request and start repeating.
	err = vm.Do(func(env *jni.Env) error {
		devCls := env.GetObjectClass(cs.device)

		// createCaptureRequest(TEMPLATE_RECORD) -> CaptureRequest.Builder
		createReqMid, err := env.GetMethodID(devCls, "createCaptureRequest",
			"(I)Landroid/hardware/camera2/CaptureRequest$Builder;")
		if err != nil {
			return fmt.Errorf("get createCaptureRequest: %w", err)
		}
		builder, err := env.CallObjectMethod(cs.device, createReqMid, jni.IntValue(templateRecord))
		if err != nil {
			return fmt.Errorf("createCaptureRequest: %w", err)
		}

		// Builder.addTarget(surface)
		builderCls := env.GetObjectClass(builder)
		addTargetMid, err := env.GetMethodID(builderCls, "addTarget", "(Landroid/view/Surface;)V")
		if err != nil {
			return fmt.Errorf("get addTarget: %w", err)
		}
		if err := env.CallVoidMethod(builder, addTargetMid, jni.ObjectValue(recSurface)); err != nil {
			return fmt.Errorf("addTarget: %w", err)
		}

		// Builder.build() -> CaptureRequest
		buildMid, err := env.GetMethodID(builderCls, "build", "()Landroid/hardware/camera2/CaptureRequest;")
		if err != nil {
			return fmt.Errorf("get build: %w", err)
		}
		request, err := env.CallObjectMethod(builder, buildMid)
		if err != nil {
			return fmt.Errorf("build: %w", err)
		}

		// session.setRepeatingRequest(request, null, null)
		sessCls := env.GetObjectClass(cs.session)
		setRepMid, err := env.GetMethodID(sessCls, "setRepeatingRequest",
			"(Landroid/hardware/camera2/CaptureRequest;Landroid/hardware/camera2/CameraCaptureSession$CaptureCallback;Landroid/os/Handler;)I")
		if err != nil {
			return fmt.Errorf("get setRepeatingRequest: %w", err)
		}
		_, err = env.CallIntMethod(cs.session, setRepMid,
			jni.ObjectValue(request),
			jni.ObjectValue(nil),
			jni.ObjectValue((*jni.Object)(unsafe.Pointer(mainHandler))))
		if err != nil {
			return fmt.Errorf("setRepeatingRequest: %w", err)
		}
		return nil
	})
	if err != nil {
		cs.close()
		return nil, err
	}

	return cs, nil
}

func run(vm *jni.VM, output *bytes.Buffer) error {
	fmt.Fprintln(output, "=== Camera Record (JNI) ===")
	ui.RenderOutput()

	// 1. Get the app context and cache directory path.
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}

	cacheDirObj, err := ctx.GetCacheDir()
	if err != nil {
		return fmt.Errorf("getCacheDir: %w", err)
	}
	var cacheDir string
	err = vm.Do(func(env *jni.Env) error {
		fileCls, err := env.FindClass("java/io/File")
		if err != nil {
			return fmt.Errorf("find File: %w", err)
		}
		mid, err := env.GetMethodID(fileCls, "getAbsolutePath", "()Ljava/lang/String;")
		if err != nil {
			return fmt.Errorf("get getAbsolutePath: %w", err)
		}
		pathObj, err := env.CallObjectMethod(cacheDirObj, mid)
		if err != nil {
			return fmt.Errorf("getAbsolutePath: %w", err)
		}
		cacheDir = env.GoString((*jni.String)(unsafe.Pointer(pathObj)))
		return nil
	})
	if err != nil {
		return fmt.Errorf("cache dir: %w", err)
	}
	outputPath := cacheDir + "/test_recording.mp4"
	fmt.Fprintf(output, "Output: %s\n", outputPath)
	ui.RenderOutput()

	// 2. Create and configure MediaRecorder.
	var recObj *jni.GlobalRef
	err = vm.Do(func(env *jni.Env) error {
		cls, err := env.FindClass("android/media/MediaRecorder")
		if err != nil {
			return fmt.Errorf("find MediaRecorder: %w", err)
		}
		initMid, err := env.GetMethodID(cls, "<init>", "()V")
		if err != nil {
			return fmt.Errorf("get <init>: %w", err)
		}
		obj, err := env.NewObject(cls, initMid)
		if err != nil {
			return fmt.Errorf("new MediaRecorder: %w", err)
		}
		recObj = env.NewGlobalRef(obj)
		return nil
	})
	if err != nil {
		return fmt.Errorf("create recorder: %w", err)
	}
	rec := &recorder.MediaRecorder{VM: vm, Obj: recObj}

	if err := rec.SetAudioSource(recorder.AudioSourceMIC); err != nil {
		return fmt.Errorf("setAudioSource: %w", err)
	}
	if err := rec.SetVideoSource(recorder.VideoSourceSurface); err != nil {
		return fmt.Errorf("setVideoSource: %w", err)
	}
	if err := rec.SetOutputFormat(recorder.OutputFormatMPEG4); err != nil {
		return fmt.Errorf("setOutputFormat: %w", err)
	}
	if err := rec.SetAudioEncoder(recorder.AudioEncoderAAC); err != nil {
		return fmt.Errorf("setAudioEncoder: %w", err)
	}
	if err := rec.SetVideoEncoder(recorder.VideoEncoderH264); err != nil {
		return fmt.Errorf("setVideoEncoder: %w", err)
	}
	if err := rec.SetVideoSize(1920, 1080); err != nil {
		return fmt.Errorf("setVideoSize: %w", err)
	}
	if err := rec.SetVideoFrameRate(60); err != nil {
		return fmt.Errorf("setVideoFrameRate: %w", err)
	}
	if err := rec.SetVideoEncodingBitRate(10_000_000); err != nil {
		return fmt.Errorf("setVideoEncodingBitRate: %w", err)
	}
	if err := rec.SetOutputFile1_2(outputPath); err != nil {
		return fmt.Errorf("setOutputFile: %w", err)
	}
	fmt.Fprintln(output, "Recorder configured")
	ui.RenderOutput()

	// 3. Prepare the recorder.
	prepStart := time.Now()
	if err := rec.Prepare(); err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	prepDur := time.Since(prepStart)
	fmt.Fprintf(output, "prepare() OK (%v)\n", prepDur)
	ui.RenderOutput()

	// Start CPU profiling.
	profilePath := cacheDir + "/cpu.prof"
	profFile, err := os.Create(profilePath)
	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}
	pprof.StartCPUProfile(profFile)
	fmt.Fprintf(output, "CPU profiling started: %s\n", profilePath)
	ui.RenderOutput()

	// 4. Get the recorder's input surface (Java Surface object).
	surfObj, err := rec.GetSurface()
	if err != nil {
		return fmt.Errorf("getSurface: %w", err)
	}
	if surfObj == nil || surfObj.Ref() == 0 {
		return fmt.Errorf("getSurface returned null")
	}
	fmt.Fprintln(output, "Got recorder surface")
	ui.RenderOutput()

	// 5. Start the MediaRecorder before camera so frames have a consumer.
	startT := time.Now()
	if err := rec.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	startDur := time.Since(startT)
	fmt.Fprintf(output, "Recording started (%v)\n", startDur)
	ui.RenderOutput()

	// 6. Open camera and start feeding frames via JNI Camera2 API.
	activityObj := ui.ActivityRef()
	cam, err := startCamera(vm, activityObj, surfObj)
	if err != nil {
		_ = rec.Stop()
		_ = rec.Release()
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(recObj)
			return nil
		})
		return fmt.Errorf("camera: %w", err)
	}
	fmt.Fprintln(output, "Camera streaming (JNI)")
	ui.RenderOutput()

	// 7. Record for 5 seconds.
	time.Sleep(5 * time.Second)

	// 8. Stop the recorder first (while camera is still streaming).
	stopT := time.Now()
	if err := rec.Stop(); err != nil {
		fmt.Fprintf(output, "stop err: %v\n", err)
	} else {
		fmt.Fprintln(output, "Recording stopped")
	}
	stopDur := time.Since(stopT)
	fmt.Fprintf(output, "stop took %v\n", stopDur)
	ui.RenderOutput()

	// Stop CPU profiling.
	pprof.StopCPUProfile()
	profFile.Close()
	fmt.Fprintf(output, "CPU profile written: %s\n", profilePath)
	ui.RenderOutput()

	// 9. Close camera.
	cam.close()
	fmt.Fprintln(output, "Camera closed")
	ui.RenderOutput()

	// 10. Release recorder.
	_ = rec.Release()
	vm.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(recObj)
		return nil
	})
	fmt.Fprintln(output, "Released resources")
	ui.RenderOutput()

	// 11. Report file sizes.
	videoInfo, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("stat video: %w", err)
	}
	profInfo, err := os.Stat(profilePath)
	if err != nil {
		return fmt.Errorf("stat profile: %w", err)
	}

	fmt.Fprintf(output, "\nVideo: %s\n", outputPath)
	fmt.Fprintf(output, "Video size: %d bytes (%.1f MB)\n", videoInfo.Size(), float64(videoInfo.Size())/(1024*1024))
	fmt.Fprintf(output, "CPU profile: %s\n", profilePath)
	fmt.Fprintf(output, "Profile size: %d bytes\n", profInfo.Size())
	fmt.Fprintf(output, "\nTiming: prepare=%v start=%v stop=%v\n", prepDur, startDur, stopDur)
	fmt.Fprintln(output, "\nCamera record complete (zero NDK).")
	return nil
}
