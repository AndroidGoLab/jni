package jni_raw

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/xaionaro-go/jni"
	pb "github.com/xaionaro-go/jni/proto/jni_raw"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// voidDetector caches JNI class/method/field IDs needed to detect whether a
// java.lang.reflect.Method returns void. The lookups are performed once via
// sync.Once so that subsequent proxy callbacks avoid repeated FindClass /
// GetMethodID / GetStaticFieldID calls on hot paths.
type voidDetector struct {
	once sync.Once
	err  error

	getReturnMID jni.MethodID // Method.getReturnType() -> Class
	voidType     *jni.Object  // cached Void.TYPE value
}

func (d *voidDetector) init(env *jni.Env) error {
	d.once.Do(func() {
		d.err = d.doInit(env)
	})
	return d.err
}

func (d *voidDetector) doInit(env *jni.Env) error {
	methodCls, err := env.FindClass("java/lang/reflect/Method")
	if err != nil {
		return fmt.Errorf("finding java.lang.reflect.Method: %w", err)
	}

	d.getReturnMID, err = env.GetMethodID(methodCls, "getReturnType", "()Ljava/lang/Class;")
	if err != nil {
		return fmt.Errorf("finding Method.getReturnType: %w", err)
	}

	voidCls, err := env.FindClass("java/lang/Void")
	if err != nil {
		return fmt.Errorf("finding java.lang.Void: %w", err)
	}

	typeFID, err := env.GetStaticFieldID(voidCls, "TYPE", "Ljava/lang/Class;")
	if err != nil {
		return fmt.Errorf("finding Void.TYPE: %w", err)
	}

	d.voidType = env.GetStaticObjectField(voidCls, typeFID)
	if d.voidType == nil || d.voidType.Ref() == 0 {
		return fmt.Errorf("Void.TYPE is null")
	}

	return nil
}

// isVoid returns true when method's return type is java.lang.Void.TYPE.
func (d *voidDetector) isVoid(
	env *jni.Env,
	method *jni.Object,
) bool {
	if method == nil {
		return true
	}

	if err := d.init(env); err != nil {
		// Cannot determine return type; assume non-void to be safe.
		return false
	}

	retType, err := env.CallObjectMethod(method, d.getReturnMID)
	if err != nil || retType == nil || retType.Ref() == 0 {
		return false
	}

	return env.IsSameObject(retType, d.voidType)
}

// Proxy implements the bidirectional streaming RPC that creates a Java
// dynamic proxy and forwards method invocations to the gRPC client as
// CallbackEvent messages. The client responds with CallbackResponse
// messages that are dispatched back to the blocked JVM callback thread.
func (s *Server) Proxy(stream pb.JNIService_ProxyServer) error {
	// 1. Read the first message -- must be CreateProxyRequest.
	firstMsg, err := stream.Recv()
	if err != nil {
		return err
	}
	createReq := firstMsg.GetCreate()
	if createReq == nil {
		return status.Error(codes.InvalidArgument, "first message must be CreateProxyRequest")
	}

	// 2. Resolve interface classes from handles.
	ifaceHandles := createReq.GetInterfaceClassHandles()
	ifaces := make([]*jni.Class, len(ifaceHandles))
	for i, h := range ifaceHandles {
		cls, err := s.requireClass(h)
		if err != nil {
			return err
		}
		ifaces[i] = cls
	}

	// 3. Pending callbacks: map callback_id -> response channel.
	var (
		pendingMu sync.Mutex
		pending   = map[int64]chan *pb.CallbackResponse{}
		nextID    atomic.Int64
	)

	// 4. Mutex protecting stream.Send, which may be called concurrently
	// from multiple JVM callback threads.
	var sendMu sync.Mutex

	// 5. Cache for void return type detection (used only for interface proxies).
	var detector voidDetector

	// sendCallback is the shared logic for forwarding a Java callback over the
	// gRPC stream to the client. It blocks until the client responds (for
	// non-void methods) or fires-and-forgets (for void methods).
	// The expectsResponse parameter tells the caller whether the method returns
	// a value; for abstract class adapters (which don't pass a Method object)
	// we always expect a response and let the client decide.
	sendCallback := func(
		env *jni.Env,
		methodName string,
		args []*jni.Object,
		expectsResponse bool,
	) (*jni.Object, error) {
		callbackID := nextID.Add(1)

		// Store args in HandleStore so client can reference them.
		argHandles := make([]int64, len(args))
		for i, arg := range args {
			if arg != nil && arg.Ref() != 0 {
				argHandles[i] = s.Handles.Put(env, arg)
			}
		}

		// Send callback event to client.
		event := &pb.ProxyServerMessage{
			Msg: &pb.ProxyServerMessage_Callback{
				Callback: &pb.CallbackEvent{
					CallbackId:      callbackID,
					MethodName:      methodName,
					ArgHandles:      argHandles,
					ExpectsResponse: expectsResponse,
				},
			},
		}

		sendMu.Lock()
		sendErr := stream.Send(event)
		sendMu.Unlock()
		if sendErr != nil {
			return nil, fmt.Errorf("sending callback event: %w", sendErr)
		}

		// If void, fire-and-forget.
		if !expectsResponse {
			return nil, nil
		}

		// Non-void: block until client responds.
		ch := make(chan *pb.CallbackResponse, 1)
		pendingMu.Lock()
		pending[callbackID] = ch
		pendingMu.Unlock()

		defer func() {
			pendingMu.Lock()
			delete(pending, callbackID)
			pendingMu.Unlock()
		}()

		resp, ok := <-ch
		if !ok {
			return nil, fmt.Errorf("stream closed while waiting for callback response")
		}

		if resp.GetError() != "" {
			return nil, fmt.Errorf("client error: %s", resp.GetError())
		}

		resultHandle := resp.GetResultHandle()
		if resultHandle == 0 {
			return nil, nil
		}
		return s.Handles.Get(resultHandle), nil
	}

	// 6. Determine whether the first class is an interface or an abstract class,
	// then create the proxy object using the appropriate mechanism.
	var (
		proxyObj     *jni.Object
		proxyCleanup func()
		proxyHandle  int64
	)

	var isInterface bool
	if err := s.VM.Do(func(env *jni.Env) error {
		classCls, findErr := env.FindClass("java/lang/Class")
		if findErr != nil {
			return fmt.Errorf("finding java.lang.Class: %w", findErr)
		}
		isIfaceMID, findErr := env.GetMethodID(classCls, "isInterface", "()Z")
		if findErr != nil {
			return fmt.Errorf("finding Class.isInterface: %w", findErr)
		}
		ret, callErr := env.CallBooleanMethod(&ifaces[0].Object, isIfaceMID)
		if callErr != nil {
			return fmt.Errorf("calling Class.isInterface: %w", callErr)
		}
		isInterface = ret != 0
		return nil
	}); err != nil {
		return status.Errorf(codes.Internal, "checking isInterface: %v", err)
	}

	switch {
	case isInterface:
		// Interface path: use the existing NewProxyFull mechanism.
		if err := s.VM.Do(func(env *jni.Env) error {
			var createErr error
			proxyObj, proxyCleanup, createErr = env.NewProxyFull(ifaces,
				func(env *jni.Env, method *jni.Object, methodName string, args []*jni.Object) (*jni.Object, error) {
					expectsResponse := !detector.isVoid(env, method)
					return sendCallback(env, methodName, args, expectsResponse)
				},
			)
			if createErr != nil {
				return createErr
			}
			// Store in HandleStore within the same VM.Do — local refs are
			// thread-local and invalid across VM.Do boundaries.
			proxyHandle = s.Handles.Put(env, proxyObj)
			return nil
		}); err != nil {
			return status.Errorf(codes.Internal, "creating interface proxy: %v", err)
		}

	default:
		// Abstract class path: ensure native methods are registered (the
		// interface path does this inside NewProxyFull, but we bypass it here).
		if err := s.VM.Do(func(env *jni.Env) error {
			return jni.EnsureProxyInit(env)
		}); err != nil {
			return status.Errorf(codes.Internal, "proxy init: %v", err)
		}

		// Register a basic ProxyHandler, find the generated adapter class,
		// and instantiate it with the handler ID.
		handlerID := jni.RegisterProxyHandler(
			func(env *jni.Env, methodName string, args []*jni.Object) (*jni.Object, error) {
				// Abstract class adapters don't pass a Method object, so we
				// cannot detect void return type. Always expect a response and
				// let the client return a null handle for void methods.
				return sendCallback(env, methodName, args, true)
			},
		)

		if err := s.VM.Do(func(env *jni.Env) error {
			// Get the class's full name to derive the adapter class name.
			// E.g. "android.hardware.camera2.CameraDevice$StateCallback"
			//   → adapter name: "CameraDeviceStateCallbackAdapter"
			classCls, findErr := env.FindClass("java/lang/Class")
			if findErr != nil {
				return fmt.Errorf("finding java.lang.Class: %w", findErr)
			}
			getNameMID, findErr := env.GetMethodID(classCls, "getName", "()Ljava/lang/String;")
			if findErr != nil {
				return fmt.Errorf("finding Class.getName: %w", findErr)
			}

			nameObj, callErr := env.CallObjectMethod(&ifaces[0].Object, getNameMID)
			if callErr != nil {
				return fmt.Errorf("calling Class.getName: %w", callErr)
			}
			fullName := env.GoString((*jni.String)(unsafe.Pointer(nameObj)))

			// Extract the adapter name: take the last package segment + inner
			// class parts, remove '$' and '.', append "Adapter".
			// "android.hardware.camera2.CameraDevice$StateCallback"
			//  → lastDot split: "CameraDevice$StateCallback"
			//  → replace "$" with "": "CameraDeviceStateCallback"
			//  → append "Adapter": "CameraDeviceStateCallbackAdapter"
			shortName := fullName
			if idx := strings.LastIndex(fullName, "."); idx >= 0 {
				shortName = fullName[idx+1:]
			}
			shortName = strings.ReplaceAll(shortName, "$", "")
			adapterJavaName := "center.dx.jni.generated." + shortName + "Adapter"

			// Load the adapter class via AppClassLoader (native threads
			// use BootClassLoader which cannot find APK classes).
			adapterCls, loadErr := s.loadClassViaAppClassLoader(env, adapterJavaName)
			if loadErr != nil {
				return fmt.Errorf("loading adapter class %s: %w", adapterJavaName, loadErr)
			}

			// Get the adapter's constructor: <init>(long handlerID).
			ctorMID, findErr := env.GetMethodID(adapterCls, "<init>", "(J)V")
			if findErr != nil {
				return fmt.Errorf("finding %s constructor: %w", adapterJavaName, findErr)
			}

			// Instantiate the adapter, passing the handler ID.
			adapterObj, newErr := env.NewObject(adapterCls, ctorMID, jni.LongValue(handlerID))
			if newErr != nil {
				return fmt.Errorf("instantiating %s: %w", adapterJavaName, newErr)
			}

			proxyObj = adapterObj
			// Store in HandleStore within the same VM.Do — local refs are
			// thread-local and invalid across VM.Do boundaries.
			proxyHandle = s.Handles.Put(env, adapterObj)
			return nil
		}); err != nil {
			jni.UnregisterProxyHandler(handlerID)
			return status.Errorf(codes.Internal, "creating abstract class proxy: %v", err)
		}

		proxyCleanup = func() {
			jni.UnregisterProxyHandler(handlerID)
		}
	}
	defer proxyCleanup()

	// 7. proxyHandle was set inside the VM.Do callbacks above (local JNI
	// refs are thread-local — must be stored before VM.Do returns).

	sendMu.Lock()
	err = stream.Send(&pb.ProxyServerMessage{
		Msg: &pb.ProxyServerMessage_Created{
			Created: &pb.CreateProxyResponse{
				ProxyHandle: proxyHandle,
			},
		},
	})
	sendMu.Unlock()
	if err != nil {
		return err
	}

	// 8. Receive loop: read CallbackResponse messages and dispatch.
	for {
		msg, recvErr := stream.Recv()
		if recvErr == io.EOF {
			return nil
		}
		if recvErr != nil {
			return recvErr
		}

		resp := msg.GetCallbackResponse()
		if resp == nil {
			continue
		}

		pendingMu.Lock()
		ch, ok := pending[resp.GetCallbackId()]
		pendingMu.Unlock()

		if ok {
			ch <- resp
		}
	}
}

// loadClassViaAppClassLoader loads a Java class by its dot-separated
// name (e.g. "center.dx.jni.generated.FooAdapter") using the server's
// AppClassLoader. This is necessary because native threads attach to
// BootClassLoader which cannot see APK classes.
func (s *Server) loadClassViaAppClassLoader(
	env *jni.Env,
	javaClassName string,
) (*jni.Class, error) {
	// First try JNI FindClass with the slash-separated name.
	slashName := strings.ReplaceAll(javaClassName, ".", "/")
	cls, err := env.FindClass(slashName)
	if err == nil {
		return cls, nil
	}

	if s.AppClassLoader == 0 {
		return nil, fmt.Errorf("FindClass failed and no AppClassLoader available: %w", err)
	}
	env.ExceptionClear()

	clObj := s.getObject(s.AppClassLoader)
	if clObj == nil {
		return nil, fmt.Errorf("AppClassLoader handle %d not found", s.AppClassLoader)
	}

	clCls, findErr := env.FindClass("java/lang/ClassLoader")
	if findErr != nil {
		return nil, fmt.Errorf("finding java.lang.ClassLoader: %w", findErr)
	}
	loadMID, findErr := env.GetMethodID(clCls, "loadClass", "(Ljava/lang/String;)Ljava/lang/Class;")
	if findErr != nil {
		return nil, fmt.Errorf("finding ClassLoader.loadClass: %w", findErr)
	}

	nameStr, findErr := env.NewStringUTF(javaClassName)
	if findErr != nil {
		return nil, fmt.Errorf("creating Java string for class name: %w", findErr)
	}

	classObj, callErr := env.CallObjectMethod(clObj, loadMID, jni.ObjectValue(&nameStr.Object))
	if callErr != nil {
		return nil, fmt.Errorf("ClassLoader.loadClass(%s): %w", javaClassName, callErr)
	}
	if classObj == nil || classObj.Ref() == 0 {
		return nil, fmt.Errorf("ClassLoader.loadClass(%s) returned null", javaClassName)
	}

	return (*jni.Class)(unsafe.Pointer(classObj)), nil
}
