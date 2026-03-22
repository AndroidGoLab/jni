//go:build android

// Command camera_record demonstrates recording a short video from the
// camera using the Camera2 API + MediaRecorder via JNI. It uses
// setVideoSource(SURFACE) to get a recording surface from MediaRecorder,
// then opens a Camera2 device via the NDK and feeds camera frames into
// that surface. Records for 3 seconds and reports the resulting file size.
package main

/*
#cgo LDFLAGS: -lcamera2ndk -lmediandk

#include <android/native_activity.h>
#include <android/native_window_jni.h>
#include <camera/NdkCameraDevice.h>
#include <camera/NdkCameraManager.h>
#include <camera/NdkCameraCaptureSession.h>
#include <unistd.h>

extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static void _setCallbacks(ANativeActivity* a) {
    a->callbacks->onResume = _onResume;
    a->callbacks->onNativeWindowCreated = _onWindowCreated;
}

// surfaceToWindow converts a JNI Surface jobject to an ANativeWindow*.
static ANativeWindow* surfaceToWindow(void* env, void* surface) {
    return ANativeWindow_fromSurface((JNIEnv*)env, (jobject)surface);
}

// NDK Camera2 callbacks (no-ops for this example).
static void onDisconnected(void* ctx, ACameraDevice* dev) {}
static void onError(void* ctx, ACameraDevice* dev, int err) {}
static void onSessionReady(void* ctx, ACameraCaptureSession* sess) {}
static void onSessionActive(void* ctx, ACameraCaptureSession* sess) {}
static void onSessionClosed(void* ctx, ACameraCaptureSession* sess) {}

// Camera state for step-by-step control from Go.
typedef struct {
    ACameraManager*              mgr;
    ACameraDevice*               device;
    ACameraCaptureSession*       session;
    ACaptureSessionOutput*       sessOut;
    ACaptureSessionOutputContainer* outCont;
    ACaptureRequest*             request;
    ACameraOutputTarget*         outTarget;
} CameraState;

// cameraOpen opens the first camera and starts a repeating RECORD
// request targeting recorderWindow. Returns 0 on success.
static int cameraOpen(CameraState* state, ANativeWindow* recorderWindow) {
    camera_status_t status;

    state->mgr = ACameraManager_create();
    if (!state->mgr) return -1;

    ACameraIdList* idList = NULL;
    status = ACameraManager_getCameraIdList(state->mgr, &idList);
    if (status != ACAMERA_OK || !idList || idList->numCameras == 0) {
        ACameraManager_delete(state->mgr);
        state->mgr = NULL;
        return -2;
    }
    const char* cameraId = idList->cameraIds[0];

    ACameraDevice_StateCallbacks devCb = {
        .context = NULL,
        .onDisconnected = onDisconnected,
        .onError = onError,
    };
    status = ACameraManager_openCamera(state->mgr, cameraId, &devCb, &state->device);
    ACameraManager_deleteCameraIdList(idList);
    if (status != ACAMERA_OK || !state->device) {
        ACameraManager_delete(state->mgr);
        state->mgr = NULL;
        return -3;
    }

    ACaptureSessionOutput_create(recorderWindow, &state->sessOut);
    ACaptureSessionOutputContainer_create(&state->outCont);
    ACaptureSessionOutputContainer_add(state->outCont, state->sessOut);

    ACameraCaptureSession_stateCallbacks sessCb = {
        .context = NULL,
        .onReady = onSessionReady,
        .onActive = onSessionActive,
        .onClosed = onSessionClosed,
    };
    status = ACameraDevice_createCaptureSession(
        state->device, state->outCont, &sessCb, &state->session);
    if (status != ACAMERA_OK || !state->session) {
        ACameraDevice_close(state->device);
        ACaptureSessionOutputContainer_free(state->outCont);
        ACaptureSessionOutput_free(state->sessOut);
        ACameraManager_delete(state->mgr);
        state->mgr = NULL;
        return -4;
    }

    status = ACameraDevice_createCaptureRequest(
        state->device, TEMPLATE_RECORD, &state->request);
    if (status != ACAMERA_OK || !state->request) {
        ACameraCaptureSession_close(state->session);
        ACameraDevice_close(state->device);
        ACaptureSessionOutputContainer_free(state->outCont);
        ACaptureSessionOutput_free(state->sessOut);
        ACameraManager_delete(state->mgr);
        state->mgr = NULL;
        return -5;
    }

    ACameraOutputTarget_create(recorderWindow, &state->outTarget);
    ACaptureRequest_addTarget(state->request, state->outTarget);

    int seqId = 0;
    status = ACameraCaptureSession_setRepeatingRequest(
        state->session, NULL, 1, &state->request, &seqId);
    if (status != ACAMERA_OK) {
        ACameraOutputTarget_free(state->outTarget);
        ACaptureRequest_free(state->request);
        ACameraCaptureSession_close(state->session);
        ACameraDevice_close(state->device);
        ACaptureSessionOutputContainer_free(state->outCont);
        ACaptureSessionOutput_free(state->sessOut);
        ACameraManager_delete(state->mgr);
        state->mgr = NULL;
        return -6;
    }

    return 0;
}

// cameraClose stops capture and releases all camera resources.
static void cameraClose(CameraState* state) {
    if (!state->mgr) return;
    ACameraCaptureSession_stopRepeating(state->session);
    // Small delay to let in-flight frames drain.
    usleep(100 * 1000);
    ACameraCaptureSession_close(state->session);
    ACameraOutputTarget_free(state->outTarget);
    ACaptureRequest_free(state->request);
    ACameraDevice_close(state->device);
    ACaptureSessionOutputContainer_free(state->outCont);
    ACaptureSessionOutput_free(state->sessOut);
    ACameraManager_delete(state->mgr);
    state->mgr = NULL;
}
*/
import "C"
import (
	"bytes"
	"fmt"
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	ui.OnCreate(
		jni.VMFromPtr(unsafe.Pointer(activity.vm)),
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(
		jni.ObjectFromPtr(unsafe.Pointer(activity.clazz)),
	)
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
}

// MediaRecorder constants from the Android SDK.
const (
	audioSourceMIC     = 1 // MediaRecorder.AudioSource.MIC
	videoSourceSurface = 2 // MediaRecorder.VideoSource.SURFACE
	outputFormatMPEG4  = 2 // MediaRecorder.OutputFormat.MPEG_4
	audioEncoderAAC    = 3 // MediaRecorder.AudioEncoder.AAC
	videoEncoderH264   = 2 // MediaRecorder.VideoEncoder.H264
)

func run(vm *jni.VM, output *bytes.Buffer) error {
	fmt.Fprintln(output, "=== Camera Record ===")
	ui.RenderOutput()

	// 1. Get the app context.
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	fmt.Fprintln(output, "Got app context")
	ui.RenderOutput()

	// 2. Get the cache directory path.
	var cacheDir string
	err = vm.Do(func(env *jni.Env) error {
		ctxCls := env.GetObjectClass(ctx.Obj)
		mid, err := env.GetMethodID(ctxCls, "getCacheDir", "()Ljava/io/File;")
		if err != nil {
			return fmt.Errorf("get getCacheDir: %w", err)
		}
		cacheDirObj, err := env.CallObjectMethod(ctx.Obj, mid)
		if err != nil {
			return fmt.Errorf("getCacheDir: %w", err)
		}
		fileCls, err := env.FindClass("java/io/File")
		if err != nil {
			return fmt.Errorf("find File: %w", err)
		}
		getPathMid, err := env.GetMethodID(fileCls, "getAbsolutePath", "()Ljava/lang/String;")
		if err != nil {
			return fmt.Errorf("get getAbsolutePath: %w", err)
		}
		pathObj, err := env.CallObjectMethod(cacheDirObj, getPathMid)
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

	// 3. Create and configure MediaRecorder with SURFACE video source.
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

	// Helper: call a void method on recObj.
	callRecVoid := func(name, sig string, args ...jni.Value) error {
		return vm.Do(func(env *jni.Env) error {
			cls := env.GetObjectClass(recObj)
			mid, err := env.GetMethodID(cls, name, sig)
			if err != nil {
				return fmt.Errorf("get %s: %w", name, err)
			}
			return env.CallVoidMethod(recObj, mid, args...)
		})
	}

	// Configure: sources -> format -> encoders -> params -> file.
	if err := callRecVoid("setAudioSource", "(I)V", jni.IntValue(audioSourceMIC)); err != nil {
		return fmt.Errorf("setAudioSource: %w", err)
	}
	if err := callRecVoid("setVideoSource", "(I)V", jni.IntValue(videoSourceSurface)); err != nil {
		return fmt.Errorf("setVideoSource: %w", err)
	}
	if err := callRecVoid("setOutputFormat", "(I)V", jni.IntValue(outputFormatMPEG4)); err != nil {
		return fmt.Errorf("setOutputFormat: %w", err)
	}
	if err := callRecVoid("setAudioEncoder", "(I)V", jni.IntValue(audioEncoderAAC)); err != nil {
		return fmt.Errorf("setAudioEncoder: %w", err)
	}
	if err := callRecVoid("setVideoEncoder", "(I)V", jni.IntValue(videoEncoderH264)); err != nil {
		return fmt.Errorf("setVideoEncoder: %w", err)
	}
	if err := callRecVoid("setVideoSize", "(II)V", jni.IntValue(640), jni.IntValue(480)); err != nil {
		return fmt.Errorf("setVideoSize: %w", err)
	}
	if err := callRecVoid("setVideoFrameRate", "(I)V", jni.IntValue(30)); err != nil {
		return fmt.Errorf("setVideoFrameRate: %w", err)
	}
	if err := callRecVoid("setVideoEncodingBitRate", "(I)V", jni.IntValue(2_000_000)); err != nil {
		return fmt.Errorf("setVideoEncodingBitRate: %w", err)
	}

	// setOutputFile(String)
	err = vm.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(recObj)
		mid, err := env.GetMethodID(cls, "setOutputFile", "(Ljava/lang/String;)V")
		if err != nil {
			return fmt.Errorf("get setOutputFile: %w", err)
		}
		jPath, err := env.NewStringUTF(outputPath)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jPath.Object)
		return env.CallVoidMethod(recObj, mid, jni.ObjectValue(&jPath.Object))
	})
	if err != nil {
		return fmt.Errorf("setOutputFile: %w", err)
	}
	fmt.Fprintln(output, "Recorder configured")
	ui.RenderOutput()

	// 4. Prepare the recorder.
	if err := callRecVoid("prepare", "()V"); err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	fmt.Fprintln(output, "prepare() OK")
	ui.RenderOutput()

	// 5. Get the recorder's input surface as ANativeWindow.
	var recWindow *C.ANativeWindow
	err = vm.Do(func(env *jni.Env) error {
		cls := env.GetObjectClass(recObj)
		mid, err := env.GetMethodID(cls, "getSurface", "()Landroid/view/Surface;")
		if err != nil {
			return fmt.Errorf("get getSurface: %w", err)
		}
		surfObj, err := env.CallObjectMethod(recObj, mid)
		if err != nil {
			return fmt.Errorf("getSurface: %w", err)
		}
		if surfObj == nil || surfObj.Ref() == 0 {
			return fmt.Errorf("getSurface returned null")
		}
		recWindow = C.surfaceToWindow(env.Ptr(), unsafe.Pointer(surfObj.Ref()))
		if recWindow == nil {
			return fmt.Errorf("surfaceToWindow returned null")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("get surface: %w", err)
	}
	fmt.Fprintln(output, "Got recorder surface")
	ui.RenderOutput()

	// 6. Start the MediaRecorder.
	if err := callRecVoid("start", "()V"); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	fmt.Fprintln(output, "Recording started...")
	ui.RenderOutput()

	// 7. Open camera and start feeding frames.
	var camState C.CameraState
	ret := C.cameraOpen(&camState, recWindow)
	if ret != 0 {
		fmt.Fprintf(output, "NDK camera open err: %d\n", int(ret))
		ui.RenderOutput()
		// Try to stop the recorder anyway.
		callRecVoid("stop", "()V")
		callRecVoid("release", "()V")
		C.ANativeWindow_release(recWindow)
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(recObj)
			return nil
		})
		return fmt.Errorf("camera open failed: %d", int(ret))
	}
	fmt.Fprintln(output, "Camera streaming...")
	ui.RenderOutput()

	// 8. Record for 3 seconds.
	time.Sleep(3 * time.Second)

	// 9. Stop the MediaRecorder FIRST (while camera is still streaming).
	if err := callRecVoid("stop", "()V"); err != nil {
		fmt.Fprintf(output, "stop err: %v\n", err)
	} else {
		fmt.Fprintln(output, "Recording stopped")
	}
	ui.RenderOutput()

	// 10. Close the camera.
	C.cameraClose(&camState)
	fmt.Fprintln(output, "Camera closed")
	ui.RenderOutput()

	// 11. Release recorder and surface.
	callRecVoid("release", "()V")
	C.ANativeWindow_release(recWindow)
	vm.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(recObj)
		return nil
	})
	fmt.Fprintln(output, "Released resources")
	ui.RenderOutput()

	// 12. Report file size.
	var fileSize int64
	err = vm.Do(func(env *jni.Env) error {
		fileCls, err := env.FindClass("java/io/File")
		if err != nil {
			return fmt.Errorf("find File: %w", err)
		}
		initMid, err := env.GetMethodID(fileCls, "<init>", "(Ljava/lang/String;)V")
		if err != nil {
			return fmt.Errorf("get File.<init>: %w", err)
		}
		jPath, err := env.NewStringUTF(outputPath)
		if err != nil {
			return err
		}
		defer env.DeleteLocalRef(&jPath.Object)
		fileObj, err := env.NewObject(fileCls, initMid, jni.ObjectValue(&jPath.Object))
		if err != nil {
			return fmt.Errorf("new File: %w", err)
		}
		lengthMid, err := env.GetMethodID(fileCls, "length", "()J")
		if err != nil {
			return fmt.Errorf("get length: %w", err)
		}
		fileSize, err = env.CallLongMethod(fileObj, lengthMid)
		if err != nil {
			return fmt.Errorf("length(): %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("file size: %w", err)
	}

	fmt.Fprintf(output, "\nFile: %s\n", outputPath)
	fmt.Fprintf(output, "Size: %d bytes\n", fileSize)
	fmt.Fprintln(output, "\nCamera record complete.")
	return nil
}
