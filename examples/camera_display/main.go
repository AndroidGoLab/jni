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
	"github.com/AndroidGoLab/jni/examples/common/camera2"
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
	vm := jni.VMFromPtr(unsafe.Pointer(activity.vm))
	actObj := jni.ObjectFromPtr(unsafe.Pointer(activity.clazz))

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
		svCls, err := env.FindClass("android/view/SurfaceView")
		if err != nil {
			return fmt.Errorf("find SurfaceView: %w", err)
		}
		svInit, err := env.GetMethodID(svCls, "<init>", "(Landroid/content/Context;)V")
		if err != nil {
			return fmt.Errorf("get SurfaceView.<init>: %w", err)
		}
		sv, err := env.NewObject(svCls, svInit, jni.ObjectValue(actObj))
		if err != nil {
			return fmt.Errorf("new SurfaceView: %w", err)
		}

		// Get screen dimensions to compute aspect-correct SurfaceView size.
		resCls, err := env.FindClass("android/content/res/Resources")
		if err != nil {
			return fmt.Errorf("find Resources: %w", err)
		}
		actCls := env.GetObjectClass(actObj)
		getResMid, err := env.GetMethodID(actCls, "getResources", "()Landroid/content/res/Resources;")
		if err != nil {
			return fmt.Errorf("get getResources: %w", err)
		}
		res, err := env.CallObjectMethod(actObj, getResMid)
		if err != nil {
			return fmt.Errorf("getResources: %w", err)
		}
		dmCls, err := env.FindClass("android/util/DisplayMetrics")
		if err != nil {
			return fmt.Errorf("find DisplayMetrics: %w", err)
		}
		getMetricsMid, err := env.GetMethodID(resCls, "getDisplayMetrics", "()Landroid/util/DisplayMetrics;")
		if err != nil {
			return fmt.Errorf("get getDisplayMetrics: %w", err)
		}
		dm, err := env.CallObjectMethod(res, getMetricsMid)
		if err != nil {
			return fmt.Errorf("getDisplayMetrics: %w", err)
		}
		widthFid, err := env.GetFieldID(dmCls, "widthPixels", "I")
		if err != nil {
			return fmt.Errorf("get widthPixels: %w", err)
		}
		heightFid, err := env.GetFieldID(dmCls, "heightPixels", "I")
		if err != nil {
			return fmt.Errorf("get heightPixels: %w", err)
		}
		screenW := env.GetIntField(dm, widthFid)
		screenH := env.GetIntField(dm, heightFid)

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
		flCls, err := env.FindClass("android/widget/FrameLayout")
		if err != nil {
			return fmt.Errorf("find FrameLayout: %w", err)
		}
		flInit, err := env.GetMethodID(flCls, "<init>", "(Landroid/content/Context;)V")
		if err != nil {
			return fmt.Errorf("get FrameLayout.<init>: %w", err)
		}
		fl, err := env.NewObject(flCls, flInit, jni.ObjectValue(actObj))
		if err != nil {
			return fmt.Errorf("new FrameLayout: %w", err)
		}

		// Set black background on the FrameLayout.
		viewCls, err := env.FindClass("android/view/View")
		if err != nil {
			return fmt.Errorf("find View: %w", err)
		}
		setBgMid, err := env.GetMethodID(viewCls, "setBackgroundColor", "(I)V")
		if err != nil {
			return fmt.Errorf("get setBackgroundColor: %w", err)
		}
		env.CallVoidMethod(fl, setBgMid, jni.IntValue(-16777216)) // Color.BLACK

		// Create FrameLayout.LayoutParams(viewW, viewH, Gravity.CENTER=17).
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

		// frameLayout.addView(surfaceView, layoutParams)
		addViewMid, err := env.GetMethodID(flCls, "addView",
			"(Landroid/view/View;Landroid/view/ViewGroup$LayoutParams;)V")
		if err != nil {
			return fmt.Errorf("get addView: %w", err)
		}
		if err := env.CallVoidMethod(fl, addViewMid, jni.ObjectValue(sv), jni.ObjectValue(lp)); err != nil {
			return fmt.Errorf("addView: %w", err)
		}

		// activity.setContentView(frameLayout)
		setContentMid, err := env.GetMethodID(actCls, "setContentView", "(Landroid/view/View;)V")
		if err != nil {
			return fmt.Errorf("get setContentView: %w", err)
		}
		if err := env.CallVoidMethod(actObj, setContentMid, jni.ObjectValue(fl)); err != nil {
			return fmt.Errorf("setContentView: %w", err)
		}

		// Get SurfaceHolder from the SurfaceView.
		getHolderMid, err := env.GetMethodID(svCls, "getHolder", "()Landroid/view/SurfaceHolder;")
		if err != nil {
			return fmt.Errorf("get getHolder: %w", err)
		}
		holder, err := env.CallObjectMethod(sv, getHolderMid)
		if err != nil {
			return fmt.Errorf("getHolder: %w", err)
		}

		// Register SurfaceHolder.Callback via interface proxy.
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
					// args[0] is the SurfaceHolder. Get its Surface.
					holderCls := env.GetObjectClass(args[0])
					getSurfMid, err := env.GetMethodID(holderCls, "getSurface", "()Landroid/view/Surface;")
					if err != nil {
						println("getSurface method:", err.Error())
						return nil, nil
					}
					surf, err := env.CallObjectMethod(args[0], getSurfMid)
					if err != nil {
						println("getSurface call:", err.Error())
						return nil, nil
					}
					globalSurf := env.NewGlobalRef(surf)
					select {
					case ch <- globalSurf:
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
		holderCls := env.GetObjectClass(holder)
		addCallbackMid, err := env.GetMethodID(holderCls, "addCallback",
			"(Landroid/view/SurfaceHolder$Callback;)V")
		if err != nil {
			return fmt.Errorf("get addCallback: %w", err)
		}
		if err := env.CallVoidMethod(holder, addCallbackMid, jni.ObjectValue(proxy)); err != nil {
			return fmt.Errorf("addCallback: %w", err)
		}

		// Set fixed surface size to maintain camera aspect ratio (16:9).
		setFixedMid, err := env.GetMethodID(holderCls, "setFixedSize", "(II)V")
		if err != nil {
			return fmt.Errorf("get setFixedSize: %w", err)
		}
		return env.CallVoidMethod(holder, setFixedMid, jni.IntValue(1920), jni.IntValue(1080))
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
