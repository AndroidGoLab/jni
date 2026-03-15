package jni_raw

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

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

	// 5. Cache for void return type detection.
	var detector voidDetector

	// 6. Create the proxy with a handler that forwards callbacks to the stream.
	var (
		proxyObj     *jni.Object
		proxyCleanup func()
	)
	if err := s.VM.Do(func(env *jni.Env) error {
		var createErr error
		proxyObj, proxyCleanup, createErr = env.NewProxyFull(ifaces,
			func(env *jni.Env, method *jni.Object, methodName string, args []*jni.Object) (*jni.Object, error) {
				callbackID := nextID.Add(1)

				// Store args in HandleStore so client can reference them.
				argHandles := make([]int64, len(args))
				for i, arg := range args {
					if arg != nil && arg.Ref() != 0 {
						argHandles[i] = s.Handles.Put(env, arg)
					}
				}

				expectsResponse := !detector.isVoid(env, method)

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
			},
		)
		return createErr
	}); err != nil {
		return status.Errorf(codes.Internal, "creating proxy: %v", err)
	}
	defer proxyCleanup()

	// 7. Store proxy in HandleStore and send response.
	var proxyHandle int64
	s.VM.Do(func(env *jni.Env) error {
		proxyHandle = s.Handles.Put(env, proxyObj)
		return nil
	})

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
