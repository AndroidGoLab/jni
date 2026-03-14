// Package jni_raw implements a gRPC server that exposes the raw JNI Env
// surface over gRPC. All JNI objects are referenced by int64 handles stored
// in the shared HandleStore. MethodIDs and FieldIDs are passed as int64
// values cast from their pointer representation.
package jni_raw

import (
	"context"
	"fmt"
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
}

func (s *Server) withEnv(fn func(env *jni.Env) error) error {
	return s.VM.Do(fn)
}

func (s *Server) getObject(handle int64) *jni.Object {
	return s.Handles.Get(handle)
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

// ---- Class ----

func (s *Server) FindClass(_ context.Context, req *pb.FindClassRequest) (*pb.FindClassResponse, error) {
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls, err := env.FindClass(req.GetName())
		if err != nil {
			return err
		}
		handle = s.putObject(env, &cls.Object)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "FindClass: %v", err)
	}
	return &pb.FindClassResponse{ClassHandle: handle}, nil
}

func (s *Server) GetSuperclass(_ context.Context, req *pb.GetSuperclassRequest) (*pb.GetSuperclassResponse, error) {
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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
	var result bool
	if err := s.withEnv(func(env *jni.Env) error {
		c1 := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClass1())))
		c2 := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClass2())))
		result = env.IsAssignableFrom(c1, c2)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.IsAssignableFromResponse{Result: result}, nil
}

// ---- Object ----

func (s *Server) AllocObject(_ context.Context, req *pb.AllocObjectRequest) (*pb.AllocObjectResponse, error) {
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		obj := s.getObject(req.GetObjectHandle())
		cls := env.GetObjectClass(obj)
		handle = s.putObject(env, &cls.Object)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetObjectClassResponse{ClassHandle: handle}, nil
}

func (s *Server) IsInstanceOf(_ context.Context, req *pb.IsInstanceOfRequest) (*pb.IsInstanceOfResponse, error) {
	var result bool
	if err := s.withEnv(func(env *jni.Env) error {
		obj := s.getObject(req.GetObjectHandle())
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
		result = env.IsInstanceOf(obj, cls)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.IsInstanceOfResponse{Result: result}, nil
}

func (s *Server) IsSameObject(_ context.Context, req *pb.IsSameObjectRequest) (*pb.IsSameObjectResponse, error) {
	var result bool
	if err := s.withEnv(func(env *jni.Env) error {
		result = env.IsSameObject(s.getObject(req.GetObject1()), s.getObject(req.GetObject2()))
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.IsSameObjectResponse{Result: result}, nil
}

func (s *Server) GetObjectRefType(_ context.Context, req *pb.GetObjectRefTypeRequest) (*pb.GetObjectRefTypeResponse, error) {
	var refType int32
	if err := s.withEnv(func(env *jni.Env) error {
		refType = int32(env.GetObjectRefType(s.getObject(req.GetObjectHandle())))
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetObjectRefTypeResponse{RefType: refType}, nil
}

// ---- Method/Field ID lookup ----

func (s *Server) GetMethodID(_ context.Context, req *pb.GetMethodIDRequest) (*pb.GetMethodIDResponse, error) {
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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
	var result *pb.JValue
	if err := s.withEnv(func(env *jni.Env) error {
		obj := s.getObject(req.GetObjectHandle())
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
	var result *pb.JValue
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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
	var result *pb.JValue
	if err := s.withEnv(func(env *jni.Env) error {
		obj := s.getObject(req.GetObjectHandle())
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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
		if v != nil {
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
		if v != nil {
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
		if v != nil {
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
	var value string
	if err := s.withEnv(func(env *jni.Env) error {
		str := (*jni.String)(unsafe.Pointer(s.getObject(req.GetStringHandle())))
		value = env.GoString(str)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.GetStringUTFCharsResponse{Value: value}, nil
}

func (s *Server) GetStringLength(_ context.Context, req *pb.GetStringLengthRequest) (*pb.GetStringLengthResponse, error) {
	var length int32
	if err := s.withEnv(func(env *jni.Env) error {
		str := (*jni.String)(unsafe.Pointer(s.getObject(req.GetStringHandle())))
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
	if err := s.withEnv(func(env *jni.Env) error {
		t := (*jni.Throwable)(unsafe.Pointer(s.getObject(req.GetThrowableHandle())))
		return env.Throw(t)
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.ThrowResponse{}, nil
}

func (s *Server) ThrowNew(_ context.Context, req *pb.ThrowNewRequest) (*pb.ThrowNewResponse, error) {
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
		return env.ThrowNew(cls, req.GetMessage())
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.ThrowNewResponse{}, nil
}

// ---- Monitor ----

func (s *Server) MonitorEnter(_ context.Context, req *pb.MonitorEnterRequest) (*pb.MonitorEnterResponse, error) {
	if err := s.withEnv(func(env *jni.Env) error {
		return env.MonitorEnter(s.getObject(req.GetObjectHandle()))
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.MonitorEnterResponse{}, nil
}

func (s *Server) MonitorExit(_ context.Context, req *pb.MonitorExitRequest) (*pb.MonitorExitResponse, error) {
	if err := s.withEnv(func(env *jni.Env) error {
		return env.MonitorExit(s.getObject(req.GetObjectHandle()))
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
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		obj := s.getObject(req.GetMethodObject())
		mid := env.FromReflectedMethod(obj)
		id = methodIDToInt64(mid)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.FromReflectedMethodResponse{MethodId: id}, nil
}

func (s *Server) FromReflectedField(_ context.Context, req *pb.FromReflectedFieldRequest) (*pb.FromReflectedFieldResponse, error) {
	var id int64
	if err := s.withEnv(func(env *jni.Env) error {
		obj := s.getObject(req.GetFieldObject())
		fid := env.FromReflectedField(obj)
		id = fieldIDToInt64(fid)
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.FromReflectedFieldResponse{FieldId: id}, nil
}

func (s *Server) ToReflectedMethod(_ context.Context, req *pb.ToReflectedMethodRequest) (*pb.ToReflectedMethodResponse, error) {
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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
	var handle int64
	if err := s.withEnv(func(env *jni.Env) error {
		cls := (*jni.Class)(unsafe.Pointer(s.getObject(req.GetClassHandle())))
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

// Remaining RPCs (field access, arrays, references) follow the same pattern.
// They are left unimplemented (returning Unimplemented status via the embedded
// UnimplementedJNIServiceServer) and can be filled in as needed. The proto
// definitions and CLI commands are complete.
