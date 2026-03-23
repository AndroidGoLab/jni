//go:build android

// Command camera_display shows a live camera preview on a SurfaceView
// created via JNI. Every 5 seconds it toggles between front and back
// cameras. All camera access is via the shared camera2 JNI helper.
//
// The SurfaceView approach: we create a SurfaceView programmatically
// via JNI, set it as the activity's content view, and use its Surface
// (via SurfaceHolder.Callback proxy) as the Camera2 output target.
package main

/*
#include <android/native_activity.h>

extern void goOnResume(ANativeActivity*);
static void _onResume(ANativeActivity* a) { goOnResume(a); }
static uintptr_t _getVM(ANativeActivity* a) { return (uintptr_t)a->vm; }
static uintptr_t _getClazz(ANativeActivity* a) { return (uintptr_t)a->clazz; }
static void _setCallbacks(ANativeActivity* a) {
    a->callbacks->onResume = _onResume;
}
*/
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/app"
	"github.com/AndroidGoLab/jni/content"
	"github.com/AndroidGoLab/jni/examples/common/camera2"
	"github.com/AndroidGoLab/jni/view/display"
)

func main() {}

var (
	mu             sync.Mutex
	globalVM       *jni.VM
	activityObj    *jni.Object
	currentSession *camera2.Session
	currentFacing  camera2.Facing
	surfaceCh      chan *jni.Object
)

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
	vm := jni.VMFromUintptr(uintptr(C._getVM(activity)))
	actObj := jni.ObjectFromUintptr(uintptr(C._getClazz(activity)))

	mu.Lock()
	globalVM = vm
	activityObj = actObj
	surfaceCh = make(chan *jni.Object, 1)
	mu.Unlock()

	vm.Do(func(env *jni.Env) error {
		if err := camera2.Init(env, actObj); err != nil {
			println("camera2.Init:", err.Error())
			return err
		}

		// Create a SurfaceView inside a FrameLayout with correct aspect
		// ratio, then set it as the activity's content view.
		//
		// TODO: needs wrapper — SurfaceView constructor not yet wrapped.
		svCls, err := env.FindClass("android/view/SurfaceView")
		if err != nil {
			return fmt.Errorf("find SurfaceView: %w", err)
		}
		svInit, err := env.GetMethodID(svCls, "<init>", "(Landroid/content/Context;)V")
		if err != nil {
			return fmt.Errorf("get SurfaceView.<init>: %w", err)
		}
		svLocal, err := env.NewObject(svCls, svInit, jni.ObjectValue(actObj))
		if err != nil {
			return fmt.Errorf("new SurfaceView: %w", err)
		}
		svGlobal := env.NewGlobalRef(svLocal)
		env.DeleteLocalRef(svLocal)

		sv := &display.SurfaceView{VM: vm, Obj: svGlobal}

		// Get screen dimensions to compute aspect-correct SurfaceView size.
		actCtx := &app.Context{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(actObj))}
		resObj, err := actCtx.GetResources()
		if err != nil {
			return fmt.Errorf("getResources: %w", err)
		}

		res := &content.Resources{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(resObj))}
		dmObj, err := res.GetDisplayMetrics()
		if err != nil {
			return fmt.Errorf("getDisplayMetrics: %w", err)
		}

		// TODO: needs wrapper — DisplayMetrics.widthPixels/heightPixels
		// are fields, not methods; the typed wrapper does not expose them.
		dmCls, err := env.FindClass("android/util/DisplayMetrics")
		if err != nil {
			return fmt.Errorf("find DisplayMetrics: %w", err)
		}
		widthFid, err := env.GetFieldID(dmCls, "widthPixels", "I")
		if err != nil {
			return fmt.Errorf("get widthPixels: %w", err)
		}
		heightFid, err := env.GetFieldID(dmCls, "heightPixels", "I")
		if err != nil {
			return fmt.Errorf("get heightPixels: %w", err)
		}
		screenW := env.GetIntField(dmObj, widthFid)
		screenH := env.GetIntField(dmObj, heightFid)

		// Camera outputs landscape 16:9. In portrait, the preview is
		// rotated so the effective aspect ratio is 9:16 (matching the
		// phone's tall screen). Fit to screen width, scale height.
		const cameraW, cameraH = 1080, 1920 // portrait orientation
		cameraAR := float64(cameraW) / float64(cameraH)
		screenAR := float64(screenW) / float64(screenH)

		var viewW, viewH int32
		if screenAR > cameraAR {
			viewH = screenH
			viewW = int32(float64(screenH) * cameraAR)
		} else {
			viewW = screenW
			viewH = int32(float64(screenW) / cameraAR)
		}

		// Create FrameLayout as root, add SurfaceView with centered layout.
		//
		// TODO: needs wrapper — FrameLayout constructor not yet wrapped.
		flCls, err := env.FindClass("android/widget/FrameLayout")
		if err != nil {
			return fmt.Errorf("find FrameLayout: %w", err)
		}
		flInit, err := env.GetMethodID(flCls, "<init>", "(Landroid/content/Context;)V")
		if err != nil {
			return fmt.Errorf("get FrameLayout.<init>: %w", err)
		}
		flLocal, err := env.NewObject(flCls, flInit, jni.ObjectValue(actObj))
		if err != nil {
			return fmt.Errorf("new FrameLayout: %w", err)
		}
		flGlobal := env.NewGlobalRef(flLocal)
		env.DeleteLocalRef(flLocal)

		// Set black background on the FrameLayout (cast to View).
		flView := &display.View{VM: vm, Obj: flGlobal}
		if err := flView.SetBackgroundColor(-16777216); err != nil { // Color.BLACK
			return fmt.Errorf("setBackgroundColor: %w", err)
		}

		// Create FrameLayout.LayoutParams(viewW, viewH, Gravity.CENTER=17).
		//
		// TODO: needs wrapper — FrameLayout.LayoutParams constructor not yet wrapped.
		lpCls, err := env.FindClass("android/widget/FrameLayout$LayoutParams")
		if err != nil {
			return fmt.Errorf("find FrameLayout.LayoutParams: %w", err)
		}
		lpInit, err := env.GetMethodID(lpCls, "<init>", "(III)V")
		if err != nil {
			return fmt.Errorf("get LayoutParams.<init>: %w", err)
		}
		const gravityCenter = 17
		lp, err := env.NewObject(lpCls, lpInit,
			jni.IntValue(viewW), jni.IntValue(viewH), jni.IntValue(gravityCenter))
		if err != nil {
			return fmt.Errorf("new LayoutParams: %w", err)
		}

		// frameLayout.addView(surfaceView, layoutParams) — cast to ViewGroup.
		flViewGroup := &display.ViewGroup{VM: vm, Obj: flGlobal}
		if err := flViewGroup.AddView2_1(svGlobal, lp); err != nil {
			return fmt.Errorf("addView: %w", err)
		}

		// activity.setContentView(frameLayout)
		act := &app.Activity{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(actObj))}
		if err := act.SetContentView1(flGlobal); err != nil {
			return fmt.Errorf("setContentView: %w", err)
		}

		// Get SurfaceHolder from the SurfaceView.
		holderObj, err := sv.GetHolder()
		if err != nil {
			return fmt.Errorf("getHolder: %w", err)
		}
		sh := &display.SurfaceHolder{VM: vm, Obj: (*jni.GlobalRef)(unsafe.Pointer(holderObj))}

		// Register SurfaceHolder.Callback via interface proxy.
		//
		// TODO: needs wrapper — SurfaceHolder.Callback proxy creation
		// requires env.NewProxy which has no typed wrapper.
		holderCallbackCls, err := env.FindClass("android/view/SurfaceHolder$Callback")
		if err != nil {
			return fmt.Errorf("find SurfaceHolder.Callback: %w", err)
		}

		ch := surfaceCh
		proxy, _, err := env.NewProxy(
			[]*jni.Class{holderCallbackCls},
			func(env *jni.Env, method string, args []*jni.Object) (*jni.Object, error) {
				switch method {
				case "surfaceCreated":
					// args[0] is the SurfaceHolder. Use typed wrapper for getSurface.
					argGlobal := env.NewGlobalRef(args[0])
					cbHolder := &display.SurfaceHolder{
						VM:  vm,
						Obj: argGlobal,
					}
					surf, err := cbHolder.GetSurface()
					if err != nil {
						println("getSurface:", err.Error())
						return nil, nil
					}
					// surf is already a global ref (the wrapper converts it).
					select {
					case ch <- surf:
					default:
					}
				case "surfaceDestroyed":
					// Camera will be cleaned up when the activity stops.
				}
				return nil, nil
			},
		)
		if err != nil {
			return fmt.Errorf("NewProxy(SurfaceHolder.Callback): %w", err)
		}

		// holder.addCallback(proxy)
		if err := sh.AddCallback(proxy); err != nil {
			return fmt.Errorf("addCallback: %w", err)
		}

		// Set fixed surface size to maintain camera aspect ratio (16:9).
		return sh.SetFixedSize(1920, 1080)
	})

	C._setCallbacks(activity)

	// Start the camera preview loop in a goroutine.
	go runPreview()
}

//export goOnResume
func goOnResume(activity *C.ANativeActivity) {
	// No-op — lifecycle is managed by surfaceCreated callback.
}

func runPreview() {
	// Wait for the SurfaceView's Surface to be ready.
	mu.Lock()
	ch := surfaceCh
	mu.Unlock()

	surface := <-ch
	println("camera_display: Surface ready, starting camera")

	mu.Lock()
	currentFacing = camera2.FacingBack
	mu.Unlock()

	if err := openCamera(surface); err != nil {
		println("camera open error:", err.Error())
		return
	}

	// Toggle front/back every 5 seconds.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mu.Lock()
		if currentFacing == camera2.FacingBack {
			currentFacing = camera2.FacingFront
		} else {
			currentFacing = camera2.FacingBack
		}
		facing := currentFacing
		mu.Unlock()

		facingName := "BACK"
		if facing == camera2.FacingFront {
			facingName = "FRONT"
		}
		println(fmt.Sprintf("camera_display: switching to %s camera", facingName))

		if err := reopenCamera(surface); err != nil {
			println("camera reopen error:", err.Error())
		}
	}
}

func openCamera(surface *jni.Object) error {
	mu.Lock()
	vm := globalVM
	activity := activityObj
	facing := currentFacing
	mu.Unlock()

	session, err := camera2.Open(vm, activity, camera2.Config{
		Facing:   facing,
		Template: camera2.TemplatePreview,
	}, surface)
	if err != nil {
		return err
	}

	mu.Lock()
	currentSession = session
	mu.Unlock()

	facingName := "BACK"
	if facing == camera2.FacingFront {
		facingName = "FRONT"
	}
	println(fmt.Sprintf("camera_display: Camera opened (%s)", facingName))
	return nil
}

func reopenCamera(surface *jni.Object) error {
	mu.Lock()
	old := currentSession
	currentSession = nil
	mu.Unlock()

	if old != nil {
		old.Close()
	}
	return openCamera(surface)
}
