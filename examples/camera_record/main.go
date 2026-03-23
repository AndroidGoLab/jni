//go:build android

// Command camera_record demonstrates recording a short video from the
// camera using the Camera2 API + MediaRecorder — all via JNI from Go,
// with no NDK camera code. Uses the shared camera2 helper package for
// Camera2 session management.
package main

/*
#include <android/native_activity.h>

extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
extern void goOnNativeWindowCreated(ANativeActivity*, ANativeWindow*);
static void _onWindowCreated(ANativeActivity* a, ANativeWindow* w) { goOnNativeWindowCreated(a, w); }
static uintptr_t _getVM(ANativeActivity* a) { return (uintptr_t)a->vm; }
static uintptr_t _getClazz(ANativeActivity* a) { return (uintptr_t)a->clazz; }
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
	"github.com/AndroidGoLab/jni/examples/common/camera2"
	"github.com/AndroidGoLab/jni/examples/common/ui"
	"github.com/AndroidGoLab/jni/media/recorder"
)

func main() {}

func init() { ui.Register(run) }

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	vm := jni.VMFromUintptr(uintptr(C._getVM(activity)))
	actObj := jni.ObjectFromUintptr(uintptr(C._getClazz(activity)))
	ui.OnCreate(vm, actObj)

	vm.Do(func(env *jni.Env) error {
		return camera2.Init(env, actObj)
	})
	C._setCallbacks(activity)
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	ui.OnResume(jni.ObjectFromUintptr(uintptr(C._getClazz(activity))))
}

//export goOnNativeWindowCreated
func goOnNativeWindowCreated(activity *C.ANativeActivity, window *C.ANativeWindow) {
	ui.OnNativeWindowCreated(unsafe.Pointer(window))
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

	// 4. Get the recorder's input surface.
	surfObj, err := rec.GetSurface()
	if err != nil {
		return fmt.Errorf("getSurface: %w", err)
	}
	if surfObj == nil || surfObj.Ref() == 0 {
		return fmt.Errorf("getSurface returned null")
	}

	// 5. Start recording, then open camera.
	if err := rec.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	fmt.Fprintln(output, "Recording started")
	ui.RenderOutput()

	cam, err := camera2.Open(vm, ui.ActivityRef(), camera2.Config{
		Facing:   camera2.FacingBack,
		Template: camera2.TemplateRecord,
	}, surfObj)
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

	// 6. Record for 5 seconds.
	time.Sleep(5 * time.Second)

	// 7. Stop recorder, then close camera.
	stopT := time.Now()
	if err := rec.Stop(); err != nil {
		fmt.Fprintf(output, "stop err: %v\n", err)
	} else {
		fmt.Fprintln(output, "Recording stopped")
	}
	stopDur := time.Since(stopT)
	ui.RenderOutput()

	pprof.StopCPUProfile()
	profFile.Close()

	cam.Close()
	_ = rec.Release()
	vm.Do(func(env *jni.Env) error {
		env.DeleteGlobalRef(recObj)
		return nil
	})
	fmt.Fprintln(output, "Released resources")
	ui.RenderOutput()

	// 8. Report file sizes.
	videoInfo, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("stat video: %w", err)
	}
	profInfo, err := os.Stat(profilePath)
	if err != nil {
		return fmt.Errorf("stat profile: %w", err)
	}

	fmt.Fprintf(output, "\nVideo: %s (%d bytes, %.1f MB)\n",
		outputPath, videoInfo.Size(), float64(videoInfo.Size())/(1024*1024))
	fmt.Fprintf(output, "CPU profile: %s (%d bytes)\n", profilePath, profInfo.Size())
	fmt.Fprintf(output, "Timing: prepare=%v stop=%v\n", prepDur, stopDur)
	fmt.Fprintln(output, "\nCamera record complete (zero NDK).")
	return nil
}
