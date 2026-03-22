//go:build android

// Package camera2 provides a shared helper for opening an Android camera
// via the Camera2 Java API through JNI. It handles camera selection by
// facing direction, session setup with proxy callbacks, and repeating
// capture requests -- all using Handler-based overloads (API 21+).
package camera2

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

const templateRecord = 3 // CameraDevice.TEMPLATE_RECORD

// Facing selects front or back camera.
type Facing int32

const (
	FacingFront Facing = 0
	FacingBack  Facing = 1
)

// Config controls camera session parameters.
type Config struct {
	Facing Facing
	Width  int32 // 0 = camera default
	Height int32 // 0 = camera default
}

// Session holds an active camera capture session.
type Session struct {
	vm           *jni.VM
	device       *jni.Object
	session      *jni.Object
	devHandlerID int64
	sesHandlerID int64
}

// Cached class references and main-thread Handler for Camera2 callbacks.
var (
	initOnce           sync.Once
	initErr            error
	clsDeviceCallback  *jni.Object
	clsSessionCallback *jni.Object
	mainHandler        *jni.Object
)

// Init caches proxy adapter classes and creates the main-thread Handler.
// Must be called once from the main thread (ANativeActivity_onCreate)
// before Open is used. Uses sync.Once so repeated calls are safe.
func Init(env *jni.Env, activity *jni.Object) error {
	initOnce.Do(func() {
		initErr = doInit(env, activity)
	})
	return initErr
}

func doInit(env *jni.Env, activity *jni.Object) error {
	// Get the activity's ClassLoader so we can load APK classes.
	cls := env.GetObjectClass(activity)
	mid, err := env.GetMethodID(cls, "getClassLoader", "()Ljava/lang/ClassLoader;")
	if err != nil {
		return fmt.Errorf("camera2: get getClassLoader: %w", err)
	}
	cl, err := env.CallObjectMethod(activity, mid)
	if err != nil {
		return fmt.Errorf("camera2: getClassLoader: %w", err)
	}
	globalCL := env.NewGlobalRef(cl)
	jni.SetProxyClassLoader(globalCL)

	// Load adapter classes via ClassLoader.loadClass() since FindClass
	// uses the boot ClassLoader in NativeActivity context.
	clsCL := env.GetObjectClass(cl)
	loadClassMid, err := env.GetMethodID(clsCL, "loadClass", "(Ljava/lang/String;)Ljava/lang/Class;")
	if err != nil {
		return fmt.Errorf("camera2: get loadClass: %w", err)
	}

	loadClass := func(name string) (*jni.Object, error) {
		jName, err := env.NewStringUTF(name)
		if err != nil {
			return nil, fmt.Errorf("camera2: newStringUTF(%s): %w", name, err)
		}
		defer env.DeleteLocalRef(&jName.Object)
		obj, err := env.CallObjectMethod(cl, loadClassMid, jni.ObjectValue(&jName.Object))
		if err != nil {
			return nil, fmt.Errorf("camera2: loadClass(%s): %w", name, err)
		}
		if obj == nil {
			return nil, fmt.Errorf("camera2: loadClass(%s) returned null", name)
		}
		return env.NewGlobalRef(obj), nil
	}

	clsDeviceCallback, err = loadClass("center.dx.jni.generated.CameraDeviceCallback")
	if err != nil {
		return err
	}
	clsSessionCallback, err = loadClass("center.dx.jni.generated.CaptureSessionCallback")
	if err != nil {
		return err
	}

	// Create a Handler on the main Looper for Camera2 callbacks.
	looperCls, err := env.FindClass("android/os/Looper")
	if err != nil {
		return fmt.Errorf("camera2: find Looper: %w", err)
	}
	getMainMid, err := env.GetStaticMethodID(looperCls, "getMainLooper", "()Landroid/os/Looper;")
	if err != nil {
		return fmt.Errorf("camera2: get getMainLooper: %w", err)
	}
	mainLooper, err := env.CallStaticObjectMethod(looperCls, getMainMid)
	if err != nil {
		return fmt.Errorf("camera2: getMainLooper: %w", err)
	}
	handlerCls, err := env.FindClass("android/os/Handler")
	if err != nil {
		return fmt.Errorf("camera2: find Handler: %w", err)
	}
	handlerInit, err := env.GetMethodID(handlerCls, "<init>", "(Landroid/os/Looper;)V")
	if err != nil {
		return fmt.Errorf("camera2: get Handler.<init>: %w", err)
	}
	handler, err := env.NewObject(handlerCls, handlerInit, jni.ObjectValue(mainLooper))
	if err != nil {
		return fmt.Errorf("camera2: new Handler: %w", err)
	}
	mainHandler = env.NewGlobalRef(handler)

	if err := jni.EnsureProxyInit(env); err != nil {
		return fmt.Errorf("camera2: EnsureProxyInit: %w", err)
	}
	return nil
}

// Open finds a camera matching cfg.Facing, opens it, creates a capture
// session targeting the given surfaces, and starts a repeating TEMPLATE_RECORD
// request. Blocks until the session is configured.
func Open(vm *jni.VM, activity *jni.Object, cfg Config, targets ...*jni.Object) (*Session, error) {
	if initErr != nil {
		return nil, fmt.Errorf("camera2: Init failed: %w", initErr)
	}
	if mainHandler == nil {
		return nil, fmt.Errorf("camera2: Init was not called")
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("camera2: Open requires at least one target surface")
	}

	s := &Session{vm: vm}

	// ---------- Find matching camera ID ----------
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

		// Prepare for reading LENS_FACING from CameraCharacteristics.
		getCharsMid, err := env.GetMethodID(mgrCls, "getCameraCharacteristics",
			"(Ljava/lang/String;)Landroid/hardware/camera2/CameraCharacteristics;")
		if err != nil {
			return fmt.Errorf("get getCameraCharacteristics: %w", err)
		}

		// Iterate camera IDs and match by facing.
		desiredFacing := int32(cfg.Facing)
		for i := int32(0); i < length; i++ {
			idObj, err := env.GetObjectArrayElement((*jni.ObjectArray)(unsafe.Pointer(idsObj)), i)
			if err != nil {
				continue
			}
			idStr := env.GoString((*jni.String)(unsafe.Pointer(idObj)))

			// Get characteristics for this camera.
			chars, err := env.CallObjectMethod(mgrObj, getCharsMid, jni.ObjectValue(idObj))
			if err != nil || chars == nil {
				continue
			}

			// Read LENS_FACING via CameraCharacteristics.get(Key).
			charsCls := env.GetObjectClass(chars)
			getMid, err := env.GetMethodID(charsCls, "get",
				"(Landroid/hardware/camera2/CameraCharacteristics$Key;)Ljava/lang/Object;")
			if err != nil {
				continue
			}
			lensFacingFid, err := env.GetStaticFieldID(charsCls, "LENS_FACING",
				"Landroid/hardware/camera2/CameraCharacteristics$Key;")
			if err != nil {
				continue
			}
			lensFacingKey := env.GetStaticObjectField(charsCls, lensFacingFid)
			if lensFacingKey == nil {
				continue
			}
			facingObj, err := env.CallObjectMethod(chars, getMid, jni.ObjectValue(lensFacingKey))
			if err != nil || facingObj == nil {
				continue
			}

			// Unbox Integer to int.
			intCls := env.GetObjectClass(facingObj)
			intValueMid, err := env.GetMethodID(intCls, "intValue", "()I")
			if err != nil {
				continue
			}
			facing, err := env.CallIntMethod(facingObj, intValueMid)
			if err != nil {
				continue
			}

			if facing == desiredFacing {
				cameraID = idStr
				break
			}
		}

		if cameraID == "" {
			// Fallback: use the first camera if no facing match.
			firstID, err := env.GetObjectArrayElement((*jni.ObjectArray)(unsafe.Pointer(idsObj)), 0)
			if err != nil {
				return fmt.Errorf("get first camera ID: %w", err)
			}
			cameraID = env.GoString((*jni.String)(unsafe.Pointer(firstID)))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// ---------- Open camera ----------
	type deviceResult struct {
		device *jni.Object
		err    error
	}
	deviceCh := make(chan deviceResult, 1)
	s.devHandlerID = jni.RegisterProxyHandler(func(env *jni.Env, method string, args []*jni.Object) (*jni.Object, error) {
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
		cbObj, err := env.NewObject(cbCls, cbInit, jni.LongValue(s.devHandlerID))
		if err != nil {
			return fmt.Errorf("new CameraDeviceCallback: %w", err)
		}

		// Get CameraManager again (local refs from previous Do are gone).
		mgrCls, _ := env.FindClass("android/hardware/camera2/CameraManager")
		ctxCls := env.GetObjectClass(activity)
		getSysMid, _ := env.GetMethodID(ctxCls, "getSystemService", "(Ljava/lang/String;)Ljava/lang/Object;")
		svcName, _ := env.NewStringUTF("camera")
		defer env.DeleteLocalRef(&svcName.Object)
		mgrObj, _ := env.CallObjectMethod(activity, getSysMid, jni.ObjectValue(&svcName.Object))

		// CameraManager.openCamera(String, StateCallback, Handler)
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
			jni.ObjectValue(mainHandler))
	})
	if err != nil {
		jni.UnregisterProxyHandler(s.devHandlerID)
		return nil, fmt.Errorf("openCamera: %w", err)
	}

	// Wait for camera to open.
	dr := <-deviceCh
	if dr.err != nil {
		jni.UnregisterProxyHandler(s.devHandlerID)
		return nil, fmt.Errorf("camera open: %w", dr.err)
	}
	s.device = dr.device

	// ---------- Create capture session ----------
	type sessionResult struct {
		session *jni.Object
		err     error
	}
	sessionCh := make(chan sessionResult, 1)
	s.sesHandlerID = jni.RegisterProxyHandler(func(env *jni.Env, method string, args []*jni.Object) (*jni.Object, error) {
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
		scObj, err := env.NewObject(scCls, scInit, jni.LongValue(s.sesHandlerID))
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
		for _, t := range targets {
			env.CallBooleanMethod(listObj, addMid, jni.ObjectValue(t))
		}

		// CameraDevice.createCaptureSession(List, StateCallback, Handler)
		devCls := env.GetObjectClass(s.device)
		createSessMid, err := env.GetMethodID(devCls, "createCaptureSession",
			"(Ljava/util/List;Landroid/hardware/camera2/CameraCaptureSession$StateCallback;Landroid/os/Handler;)V")
		if err != nil {
			return fmt.Errorf("get createCaptureSession: %w", err)
		}
		return env.CallVoidMethod(s.device, createSessMid,
			jni.ObjectValue(listObj),
			jni.ObjectValue(scObj),
			jni.ObjectValue(mainHandler))
	})
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("createCaptureSession: %w", err)
	}

	// Wait for session.
	sr := <-sessionCh
	if sr.err != nil {
		s.Close()
		return nil, fmt.Errorf("capture session: %w", sr.err)
	}
	s.session = sr.session

	// ---------- Start repeating request ----------
	err = vm.Do(func(env *jni.Env) error {
		devCls := env.GetObjectClass(s.device)

		// createCaptureRequest(TEMPLATE_RECORD) -> CaptureRequest.Builder
		createReqMid, err := env.GetMethodID(devCls, "createCaptureRequest",
			"(I)Landroid/hardware/camera2/CaptureRequest$Builder;")
		if err != nil {
			return fmt.Errorf("get createCaptureRequest: %w", err)
		}
		builder, err := env.CallObjectMethod(s.device, createReqMid, jni.IntValue(templateRecord))
		if err != nil {
			return fmt.Errorf("createCaptureRequest: %w", err)
		}

		// Builder.addTarget(surface) for each target.
		builderCls := env.GetObjectClass(builder)
		addTargetMid, err := env.GetMethodID(builderCls, "addTarget", "(Landroid/view/Surface;)V")
		if err != nil {
			return fmt.Errorf("get addTarget: %w", err)
		}
		for _, t := range targets {
			if err := env.CallVoidMethod(builder, addTargetMid, jni.ObjectValue(t)); err != nil {
				return fmt.Errorf("addTarget: %w", err)
			}
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

		// session.setRepeatingRequest(request, null, handler)
		sessCls := env.GetObjectClass(s.session)
		setRepMid, err := env.GetMethodID(sessCls, "setRepeatingRequest",
			"(Landroid/hardware/camera2/CaptureRequest;Landroid/hardware/camera2/CameraCaptureSession$CaptureCallback;Landroid/os/Handler;)I")
		if err != nil {
			return fmt.Errorf("get setRepeatingRequest: %w", err)
		}
		_, err = env.CallIntMethod(s.session, setRepMid,
			jni.ObjectValue(request),
			jni.ObjectValue(nil),
			jni.ObjectValue(mainHandler))
		if err != nil {
			return fmt.Errorf("setRepeatingRequest: %w", err)
		}
		return nil
	})
	if err != nil {
		s.Close()
		return nil, err
	}

	return s, nil
}

// Close stops the capture session and releases all camera resources.
func (s *Session) Close() {
	// Stop repeating and close session.
	if s.session != nil {
		s.vm.Do(func(env *jni.Env) error {
			cls := env.GetObjectClass(s.session)
			if mid, err := env.GetMethodID(cls, "stopRepeating", "()V"); err == nil {
				env.CallVoidMethod(s.session, mid)
			}
			if mid, err := env.GetMethodID(cls, "close", "()V"); err == nil {
				env.CallVoidMethod(s.session, mid)
			}
			return nil
		})
		// Small delay to let in-flight frames drain.
		time.Sleep(100 * time.Millisecond)
		s.vm.Do(func(env *jni.Env) error {
			env.DeleteGlobalRef(s.session)
			return nil
		})
	}

	// Close camera device.
	if s.device != nil {
		s.vm.Do(func(env *jni.Env) error {
			cls := env.GetObjectClass(s.device)
			if mid, err := env.GetMethodID(cls, "close", "()V"); err == nil {
				env.CallVoidMethod(s.device, mid)
			}
			env.DeleteGlobalRef(s.device)
			return nil
		})
	}

	jni.UnregisterProxyHandler(s.devHandlerID)
	jni.UnregisterProxyHandler(s.sesHandlerID)
}
