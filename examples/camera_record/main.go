//go:build android

// Command camera_record demonstrates recording a short video from the
// camera using the Camera2 API + MediaRecorder via JNI. It uses
// setVideoSource(SURFACE) to get a recording surface from MediaRecorder,
// then opens a Camera2 device via the NDK and feeds camera frames into
// that surface. Records for 5 seconds at 1080p60 with CPU profiling and
// reports the resulting file sizes.
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
	"os"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/media/recorder"
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

func run(vm *jni.VM, output *bytes.Buffer) error {
	fmt.Fprintln(output, "=== Camera Record ===")
	ui.RenderOutput()

	// 1. Get the app context and cache directory path.
	ctx, err := ui.GetAppContext(vm)
	if err != nil {
		return fmt.Errorf("get context: %w", err)
	}
	fmt.Fprintln(output, "Got app context")
	ui.RenderOutput()

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

	// 2. Create and configure MediaRecorder with SURFACE video source.
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

	// Configure: sources -> format -> encoders -> params -> file.
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

	// 4. Get the recorder's input surface as ANativeWindow.
	surfObj, err := rec.GetSurface()
	if err != nil {
		return fmt.Errorf("getSurface: %w", err)
	}
	if surfObj == nil || surfObj.Ref() == 0 {
		return fmt.Errorf("getSurface returned null")
	}
	var recWindow *C.ANativeWindow
	vm.Do(func(env *jni.Env) error {
		recWindow = C.surfaceToWindow(env.Ptr(), unsafe.Pointer(surfObj.Ref()))
		return nil
	})
	if recWindow == nil {
		return fmt.Errorf("surfaceToWindow returned null")
	}
	fmt.Fprintln(output, "Got recorder surface")
	ui.RenderOutput()

	// 5. Start the MediaRecorder.
	startT := time.Now()
	if err := rec.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	startDur := time.Since(startT)
	fmt.Fprintf(output, "Recording started (%v)\n", startDur)
	ui.RenderOutput()

	// 6. Open camera and start feeding frames.
	var camState C.CameraState
	ret := C.cameraOpen(&camState, recWindow)
	if ret != 0 {
		fmt.Fprintf(output, "NDK camera open err: %d\n", int(ret))
		ui.RenderOutput()
		_ = rec.Stop()
		_ = rec.Release()
		C.ANativeWindow_release(recWindow)
		vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(recObj)
			return nil
		})
		return fmt.Errorf("camera open failed: %d", int(ret))
	}
	fmt.Fprintln(output, "Camera streaming...")
	ui.RenderOutput()

	// 7. Record for 5 seconds.
	time.Sleep(5 * time.Second)

	// 8. Stop the MediaRecorder FIRST (while camera is still streaming).
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

	// 9. Close the camera.
	C.cameraClose(&camState)
	fmt.Fprintln(output, "Camera closed")
	ui.RenderOutput()

	// 10. Release recorder and surface.
	_ = rec.Release()
	C.ANativeWindow_release(recWindow)
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
	fmt.Fprintln(output, "\nCamera record complete.")
	return nil
}
