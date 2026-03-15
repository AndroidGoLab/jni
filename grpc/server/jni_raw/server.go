// Package jni_raw implements a gRPC server that exposes the raw JNI Env
// surface over gRPC. All JNI objects are referenced by int64 handles stored
// in the shared HandleStore. MethodIDs and FieldIDs are passed as int64
// values cast from their pointer representation.
package jni_raw

import (
	"context"
	"fmt"
	"strings"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/handlestore"
	pb "github.com/xaionaro-go/jni/proto/jni_raw"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements pb.JNIServiceServer.
type Server struct {
	pb.UnimplementedJNIServiceServer
	VM      *jni.VM
	Handles *handlestore.HandleStore
	// AppContextHandle is the HandleStore handle for the Android
	// application Context. Set during server startup so clients can
	// retrieve it via the GetAppContext RPC instead of guessing.
	AppContextHandle int64
	// AppClassLoader is an optional handle to the APK's ClassLoader.
	// When set, FindClass falls back to ClassLoader.loadClass() if the
	// JNI FindClass fails (native threads use BootClassLoader which
	// can't find APK classes).
	AppClassLoader int64
}

func (s *Server) withEnv(fn func(env *jni.Env) error) error {
	return s.VM.Do(fn)
}

func (s *Server) getObject(handle int64) *jni.Object {
	return s.Handles.Get(handle)
}

func (s *Server) requireObject(handle int64) (*jni.Object, error) {
	obj := s.Handles.Get(handle)
	if obj == nil {
		return nil, status.Errorf(codes.NotFound, "handle %d not found", handle)
	}
	return obj, nil
}

func (s *Server) requireClass(handle int64) (*jni.Class, error) {
	obj, err := s.requireObject(handle)
	if err != nil {
		return nil, err
	}
	return (*jni.Class)(unsafe.Pointer(obj)), nil
}

func (s *Server) putObject(env *jni.Env, obj *jni.Object) int64 {
	return s.Handles.Put(env, obj)
}

func methodID(id int64) jni.MethodID {
	return jni.MethodID(unsafe.Pointer(uintptr(id)))
}

func fieldID(id int64) jni.FieldID {
	return jni.FieldID(unsafe.Pointer(uintptr(id)))
}

func methodIDToInt64(id jni.MethodID) int64 {
	return int64(uintptr(unsafe.Pointer(id)))
}

func fieldIDToInt64(id jni.FieldID) int64 {
	return int64(uintptr(unsafe.Pointer(id)))
}

func jvalueFromProto(v *pb.JValue) jni.Value {
	switch val := v.GetValue().(type) {
	case *pb.JValue_Z:
		if val.Z {
			return jni.BooleanValue(1)
		}
		return jni.BooleanValue(0)
	case *pb.JValue_B:
		return jni.ByteValue(int8(val.B))
	case *pb.JValue_C:
		return jni.CharValue(uint16(val.C))
	case *pb.JValue_S:
		return jni.ShortValue(int16(val.S))
	case *pb.JValue_I:
		return jni.IntValue(val.I)
	case *pb.JValue_J:
		return jni.LongValue(val.J)
	case *pb.JValue_F:
		return jni.FloatValue(val.F)
	case *pb.JValue_D:
		return jni.DoubleValue(val.D)
	case *pb.JValue_L:
		return jni.ObjectValue(nil) // caller resolves handle
	default:
		return jni.IntValue(0)
	}
}

func jvaluesFromProto(args []*pb.JValue, handles *handlestore.HandleStore) []jni.Value {
	vals := make([]jni.Value, len(args))
	for i, a := range args {
		if l, ok := a.GetValue().(*pb.JValue_L); ok {
			vals[i] = jni.ObjectValue(handles.Get(l.L))
		} else {
			vals[i] = jvalueFromProto(a)
		}
	}
	return vals
}

// ---- Version ----

func (s *Server) GetVersion(_ context.Context, _ *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	var version int32
	if err := s.withEnv(func(env *jni.Env) error {
		version = env.GetVersion()
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetVersionResponse{Version: version}, nil
}

// ---- App Context ----

func (s *Server) GetAppContext(_ context.Context, _ *pb.GetAppContextRequest) (*pb.GetAppContextResponse, error) {
	if s.AppContextHandle == 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "app context not available")
	}
	return &pb.GetAppContextResponse{ContextHandle: s.AppContextHandle}, nil
}

// ---- Class ----

func (s *Server) FindClass(_ context.Context, req *pb.FindClassRequest) (*pb.FindClassResponse, error) {
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls, err := env.FindClass(req.GetName())
		if err == nil {
			handle = s.putObject(env, &cls.Object)
			return nil
		}

		// Fallback: native threads use BootClassLoader which can't find APK
		// classes. Use the stored AppClassLoader (if available) to retry via
		// ClassLoader.loadClass(). The class name uses dots, not slashes.
		if s.AppClassLoader == 0 {
			return err
		}
		env.ExceptionClear()

		clObj := s.getObject(s.AppClassLoader)
		if clObj == nil {
			return err
		}

		clCls, findErr := env.FindClass("java/lang/ClassLoader")
		if findErr != nil {
			return err
		}
		loadMID, findErr := env.GetMethodID(clCls, "loadClass", "(Ljava/lang/String;)Ljava/lang/Class;")
		if findErr != nil {
			return err
		}

		// Convert JNI name (e.g. "center/dx/jni/jniservice/CameraCapture")
		// to Java name (e.g. "center.dx.jni.jniservice.CameraCapture").
		javaName := strings.ReplaceAll(req.GetName(), "/", ".")
		nameStr, findErr := env.NewStringUTF(javaName)
		if findErr != nil {
			return err
		}

		classObj, findErr := env.CallObjectMethod(clObj, loadMID, jni.ObjectValue(&nameStr.Object))
		if findErr != nil {
			return findErr
		}
		if classObj == nil || classObj.Ref() == 0 {
			return err
		}

		handle = s.putObject(env, classObj)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "FindClass: %v", err)
	}
	return &pb.FindClassResponse{ClassHandle: handle}, nil
}

func (s *Server) GetSuperclass(_ context.Context, req *pb.GetSuperclassRequest) (*pb.GetSuperclassResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		super := env.GetSuperclass(cls)
		if super != nil {
			handle = s.putObject(env, &super.Object)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetSuperclassResponse{ClassHandle: handle}, nil
}

func (s *Server) IsAssignableFrom(_ context.Context, req *pb.IsAssignableFromRequest) (*pb.IsAssignableFromResponse, error) {
	c1, err := s.requireClass(req.GetClass1())
	if err != nil {
		return nil, err
	}
	c2, err := s.requireClass(req.GetClass2())
	if err != nil {
		return nil, err
	}
	var result bool
	if err := s.withEnv(func(env *jni.Env) error {
		result = env.IsAssignableFrom(c1, c2)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.IsAssignableFromResponse{Result: result}, nil
}

// ---- Object ----

func (s *Server) AllocObject(_ context.Context, req *pb.AllocObjectRequest) (*pb.AllocObjectResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		obj, err := env.AllocObject(cls)
		if err != nil {
			return err
		}
		handle = s.putObject(env, obj)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.AllocObjectResponse{ObjectHandle: handle}, nil
}

func (s *Server) NewObject(_ context.Context, req *pb.NewObjectRequest) (*pb.NewObjectResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		args := jvaluesFromProto(req.GetArgs(), s.Handles)
		obj, err := env.NewObject(cls, methodID(req.GetMethodId()), args...)
		if err != nil {
			return err
		}
		handle = s.putObject(env, obj)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.NewObjectResponse{ObjectHandle: handle}, nil
}

func (s *Server) GetObjectClass(_ context.Context, req *pb.GetObjectClassRequest) (*pb.GetObjectClassResponse, error) {
	obj, err := s.requireObject(req.GetObjectHandle())
	if err != nil {
		return nil, err
	}
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls := env.GetObjectClass(obj)
		handle = s.putObject(env, &cls.Object)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetObjectClassResponse{ClassHandle: handle}, nil
}

func (s *Server) IsInstanceOf(_ context.Context, req *pb.IsInstanceOfRequest) (*pb.IsInstanceOfResponse, error) {
	obj, err := s.requireObject(req.GetObjectHandle())
	if err != nil {
		return nil, err
	}
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var result bool
	if err := s.withEnv(func(env *jni.Env) error {
		result = env.IsInstanceOf(obj, cls)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.IsInstanceOfResponse{Result: result}, nil
}

func (s *Server) IsSameObject(_ context.Context, req *pb.IsSameObjectRequest) (*pb.IsSameObjectResponse, error) {
	o1, err := s.requireObject(req.GetObject1())
	if err != nil {
		return nil, err
	}
	o2, err := s.requireObject(req.GetObject2())
	if err != nil {
		return nil, err
	}
	var result bool
	if err := s.withEnv(func(env *jni.Env) error {
		result = env.IsSameObject(o1, o2)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.IsSameObjectResponse{Result: result}, nil
}

func (s *Server) GetObjectRefType(_ context.Context, req *pb.GetObjectRefTypeRequest) (*pb.GetObjectRefTypeResponse, error) {
	obj, err := s.requireObject(req.GetObjectHandle())
	if err != nil {
		return nil, err
	}
	var refType int32
	if err := s.withEnv(func(env *jni.Env) error {
		refType = int32(env.GetObjectRefType(obj))
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetObjectRefTypeResponse{RefType: refType}, nil
}

// ---- Method/Field ID lookup ----

func (s *Server) GetMethodID(_ context.Context, req *pb.GetMethodIDRequest) (*pb.GetMethodIDResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		mid, err := env.GetMethodID(cls, req.GetName(), req.GetSig())
		if err != nil {
			return err
		}
		id = methodIDToInt64(mid)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetMethodIDResponse{MethodId: id}, nil
}

func (s *Server) GetStaticMethodID(_ context.Context, req *pb.GetStaticMethodIDRequest) (*pb.GetStaticMethodIDResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		mid, err := env.GetStaticMethodID(cls, req.GetName(), req.GetSig())
		if err != nil {
			return err
		}
		id = methodIDToInt64(mid)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetStaticMethodIDResponse{MethodId: id}, nil
}

func (s *Server) GetFieldID(_ context.Context, req *pb.GetFieldIDRequest) (*pb.GetFieldIDResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		fid, err := env.GetFieldID(cls, req.GetName(), req.GetSig())
		if err != nil {
			return err
		}
		id = fieldIDToInt64(fid)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetFieldIDResponse{FieldId: id}, nil
}

func (s *Server) GetStaticFieldID(_ context.Context, req *pb.GetStaticFieldIDRequest) (*pb.GetStaticFieldIDResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		fid, err := env.GetStaticFieldID(cls, req.GetName(), req.GetSig())
		if err != nil {
			return err
		}
		id = fieldIDToInt64(fid)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetStaticFieldIDResponse{FieldId: id}, nil
}

// ---- Method calls ----

func (s *Server) CallMethod(_ context.Context, req *pb.CallMethodRequest) (*pb.CallMethodResponse, error) {
	obj, err := s.requireObject(req.GetObjectHandle())
	if err != nil {
		return nil, err
	}
	var result *pb.JValue
	if err := s.withEnv(func(env *jni.Env) error {
		mid := methodID(req.GetMethodId())
		args := jvaluesFromProto(req.GetArgs(), s.Handles)
		var err error
		result, err = s.callMethod(env, obj, mid, req.GetReturnType(), args)
		return err
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.CallMethodResponse{Result: result}, nil
}

func (s *Server) CallStaticMethod(_ context.Context, req *pb.CallStaticMethodRequest) (*pb.CallStaticMethodResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var result *pb.JValue
	if err := s.withEnv(func(env *jni.Env) error {
		mid := methodID(req.GetMethodId())
		args := jvaluesFromProto(req.GetArgs(), s.Handles)
		var err error
		result, err = s.callStaticMethod(env, cls, mid, req.GetReturnType(), args)
		return err
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.CallStaticMethodResponse{Result: result}, nil
}

func (s *Server) CallNonvirtualMethod(_ context.Context, req *pb.CallNonvirtualMethodRequest) (*pb.CallNonvirtualMethodResponse, error) {
	obj, err := s.requireObject(req.GetObjectHandle())
	if err != nil {
		return nil, err
	}
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var result *pb.JValue
	if err := s.withEnv(func(env *jni.Env) error {
		mid := methodID(req.GetMethodId())
		args := jvaluesFromProto(req.GetArgs(), s.Handles)
		var err error
		result, err = s.callNonvirtualMethod(env, obj, cls, mid, req.GetReturnType(), args)
		return err
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.CallNonvirtualMethodResponse{Result: result}, nil
}

func (s *Server) callMethod(
	env *jni.Env,
	obj *jni.Object,
	mid jni.MethodID,
	retType pb.JType,
	args []jni.Value,
) (*pb.JValue, error) {
	switch retType {
	case pb.JType_VOID:
		return nil, env.CallVoidMethod(obj, mid, args...)
	case pb.JType_BOOLEAN:
		v, err := env.CallBooleanMethod(obj, mid, args...)
		return &pb.JValue{Value: &pb.JValue_Z{Z: v != 0}}, err
	case pb.JType_BYTE:
		v, err := env.CallByteMethod(obj, mid, args...)
		return &pb.JValue{Value: &pb.JValue_B{B: int32(v)}}, err
	case pb.JType_CHAR:
		v, err := env.CallCharMethod(obj, mid, args...)
		return &pb.JValue{Value: &pb.JValue_C{C: uint32(v)}}, err
	case pb.JType_SHORT:
		v, err := env.CallShortMethod(obj, mid, args...)
		return &pb.JValue{Value: &pb.JValue_S{S: int32(v)}}, err
	case pb.JType_INT:
		v, err := env.CallIntMethod(obj, mid, args...)
		return &pb.JValue{Value: &pb.JValue_I{I: v}}, err
	case pb.JType_LONG:
		v, err := env.CallLongMethod(obj, mid, args...)
		return &pb.JValue{Value: &pb.JValue_J{J: v}}, err
	case pb.JType_FLOAT:
		v, err := env.CallFloatMethod(obj, mid, args...)
		return &pb.JValue{Value: &pb.JValue_F{F: v}}, err
	case pb.JType_DOUBLE:
		v, err := env.CallDoubleMethod(obj, mid, args...)
		return &pb.JValue{Value: &pb.JValue_D{D: v}}, err
	case pb.JType_OBJECT:
		v, err := env.CallObjectMethod(obj, mid, args...)
		if err != nil {
			return nil, err
		}
		var h int64
		if v != nil && v.Ref() != 0 {
			h = s.putObject(env, v)
		}
		return &pb.JValue{Value: &pb.JValue_L{L: h}}, nil
	default:
		return nil, fmt.Errorf("unknown return type: %v", retType)
	}
}

func (s *Server) callStaticMethod(
	env *jni.Env,
	cls *jni.Class,
	mid jni.MethodID,
	retType pb.JType,
	args []jni.Value,
) (*pb.JValue, error) {
	switch retType {
	case pb.JType_VOID:
		return nil, env.CallStaticVoidMethod(cls, mid, args...)
	case pb.JType_BOOLEAN:
		v, err := env.CallStaticBooleanMethod(cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_Z{Z: v != 0}}, err
	case pb.JType_BYTE:
		v, err := env.CallStaticByteMethod(cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_B{B: int32(v)}}, err
	case pb.JType_CHAR:
		v, err := env.CallStaticCharMethod(cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_C{C: uint32(v)}}, err
	case pb.JType_SHORT:
		v, err := env.CallStaticShortMethod(cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_S{S: int32(v)}}, err
	case pb.JType_INT:
		v, err := env.CallStaticIntMethod(cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_I{I: v}}, err
	case pb.JType_LONG:
		v, err := env.CallStaticLongMethod(cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_J{J: v}}, err
	case pb.JType_FLOAT:
		v, err := env.CallStaticFloatMethod(cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_F{F: v}}, err
	case pb.JType_DOUBLE:
		v, err := env.CallStaticDoubleMethod(cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_D{D: v}}, err
	case pb.JType_OBJECT:
		v, err := env.CallStaticObjectMethod(cls, mid, args...)
		if err != nil {
			return nil, err
		}
		var h int64
		if v != nil && v.Ref() != 0 {
			h = s.putObject(env, v)
		}
		return &pb.JValue{Value: &pb.JValue_L{L: h}}, nil
	default:
		return nil, fmt.Errorf("unknown return type: %v", retType)
	}
}

func (s *Server) callNonvirtualMethod(
	env *jni.Env,
	obj *jni.Object,
	cls *jni.Class,
	mid jni.MethodID,
	retType pb.JType,
	args []jni.Value,
) (*pb.JValue, error) {
	switch retType {
	case pb.JType_VOID:
		return nil, env.CallNonvirtualVoidMethod(obj, cls, mid, args...)
	case pb.JType_BOOLEAN:
		v, err := env.CallNonvirtualBooleanMethod(obj, cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_Z{Z: v != 0}}, err
	case pb.JType_BYTE:
		v, err := env.CallNonvirtualByteMethod(obj, cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_B{B: int32(v)}}, err
	case pb.JType_CHAR:
		v, err := env.CallNonvirtualCharMethod(obj, cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_C{C: uint32(v)}}, err
	case pb.JType_SHORT:
		v, err := env.CallNonvirtualShortMethod(obj, cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_S{S: int32(v)}}, err
	case pb.JType_INT:
		v, err := env.CallNonvirtualIntMethod(obj, cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_I{I: v}}, err
	case pb.JType_LONG:
		v, err := env.CallNonvirtualLongMethod(obj, cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_J{J: v}}, err
	case pb.JType_FLOAT:
		v, err := env.CallNonvirtualFloatMethod(obj, cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_F{F: v}}, err
	case pb.JType_DOUBLE:
		v, err := env.CallNonvirtualDoubleMethod(obj, cls, mid, args...)
		return &pb.JValue{Value: &pb.JValue_D{D: v}}, err
	case pb.JType_OBJECT:
		v, err := env.CallNonvirtualObjectMethod(obj, cls, mid, args...)
		if err != nil {
			return nil, err
		}
		var h int64
		if v != nil && v.Ref() != 0 {
			h = s.putObject(env, v)
		}
		return &pb.JValue{Value: &pb.JValue_L{L: h}}, nil
	default:
		return nil, fmt.Errorf("unknown return type: %v", retType)
	}
}

// ---- String ----

func (s *Server) NewStringUTF(_ context.Context, req *pb.NewStringUTFRequest) (*pb.NewStringUTFResponse, error) {
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		str, err := env.NewStringUTF(req.GetValue())
		if err != nil {
			return err
		}
		handle = s.putObject(env, &str.Object)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.NewStringUTFResponse{StringHandle: handle}, nil
}

func (s *Server) GetStringUTFChars(_ context.Context, req *pb.GetStringUTFCharsRequest) (*pb.GetStringUTFCharsResponse, error) {
	obj, err := s.requireObject(req.GetStringHandle())
	if err != nil {
		return nil, err
	}
	var value string
	if err := s.withEnv(func(env *jni.Env) error {
		str := (*jni.String)(unsafe.Pointer(obj))
		value = env.GoString(str)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetStringUTFCharsResponse{Value: value}, nil
}

func (s *Server) GetStringLength(_ context.Context, req *pb.GetStringLengthRequest) (*pb.GetStringLengthResponse, error) {
	obj, err := s.requireObject(req.GetStringHandle())
	if err != nil {
		return nil, err
	}
	var length int32
	if err := s.withEnv(func(env *jni.Env) error {
		str := (*jni.String)(unsafe.Pointer(obj))
		length = env.GetStringLength(str)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetStringLengthResponse{Length: length}, nil
}

// ---- Exception handling ----

func (s *Server) ExceptionCheck(_ context.Context, _ *pb.ExceptionCheckRequest) (*pb.ExceptionCheckResponse, error) {
	var has bool
	if err := s.withEnv(func(env *jni.Env) error {
		has = env.ExceptionCheck()
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.ExceptionCheckResponse{HasException: has}, nil
}

func (s *Server) ExceptionClear(_ context.Context, _ *pb.ExceptionClearRequest) (*pb.ExceptionClearResponse, error) {
	if err := s.withEnv(func(env *jni.Env) error {
		env.ExceptionClear()
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.ExceptionClearResponse{}, nil
}

func (s *Server) ExceptionDescribe(_ context.Context, _ *pb.ExceptionDescribeRequest) (*pb.ExceptionDescribeResponse, error) {
	if err := s.withEnv(func(env *jni.Env) error {
		env.ExceptionDescribe()
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.ExceptionDescribeResponse{}, nil
}

func (s *Server) ExceptionOccurred(_ context.Context, _ *pb.ExceptionOccurredRequest) (*pb.ExceptionOccurredResponse, error) {
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		t := env.ExceptionOccurred()
		if t != nil {
			handle = s.putObject(env, &t.Object)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.ExceptionOccurredResponse{ThrowableHandle: handle}, nil
}

func (s *Server) Throw(_ context.Context, req *pb.ThrowRequest) (*pb.ThrowResponse, error) {
	obj, err := s.requireObject(req.GetThrowableHandle())
	if err != nil {
		return nil, err
	}
	if err := s.withEnv(func(env *jni.Env) error {
		return env.Throw((*jni.Throwable)(unsafe.Pointer(obj)))
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.ThrowResponse{}, nil
}

func (s *Server) ThrowNew(_ context.Context, req *pb.ThrowNewRequest) (*pb.ThrowNewResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	if err := s.withEnv(func(env *jni.Env) error {
		return env.ThrowNew(cls, req.GetMessage())
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.ThrowNewResponse{}, nil
}

// ---- Monitor ----

func (s *Server) MonitorEnter(_ context.Context, req *pb.MonitorEnterRequest) (*pb.MonitorEnterResponse, error) {
	obj, err := s.requireObject(req.GetObjectHandle())
	if err != nil {
		return nil, err
	}
	if err := s.withEnv(func(env *jni.Env) error {
		return env.MonitorEnter(obj)
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.MonitorEnterResponse{}, nil
}

func (s *Server) MonitorExit(_ context.Context, req *pb.MonitorExitRequest) (*pb.MonitorExitResponse, error) {
	obj, err := s.requireObject(req.GetObjectHandle())
	if err != nil {
		return nil, err
	}
	if err := s.withEnv(func(env *jni.Env) error {
		return env.MonitorExit(obj)
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.MonitorExitResponse{}, nil
}

// ---- Local frame ----

func (s *Server) PushLocalFrame(_ context.Context, req *pb.PushLocalFrameRequest) (*pb.PushLocalFrameResponse, error) {
	if err := s.withEnv(func(env *jni.Env) error {
		return env.PushLocalFrame(req.GetCapacity())
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.PushLocalFrameResponse{}, nil
}

func (s *Server) PopLocalFrame(_ context.Context, req *pb.PopLocalFrameRequest) (*pb.PopLocalFrameResponse, error) {
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		result := env.PopLocalFrame(s.getObject(req.GetResultHandle()))
		if result != nil {
			handle = s.putObject(env, result)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.PopLocalFrameResponse{ResultHandle: handle}, nil
}

func (s *Server) EnsureLocalCapacity(_ context.Context, req *pb.EnsureLocalCapacityRequest) (*pb.EnsureLocalCapacityResponse, error) {
	if err := s.withEnv(func(env *jni.Env) error {
		return env.EnsureLocalCapacity(req.GetCapacity())
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.EnsureLocalCapacityResponse{}, nil
}

// ---- Reflection ----

func (s *Server) FromReflectedMethod(_ context.Context, req *pb.FromReflectedMethodRequest) (*pb.FromReflectedMethodResponse, error) {
	obj, err := s.requireObject(req.GetMethodObject())
	if err != nil {
		return nil, err
	}
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		mid := env.FromReflectedMethod(obj)
		id = methodIDToInt64(mid)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.FromReflectedMethodResponse{MethodId: id}, nil
}

func (s *Server) FromReflectedField(_ context.Context, req *pb.FromReflectedFieldRequest) (*pb.FromReflectedFieldResponse, error) {
	obj, err := s.requireObject(req.GetFieldObject())
	if err != nil {
		return nil, err
	}
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		fid := env.FromReflectedField(obj)
		id = fieldIDToInt64(fid)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.FromReflectedFieldResponse{FieldId: id}, nil
}

func (s *Server) ToReflectedMethod(_ context.Context, req *pb.ToReflectedMethodRequest) (*pb.ToReflectedMethodResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		obj := env.ToReflectedMethod(cls, methodID(req.GetMethodId()), req.GetIsStatic())
		if obj != nil {
			handle = s.putObject(env, obj)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.ToReflectedMethodResponse{MethodObject: handle}, nil
}

func (s *Server) ToReflectedField(_ context.Context, req *pb.ToReflectedFieldRequest) (*pb.ToReflectedFieldResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		obj := env.ToReflectedField(cls, fieldID(req.GetFieldId()), req.GetIsStatic())
		if obj != nil {
			handle = s.putObject(env, obj)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.ToReflectedFieldResponse{FieldObject: handle}, nil
}

// ---- Array operations ----

func (s *Server) GetArrayLength(_ context.Context, req *pb.GetArrayLengthRequest) (*pb.GetArrayLengthResponse, error) {
	obj, err := s.requireObject(req.GetArrayHandle())
	if err != nil {
		return nil, err
	}
	var length int32
	if err := s.withEnv(func(env *jni.Env) error {
		length = env.GetArrayLength((*jni.Array)(unsafe.Pointer(obj)))
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetArrayLengthResponse{Length: length}, nil
}

func (s *Server) NewObjectArray(_ context.Context, req *pb.NewObjectArrayRequest) (*pb.NewObjectArrayResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		var initElem *jni.Object
		if req.GetInitElement() != 0 {
			initElem = s.getObject(req.GetInitElement())
		}
		arr, err := env.NewObjectArray(req.GetLength(), cls, initElem)
		if err != nil {
			return err
		}
		handle = s.putObject(env, &arr.Object)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.NewObjectArrayResponse{ArrayHandle: handle}, nil
}

func (s *Server) GetObjectArrayElement(_ context.Context, req *pb.GetObjectArrayElementRequest) (*pb.GetObjectArrayElementResponse, error) {
	obj, err := s.requireObject(req.GetArrayHandle())
	if err != nil {
		return nil, err
	}
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		arr := (*jni.ObjectArray)(unsafe.Pointer(obj))
		elem, err := env.GetObjectArrayElement(arr, req.GetIndex())
		if err != nil {
			return err
		}
		if elem != nil && elem.Ref() != 0 {
			handle = s.putObject(env, elem)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetObjectArrayElementResponse{ElementHandle: handle}, nil
}

func (s *Server) SetObjectArrayElement(_ context.Context, req *pb.SetObjectArrayElementRequest) (*pb.SetObjectArrayElementResponse, error) {
	obj, err := s.requireObject(req.GetArrayHandle())
	if err != nil {
		return nil, err
	}
	if err := s.withEnv(func(env *jni.Env) error {
		arr := (*jni.ObjectArray)(unsafe.Pointer(obj))
		var elem *jni.Object
		if req.GetElementHandle() != 0 {
			elem = s.getObject(req.GetElementHandle())
		}
		return env.SetObjectArrayElement(arr, req.GetIndex(), elem)
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.SetObjectArrayElementResponse{}, nil
}

func (s *Server) NewPrimitiveArray(_ context.Context, req *pb.NewPrimitiveArrayRequest) (*pb.NewPrimitiveArrayResponse, error) {
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		switch req.GetElementType() {
		case pb.JType_BYTE:
			arr := env.NewByteArray(req.GetLength())
			handle = s.putObject(env, &arr.Object)
		default:
			return fmt.Errorf("unsupported element type: %v (only byte arrays currently supported)", req.GetElementType())
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.NewPrimitiveArrayResponse{ArrayHandle: handle}, nil
}

// ---- Bulk byte array transfer ----

func (s *Server) GetByteArrayData(_ context.Context, req *pb.GetByteArrayDataRequest) (*pb.GetByteArrayDataResponse, error) {
	obj, err := s.requireObject(req.GetArrayHandle())
	if err != nil {
		return nil, err
	}
	var data []byte
	if err := s.withEnv(func(env *jni.Env) error {
		arr := (*jni.Array)(unsafe.Pointer(obj))
		byteArr := (*jni.ByteArray)(unsafe.Pointer(obj))
		length := env.GetArrayLength(arr)
		data = make([]byte, length)
		env.GetByteArrayRegion(byteArr, 0, length, unsafe.Pointer(&data[0]))
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetByteArrayDataResponse{Data: data}, nil
}

// ---- Field access ----

func (s *Server) GetField(_ context.Context, req *pb.GetFieldValueRequest) (*pb.GetFieldValueResponse, error) {
	obj, err := s.requireObject(req.GetObjectHandle())
	if err != nil {
		return nil, err
	}
	var result *pb.JValue
	if err := s.withEnv(func(env *jni.Env) error {
		fid := fieldID(req.GetFieldId())
		var err error
		result, err = s.getFieldValue(env, obj, fid, req.GetFieldType())
		return err
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetFieldValueResponse{Result: result}, nil
}

func (s *Server) SetField(_ context.Context, req *pb.SetFieldValueRequest) (*pb.SetFieldValueResponse, error) {
	obj, err := s.requireObject(req.GetObjectHandle())
	if err != nil {
		return nil, err
	}
	if err := s.withEnv(func(env *jni.Env) error {
		fid := fieldID(req.GetFieldId())
		s.setFieldValue(env, obj, fid, req.GetValue())
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.SetFieldValueResponse{}, nil
}

func (s *Server) GetStaticField(_ context.Context, req *pb.GetStaticFieldValueRequest) (*pb.GetStaticFieldValueResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	var result *pb.JValue
	if err := s.withEnv(func(env *jni.Env) error {
		fid := fieldID(req.GetFieldId())
		var err error
		result, err = s.getStaticFieldValue(env, cls, fid, req.GetFieldType())
		return err
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetStaticFieldValueResponse{Result: result}, nil
}

func (s *Server) SetStaticField(_ context.Context, req *pb.SetStaticFieldValueRequest) (*pb.SetStaticFieldValueResponse, error) {
	cls, err := s.requireClass(req.GetClassHandle())
	if err != nil {
		return nil, err
	}
	if err := s.withEnv(func(env *jni.Env) error {
		fid := fieldID(req.GetFieldId())
		s.setStaticFieldValue(env, cls, fid, req.GetValue())
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.SetStaticFieldValueResponse{}, nil
}

func (s *Server) getFieldValue(
	env *jni.Env,
	obj *jni.Object,
	fid jni.FieldID,
	fieldType pb.JType,
) (*pb.JValue, error) {
	switch fieldType {
	case pb.JType_BOOLEAN:
		return &pb.JValue{Value: &pb.JValue_Z{Z: env.GetBooleanField(obj, fid) != 0}}, nil
	case pb.JType_BYTE:
		return &pb.JValue{Value: &pb.JValue_B{B: int32(env.GetByteField(obj, fid))}}, nil
	case pb.JType_CHAR:
		return &pb.JValue{Value: &pb.JValue_C{C: uint32(env.GetCharField(obj, fid))}}, nil
	case pb.JType_SHORT:
		return &pb.JValue{Value: &pb.JValue_S{S: int32(env.GetShortField(obj, fid))}}, nil
	case pb.JType_INT:
		return &pb.JValue{Value: &pb.JValue_I{I: env.GetIntField(obj, fid)}}, nil
	case pb.JType_LONG:
		return &pb.JValue{Value: &pb.JValue_J{J: env.GetLongField(obj, fid)}}, nil
	case pb.JType_FLOAT:
		return &pb.JValue{Value: &pb.JValue_F{F: env.GetFloatField(obj, fid)}}, nil
	case pb.JType_DOUBLE:
		return &pb.JValue{Value: &pb.JValue_D{D: env.GetDoubleField(obj, fid)}}, nil
	case pb.JType_OBJECT:
		v := env.GetObjectField(obj, fid)
		var h int64
		if v != nil && v.Ref() != 0 {
			h = s.putObject(env, v)
		}
		return &pb.JValue{Value: &pb.JValue_L{L: h}}, nil
	default:
		return nil, fmt.Errorf("unknown field type: %v", fieldType)
	}
}

func (s *Server) setFieldValue(
	env *jni.Env,
	obj *jni.Object,
	fid jni.FieldID,
	val *pb.JValue,
) {
	switch v := val.GetValue().(type) {
	case *pb.JValue_Z:
		var b uint8
		if v.Z {
			b = 1
		}
		env.SetBooleanField(obj, fid, b)
	case *pb.JValue_B:
		env.SetByteField(obj, fid, int8(v.B))
	case *pb.JValue_C:
		env.SetCharField(obj, fid, uint16(v.C))
	case *pb.JValue_S:
		env.SetShortField(obj, fid, int16(v.S))
	case *pb.JValue_I:
		env.SetIntField(obj, fid, v.I)
	case *pb.JValue_J:
		env.SetLongField(obj, fid, v.J)
	case *pb.JValue_F:
		env.SetFloatField(obj, fid, v.F)
	case *pb.JValue_D:
		env.SetDoubleField(obj, fid, v.D)
	case *pb.JValue_L:
		env.SetObjectField(obj, fid, s.getObject(v.L))
	}
}

func (s *Server) getStaticFieldValue(
	env *jni.Env,
	cls *jni.Class,
	fid jni.FieldID,
	fieldType pb.JType,
) (*pb.JValue, error) {
	switch fieldType {
	case pb.JType_BOOLEAN:
		return &pb.JValue{Value: &pb.JValue_Z{Z: env.GetStaticBooleanField(cls, fid) != 0}}, nil
	case pb.JType_BYTE:
		return &pb.JValue{Value: &pb.JValue_B{B: int32(env.GetStaticByteField(cls, fid))}}, nil
	case pb.JType_CHAR:
		return &pb.JValue{Value: &pb.JValue_C{C: uint32(env.GetStaticCharField(cls, fid))}}, nil
	case pb.JType_SHORT:
		return &pb.JValue{Value: &pb.JValue_S{S: int32(env.GetStaticShortField(cls, fid))}}, nil
	case pb.JType_INT:
		return &pb.JValue{Value: &pb.JValue_I{I: env.GetStaticIntField(cls, fid)}}, nil
	case pb.JType_LONG:
		return &pb.JValue{Value: &pb.JValue_J{J: env.GetStaticLongField(cls, fid)}}, nil
	case pb.JType_FLOAT:
		return &pb.JValue{Value: &pb.JValue_F{F: env.GetStaticFloatField(cls, fid)}}, nil
	case pb.JType_DOUBLE:
		return &pb.JValue{Value: &pb.JValue_D{D: env.GetStaticDoubleField(cls, fid)}}, nil
	case pb.JType_OBJECT:
		v := env.GetStaticObjectField(cls, fid)
		var h int64
		if v != nil && v.Ref() != 0 {
			h = s.putObject(env, v)
		}
		return &pb.JValue{Value: &pb.JValue_L{L: h}}, nil
	default:
		return nil, fmt.Errorf("unknown field type: %v", fieldType)
	}
}

func (s *Server) setStaticFieldValue(
	env *jni.Env,
	cls *jni.Class,
	fid jni.FieldID,
	val *pb.JValue,
) {
	switch v := val.GetValue().(type) {
	case *pb.JValue_Z:
		var b uint8
		if v.Z {
			b = 1
		}
		env.SetStaticBooleanField(cls, fid, b)
	case *pb.JValue_B:
		env.SetStaticByteField(cls, fid, int8(v.B))
	case *pb.JValue_C:
		env.SetStaticCharField(cls, fid, uint16(v.C))
	case *pb.JValue_S:
		env.SetStaticShortField(cls, fid, int16(v.S))
	case *pb.JValue_I:
		env.SetStaticIntField(cls, fid, v.I)
	case *pb.JValue_J:
		env.SetStaticLongField(cls, fid, v.J)
	case *pb.JValue_F:
		env.SetStaticFloatField(cls, fid, v.F)
	case *pb.JValue_D:
		env.SetStaticDoubleField(cls, fid, v.D)
	case *pb.JValue_L:
		env.SetStaticObjectField(cls, fid, s.getObject(v.L))
	}
}
