package main

import (
	"context"
	"fmt"

	pb "github.com/xaionaro-go/jni/proto/jni_raw"
)

// jniCaller wraps the raw JNI gRPC client and provides concise methods
// for common JNI operations (find class, get method, call method, etc.).
type jniCaller struct {
	client pb.JNIServiceClient
	ctx    context.Context
}

func (j *jniCaller) getAppContext() (int64, error) {
	resp, err := j.client.GetAppContext(j.ctx, &pb.GetAppContextRequest{})
	if err != nil {
		return 0, fmt.Errorf("GetAppContext: %w", err)
	}
	return resp.GetContextHandle(), nil
}

func (j *jniCaller) findClass(name string) (int64, error) {
	resp, err := j.client.FindClass(j.ctx, &pb.FindClassRequest{Name: name})
	if err != nil {
		return 0, fmt.Errorf("FindClass(%q): %w", name, err)
	}
	return resp.GetClassHandle(), nil
}

func (j *jniCaller) getMethodID(
	cls int64,
	name, sig string,
) (int64, error) {
	resp, err := j.client.GetMethodID(j.ctx, &pb.GetMethodIDRequest{
		ClassHandle: cls,
		Name:        name,
		Sig:         sig,
	})
	if err != nil {
		return 0, fmt.Errorf("GetMethodID(%d, %q, %q): %w", cls, name, sig, err)
	}
	return resp.GetMethodId(), nil
}

func (j *jniCaller) getStaticMethodID(
	cls int64,
	name, sig string,
) (int64, error) {
	resp, err := j.client.GetStaticMethodID(j.ctx, &pb.GetStaticMethodIDRequest{
		ClassHandle: cls,
		Name:        name,
		Sig:         sig,
	})
	if err != nil {
		return 0, fmt.Errorf("GetStaticMethodID(%d, %q, %q): %w", cls, name, sig, err)
	}
	return resp.GetMethodId(), nil
}

func (j *jniCaller) callMethod(
	obj, method int64,
	retType pb.JType,
	args ...*pb.JValue,
) (*pb.JValue, error) {
	resp, err := j.client.CallMethod(j.ctx, &pb.CallMethodRequest{
		ObjectHandle: obj,
		MethodId:     method,
		ReturnType:   retType,
		Args:         args,
	})
	if err != nil {
		return nil, fmt.Errorf("CallMethod(obj=%d, mid=%d): %w", obj, method, err)
	}
	return resp.GetResult(), nil
}

func (j *jniCaller) callStaticMethod(
	cls, method int64,
	retType pb.JType,
	args ...*pb.JValue,
) (*pb.JValue, error) {
	resp, err := j.client.CallStaticMethod(j.ctx, &pb.CallStaticMethodRequest{
		ClassHandle: cls,
		MethodId:    method,
		ReturnType:  retType,
		Args:        args,
	})
	if err != nil {
		return nil, fmt.Errorf("CallStaticMethod(cls=%d, mid=%d): %w", cls, method, err)
	}
	return resp.GetResult(), nil
}

func (j *jniCaller) callVoidMethod(
	obj, method int64,
	args ...*pb.JValue,
) error {
	_, err := j.callMethod(obj, method, pb.JType_VOID, args...)
	return err
}

func (j *jniCaller) callObjectMethod(
	obj, method int64,
	args ...*pb.JValue,
) (int64, error) {
	v, err := j.callMethod(obj, method, pb.JType_OBJECT, args...)
	if err != nil {
		return 0, err
	}
	return v.GetL(), nil
}

func (j *jniCaller) callIntMethod(
	obj, method int64,
	args ...*pb.JValue,
) (int32, error) {
	v, err := j.callMethod(obj, method, pb.JType_INT, args...)
	if err != nil {
		return 0, err
	}
	return v.GetI(), nil
}

func (j *jniCaller) newString(s string) (int64, error) {
	resp, err := j.client.NewStringUTF(j.ctx, &pb.NewStringUTFRequest{Value: s})
	if err != nil {
		return 0, fmt.Errorf("NewStringUTF(%q): %w", s, err)
	}
	return resp.GetStringHandle(), nil
}

func (j *jniCaller) newObject(
	cls, constructor int64,
	args ...*pb.JValue,
) (int64, error) {
	resp, err := j.client.NewObject(j.ctx, &pb.NewObjectRequest{
		ClassHandle: cls,
		MethodId:    constructor,
		Args:        args,
	})
	if err != nil {
		return 0, fmt.Errorf("NewObject(cls=%d, ctor=%d): %w", cls, constructor, err)
	}
	return resp.GetObjectHandle(), nil
}

func (j *jniCaller) getObjectArrayElement(
	arrayHandle int64,
	index int32,
) (int64, error) {
	resp, err := j.client.GetObjectArrayElement(j.ctx, &pb.GetObjectArrayElementRequest{
		ArrayHandle: arrayHandle,
		Index:       index,
	})
	if err != nil {
		return 0, fmt.Errorf("GetObjectArrayElement(arr=%d, idx=%d): %w", arrayHandle, index, err)
	}
	return resp.GetElementHandle(), nil
}

func (j *jniCaller) newObjectArray(
	length int32,
	classHandle int64,
	initElement int64,
) (int64, error) {
	resp, err := j.client.NewObjectArray(j.ctx, &pb.NewObjectArrayRequest{
		Length:      length,
		ClassHandle: classHandle,
		InitElement: initElement,
	})
	if err != nil {
		return 0, fmt.Errorf("NewObjectArray(len=%d, cls=%d): %w", length, classHandle, err)
	}
	return resp.GetArrayHandle(), nil
}

func (j *jniCaller) setObjectArrayElement(
	arrayHandle int64,
	index int32,
	elementHandle int64,
) error {
	_, err := j.client.SetObjectArrayElement(j.ctx, &pb.SetObjectArrayElementRequest{
		ArrayHandle:   arrayHandle,
		Index:         index,
		ElementHandle: elementHandle,
	})
	if err != nil {
		return fmt.Errorf("SetObjectArrayElement(arr=%d, idx=%d, el=%d): %w", arrayHandle, index, elementHandle, err)
	}
	return nil
}

func (j *jniCaller) getByteArrayData(arrayHandle int64) ([]byte, error) {
	resp, err := j.client.GetByteArrayData(j.ctx, &pb.GetByteArrayDataRequest{
		ArrayHandle: arrayHandle,
	})
	if err != nil {
		return nil, fmt.Errorf("GetByteArrayData(arr=%d): %w", arrayHandle, err)
	}
	return resp.GetData(), nil
}

func (j *jniCaller) getStringUTFChars(stringHandle int64) (string, error) {
	resp, err := j.client.GetStringUTFChars(j.ctx, &pb.GetStringUTFCharsRequest{
		StringHandle: stringHandle,
	})
	if err != nil {
		return "", fmt.Errorf("GetStringUTFChars(str=%d): %w", stringHandle, err)
	}
	return resp.GetValue(), nil
}

// objVal is a shorthand for creating a JValue with an object handle.
func objVal(handle int64) *pb.JValue {
	return &pb.JValue{Value: &pb.JValue_L{L: handle}}
}

// intVal is a shorthand for creating a JValue with an int.
func intVal(v int32) *pb.JValue {
	return &pb.JValue{Value: &pb.JValue_I{I: v}}
}

// longVal is a shorthand for creating a JValue with a long.
func longVal(v int64) *pb.JValue {
	return &pb.JValue{Value: &pb.JValue_J{J: v}}
}

// boolVal is a shorthand for creating a JValue with a boolean.
func boolVal(v bool) *pb.JValue {
	return &pb.JValue{Value: &pb.JValue_Z{Z: v}}
}
