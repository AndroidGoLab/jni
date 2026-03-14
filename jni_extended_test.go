package jni

import (
	"testing"
	"unsafe"
)

// --- Instance field access (using Integer's private 'value' field) ---

// integerWithValue creates a java.lang.Integer with the given value and returns
// the object, its class, and the 'value' field ID.
func integerWithValue(t *testing.T, env *Env, val int32) (*Object, *Class, FieldID) {
	t.Helper()
	cls, err := env.FindClass("java/lang/Integer")
	if err != nil {
		t.Fatalf("FindClass: %v", err)
	}
	valueOf, err := env.GetStaticMethodID(cls, "valueOf", "(I)Ljava/lang/Integer;")
	if err != nil {
		t.Fatalf("GetStaticMethodID: %v", err)
	}
	obj, err := env.CallStaticObjectMethod(cls, valueOf, IntValue(val))
	if err != nil {
		t.Fatalf("CallStaticObjectMethod: %v", err)
	}
	fid, err := env.GetFieldID(cls, "value", "I")
	if err != nil {
		t.Fatalf("GetFieldID: %v", err)
	}
	return obj, cls, fid
}

func TestGetSetIntField(t *testing.T) {
	withEnv(t, func(env *Env) {
		obj, cls, fid := integerWithValue(t, env, 99)
		v := env.GetIntField(obj, fid)
		if v != 99 {
			t.Errorf("GetIntField = %d, want 99", v)
		}
		env.SetIntField(obj, fid, 200)
		v = env.GetIntField(obj, fid)
		if v != 200 {
			t.Errorf("after SetIntField, got %d, want 200", v)
		}
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestGetSetBooleanField(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Boolean")
		fid, err := env.GetFieldID(cls, "value", "Z")
		if err != nil {
			t.Fatalf("GetFieldID: %v", err)
		}
		valueOf, _ := env.GetStaticMethodID(cls, "valueOf", "(Z)Ljava/lang/Boolean;")
		obj, _ := env.CallStaticObjectMethod(cls, valueOf, BooleanValue(1))
		v := env.GetBooleanField(obj, fid)
		if v == 0 {
			t.Error("expected true")
		}
		env.SetBooleanField(obj, fid, 0)
		v = env.GetBooleanField(obj, fid)
		if v != 0 {
			t.Error("expected false after set")
		}
		// Restore the original value. Boolean.valueOf(true) returns the
		// cached Boolean.TRUE singleton, so mutating its 'value' field
		// corrupts global JVM state for all subsequent tests.
		env.SetBooleanField(obj, fid, 1)
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestGetSetByteField(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Byte")
		fid, err := env.GetFieldID(cls, "value", "B")
		if err != nil {
			t.Fatalf("GetFieldID: %v", err)
		}
		valueOf, _ := env.GetStaticMethodID(cls, "valueOf", "(B)Ljava/lang/Byte;")
		obj, _ := env.CallStaticObjectMethod(cls, valueOf, ByteValue(42))
		v := env.GetByteField(obj, fid)
		if v != 42 {
			t.Errorf("got %d, want 42", v)
		}
		env.SetByteField(obj, fid, 7)
		if env.GetByteField(obj, fid) != 7 {
			t.Error("SetByteField failed")
		}
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestGetSetCharField(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Character")
		fid, err := env.GetFieldID(cls, "value", "C")
		if err != nil {
			t.Fatalf("GetFieldID: %v", err)
		}
		valueOf, _ := env.GetStaticMethodID(cls, "valueOf", "(C)Ljava/lang/Character;")
		obj, _ := env.CallStaticObjectMethod(cls, valueOf, CharValue('X'))
		v := env.GetCharField(obj, fid)
		if v != 'X' {
			t.Errorf("got %c, want X", v)
		}
		env.SetCharField(obj, fid, 'Y')
		if env.GetCharField(obj, fid) != 'Y' {
			t.Error("SetCharField failed")
		}
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestGetSetShortField(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Short")
		fid, err := env.GetFieldID(cls, "value", "S")
		if err != nil {
			t.Fatalf("GetFieldID: %v", err)
		}
		valueOf, _ := env.GetStaticMethodID(cls, "valueOf", "(S)Ljava/lang/Short;")
		obj, _ := env.CallStaticObjectMethod(cls, valueOf, ShortValue(100))
		v := env.GetShortField(obj, fid)
		if v != 100 {
			t.Errorf("got %d, want 100", v)
		}
		env.SetShortField(obj, fid, 200)
		if env.GetShortField(obj, fid) != 200 {
			t.Error("SetShortField failed")
		}
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestGetSetLongField(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Long")
		fid, err := env.GetFieldID(cls, "value", "J")
		if err != nil {
			t.Fatalf("GetFieldID: %v", err)
		}
		valueOf, _ := env.GetStaticMethodID(cls, "valueOf", "(J)Ljava/lang/Long;")
		obj, _ := env.CallStaticObjectMethod(cls, valueOf, LongValue(1<<40))
		v := env.GetLongField(obj, fid)
		if v != 1<<40 {
			t.Errorf("got %d, want %d", v, int64(1<<40))
		}
		env.SetLongField(obj, fid, 999)
		if env.GetLongField(obj, fid) != 999 {
			t.Error("SetLongField failed")
		}
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestGetSetFloatField(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Float")
		fid, err := env.GetFieldID(cls, "value", "F")
		if err != nil {
			t.Fatalf("GetFieldID: %v", err)
		}
		valueOf, _ := env.GetStaticMethodID(cls, "valueOf", "(F)Ljava/lang/Float;")
		obj, _ := env.CallStaticObjectMethod(cls, valueOf, FloatValue(3.14))
		v := env.GetFloatField(obj, fid)
		if v < 3.13 || v > 3.15 {
			t.Errorf("got %f", v)
		}
		env.SetFloatField(obj, fid, 1.0)
		if env.GetFloatField(obj, fid) != 1.0 {
			t.Error("SetFloatField failed")
		}
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestGetSetDoubleField(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Double")
		fid, err := env.GetFieldID(cls, "value", "D")
		if err != nil {
			t.Fatalf("GetFieldID: %v", err)
		}
		valueOf, _ := env.GetStaticMethodID(cls, "valueOf", "(D)Ljava/lang/Double;")
		obj, _ := env.CallStaticObjectMethod(cls, valueOf, DoubleValue(2.718))
		v := env.GetDoubleField(obj, fid)
		if v < 2.71 || v > 2.72 {
			t.Errorf("got %f", v)
		}
		env.SetDoubleField(obj, fid, 9.9)
		if env.GetDoubleField(obj, fid) != 9.9 {
			t.Error("SetDoubleField failed")
		}
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestGetSetObjectField(t *testing.T) {
	withEnv(t, func(env *Env) {
		// Use AtomicReference which has a volatile Object value field
		cls, err := env.FindClass("java/util/concurrent/atomic/AtomicReference")
		if err != nil {
			t.Fatalf("FindClass: %v", err)
		}
		ctor, _ := env.GetMethodID(cls, "<init>", "()V")
		obj, _ := env.NewObject(cls, ctor)
		fid, err := env.GetFieldID(cls, "value", "Ljava/lang/Object;")
		if err != nil {
			t.Fatalf("GetFieldID: %v", err)
		}
		str, _ := env.NewStringUTF("hello")
		env.SetObjectField(obj, fid, &str.Object)
		got := env.GetObjectField(obj, fid)
		if got == nil {
			t.Fatal("GetObjectField returned nil")
		}
		if !env.IsSameObject(got, &str.Object) {
			t.Error("field value mismatch")
		}
		env.DeleteLocalRef(got)
		env.DeleteLocalRef(&str.Object)
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

// --- Static field set (all types) using a mutable static ---

func TestSetStaticIntField(t *testing.T) {
	withEnv(t, func(env *Env) {
		// Read Integer.MAX_VALUE, set it, read back, restore.
		cls, _ := env.FindClass("java/lang/Integer")
		fid, _ := env.GetStaticFieldID(cls, "MAX_VALUE", "I")
		orig := env.GetStaticIntField(cls, fid)
		env.SetStaticIntField(cls, fid, 12345)
		v := env.GetStaticIntField(cls, fid)
		if v != 12345 {
			t.Errorf("SetStaticIntField: got %d, want 12345", v)
		}
		env.SetStaticIntField(cls, fid, orig) // restore
		env.DeleteLocalRef(&cls.Object)
	})
}

// --- Instance Call* methods for remaining types ---

func TestCallLongMethod(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Long")
		valueOf, _ := env.GetStaticMethodID(cls, "valueOf", "(J)Ljava/lang/Long;")
		obj, _ := env.CallStaticObjectMethod(cls, valueOf, LongValue(1<<40))
		longValue, _ := env.GetMethodID(cls, "longValue", "()J")
		v, err := env.CallLongMethod(obj, longValue)
		if err != nil {
			t.Fatalf("CallLongMethod: %v", err)
		}
		if v != 1<<40 {
			t.Errorf("got %d", v)
		}
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestCallDoubleMethod(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Double")
		valueOf, _ := env.GetStaticMethodID(cls, "valueOf", "(D)Ljava/lang/Double;")
		obj, _ := env.CallStaticObjectMethod(cls, valueOf, DoubleValue(3.14))
		doubleValue, _ := env.GetMethodID(cls, "doubleValue", "()D")
		v, err := env.CallDoubleMethod(obj, doubleValue)
		if err != nil {
			t.Fatalf("CallDoubleMethod: %v", err)
		}
		if v != 3.14 {
			t.Errorf("got %f", v)
		}
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestCallFloatMethod(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Float")
		valueOf, _ := env.GetStaticMethodID(cls, "valueOf", "(F)Ljava/lang/Float;")
		obj, _ := env.CallStaticObjectMethod(cls, valueOf, FloatValue(2.5))
		floatValue, _ := env.GetMethodID(cls, "floatValue", "()F")
		v, err := env.CallFloatMethod(obj, floatValue)
		if err != nil {
			t.Fatalf("CallFloatMethod: %v", err)
		}
		if v != 2.5 {
			t.Errorf("got %f", v)
		}
		env.DeleteLocalRef(obj)
		env.DeleteLocalRef(&cls.Object)
	})
}

// --- Remaining typed array element access ---

func TestGetBooleanArrayElements(t *testing.T) {
	withEnv(t, func(env *Env) {
		arr := env.NewBooleanArray(2)
		data := [2]uint8{1, 0}
		env.SetBooleanArrayRegion(arr, 0, 2, unsafe.Pointer(&data[0]))
		elems := env.GetBooleanArrayElements(arr, nil)
		if elems == nil {
			t.Fatal("nil")
		}
		got := (*[2]uint8)(elems)
		if got[0] != 1 || got[1] != 0 {
			t.Errorf("got %v", got)
		}
		env.ReleaseBooleanArrayElements(arr, elems, 0)
		env.DeleteLocalRef(&arr.Object)
	})
}

func TestGetByteArrayElements(t *testing.T) {
	withEnv(t, func(env *Env) {
		arr := env.NewByteArray(2)
		data := [2]int8{5, 6}
		env.SetByteArrayRegion(arr, 0, 2, unsafe.Pointer(&data[0]))
		elems := env.GetByteArrayElements(arr, nil)
		got := (*[2]int8)(elems)
		if *got != data {
			t.Errorf("got %v", got)
		}
		env.ReleaseByteArrayElements(arr, elems, 0)
		env.DeleteLocalRef(&arr.Object)
	})
}

func TestGetCharArrayElements(t *testing.T) {
	withEnv(t, func(env *Env) {
		arr := env.NewCharArray(2)
		data := [2]uint16{'X', 'Y'}
		env.SetCharArrayRegion(arr, 0, 2, unsafe.Pointer(&data[0]))
		elems := env.GetCharArrayElements(arr, nil)
		got := (*[2]uint16)(elems)
		if *got != data {
			t.Errorf("got %v", got)
		}
		env.ReleaseCharArrayElements(arr, elems, 0)
		env.DeleteLocalRef(&arr.Object)
	})
}

func TestGetShortArrayElements(t *testing.T) {
	withEnv(t, func(env *Env) {
		arr := env.NewShortArray(2)
		data := [2]int16{10, 20}
		env.SetShortArrayRegion(arr, 0, 2, unsafe.Pointer(&data[0]))
		elems := env.GetShortArrayElements(arr, nil)
		got := (*[2]int16)(elems)
		if *got != data {
			t.Errorf("got %v", got)
		}
		env.ReleaseShortArrayElements(arr, elems, 0)
		env.DeleteLocalRef(&arr.Object)
	})
}

func TestGetLongArrayElements(t *testing.T) {
	withEnv(t, func(env *Env) {
		arr := env.NewLongArray(2)
		data := [2]int64{100, 200}
		env.SetLongArrayRegion(arr, 0, 2, unsafe.Pointer(&data[0]))
		elems := env.GetLongArrayElements(arr, nil)
		got := (*[2]int64)(elems)
		if *got != data {
			t.Errorf("got %v", got)
		}
		env.ReleaseLongArrayElements(arr, elems, 0)
		env.DeleteLocalRef(&arr.Object)
	})
}

func TestGetFloatArrayElements(t *testing.T) {
	withEnv(t, func(env *Env) {
		arr := env.NewFloatArray(2)
		data := [2]float32{1.1, 2.2}
		env.SetFloatArrayRegion(arr, 0, 2, unsafe.Pointer(&data[0]))
		elems := env.GetFloatArrayElements(arr, nil)
		got := (*[2]float32)(elems)
		if *got != data {
			t.Errorf("got %v", got)
		}
		env.ReleaseFloatArrayElements(arr, elems, 0)
		env.DeleteLocalRef(&arr.Object)
	})
}

func TestGetDoubleArrayElements(t *testing.T) {
	withEnv(t, func(env *Env) {
		arr := env.NewDoubleArray(2)
		data := [2]float64{3.3, 4.4}
		env.SetDoubleArrayRegion(arr, 0, 2, unsafe.Pointer(&data[0]))
		elems := env.GetDoubleArrayElements(arr, nil)
		got := (*[2]float64)(elems)
		if *got != data {
			t.Errorf("got %v", got)
		}
		env.ReleaseDoubleArrayElements(arr, elems, 0)
		env.DeleteLocalRef(&arr.Object)
	})
}

// --- Throw (not ThrowNew) ---

func TestThrow(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/RuntimeException")
		ctor, _ := env.GetMethodID(cls, "<init>", "(Ljava/lang/String;)V")
		msg, _ := env.NewStringUTF("test throw")
		exc, _ := env.NewObject(cls, ctor, ObjectValue(&msg.Object))
		throwable := &Throwable{Object{ref: exc.ref}}
		err := env.Throw(throwable)
		if err != nil {
			t.Fatalf("Throw: %v", err)
		}
		if !env.ExceptionCheck() {
			t.Fatal("expected exception")
		}
		env.ExceptionClear()
		env.DeleteLocalRef(&msg.Object)
		env.DeleteLocalRef(&cls.Object)
	})
}

// --- ExceptionDescribe ---

func TestExceptionDescribe(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/RuntimeException")
		_ = env.ThrowNew(cls, "describe test")
		env.ExceptionDescribe() // prints to stderr
		env.ExceptionClear()
		env.DeleteLocalRef(&cls.Object)
	})
}

// --- Nonvirtual methods (remaining typed variants) ---

func TestCallNonvirtualIntMethod(t *testing.T) {
	withEnv(t, func(env *Env) {
		str, _ := env.NewStringUTF("abc")
		objCls, _ := env.FindClass("java/lang/Object")
		mid, _ := env.GetMethodID(objCls, "hashCode", "()I")
		v, err := env.CallNonvirtualIntMethod(&str.Object, objCls, mid)
		if err != nil {
			t.Fatalf("CallNonvirtualIntMethod: %v", err)
		}
		// Object.hashCode() returns identity hash; just verify no error.
		_ = v
		env.DeleteLocalRef(&objCls.Object)
		env.DeleteLocalRef(&str.Object)
	})
}

func TestCallNonvirtualBooleanMethod(t *testing.T) {
	withEnv(t, func(env *Env) {
		str1, _ := env.NewStringUTF("a")
		str2, _ := env.NewStringUTF("a")
		objCls, _ := env.FindClass("java/lang/Object")
		mid, _ := env.GetMethodID(objCls, "equals", "(Ljava/lang/Object;)Z")
		// Object.equals compares identity, not value.
		v, err := env.CallNonvirtualBooleanMethod(&str1.Object, objCls, mid, ObjectValue(&str2.Object))
		if err != nil {
			t.Fatalf("CallNonvirtualBooleanMethod: %v", err)
		}
		// str1 and str2 are different objects, so Object.equals returns false.
		if v != 0 {
			t.Error("expected false for Object.equals on different instances")
		}
		env.DeleteLocalRef(&str2.Object)
		env.DeleteLocalRef(&str1.Object)
		env.DeleteLocalRef(&objCls.Object)
	})
}

func TestCallNonvirtualVoidMethod(t *testing.T) {
	withEnv(t, func(env *Env) {
		sbCls, _ := env.FindClass("java/lang/StringBuilder")
		ctor, _ := env.GetMethodID(sbCls, "<init>", "()V")
		sb, _ := env.NewObject(sbCls, ctor)
		mid, _ := env.GetMethodID(sbCls, "setLength", "(I)V")
		err := env.CallNonvirtualVoidMethod(sb, sbCls, mid, IntValue(0))
		if err != nil {
			t.Fatalf("CallNonvirtualVoidMethod: %v", err)
		}
		env.DeleteLocalRef(sb)
		env.DeleteLocalRef(&sbCls.Object)
	})
}

// --- Static Call remaining types ---

func TestCallStaticByteMethod(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Byte")
		mid, _ := env.GetStaticMethodID(cls, "parseByte", "(Ljava/lang/String;)B")
		str, _ := env.NewStringUTF("42")
		v, err := env.CallStaticByteMethod(cls, mid, ObjectValue(&str.Object))
		if err != nil {
			t.Fatalf("CallStaticByteMethod: %v", err)
		}
		if v != 42 {
			t.Errorf("got %d", v)
		}
		env.DeleteLocalRef(&str.Object)
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestCallStaticCharMethod(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Character")
		mid, _ := env.GetStaticMethodID(cls, "toUpperCase", "(C)C")
		v, err := env.CallStaticCharMethod(cls, mid, CharValue('a'))
		if err != nil {
			t.Fatalf("CallStaticCharMethod: %v", err)
		}
		if v != 'A' {
			t.Errorf("got %c", v)
		}
		env.DeleteLocalRef(&cls.Object)
	})
}

func TestCallStaticShortMethod(t *testing.T) {
	withEnv(t, func(env *Env) {
		cls, _ := env.FindClass("java/lang/Short")
		mid, _ := env.GetStaticMethodID(cls, "parseShort", "(Ljava/lang/String;)S")
		str, _ := env.NewStringUTF("100")
		v, err := env.CallStaticShortMethod(cls, mid, ObjectValue(&str.Object))
		if err != nil {
			t.Fatalf("CallStaticShortMethod: %v", err)
		}
		if v != 100 {
			t.Errorf("got %d", v)
		}
		env.DeleteLocalRef(&str.Object)
		env.DeleteLocalRef(&cls.Object)
	})
}

// --- Error type ---

func TestErrorType(t *testing.T) {
	err := Error(-1)
	if err.Error() != "jni: general error" {
		t.Errorf("Error(-1) = %q", err.Error())
	}
	err = Error(-99)
	if err.Error() == "" {
		t.Error("expected non-empty error string for unknown code")
	}
}

// --- Thread operations ---

func TestVMGetEnv(t *testing.T) {
	withEnv(t, func(env *Env) {
		// Already attached, GetEnv should succeed.
		env2, err := testVM.GetEnv(JNI_VERSION_1_6)
		if err != nil {
			t.Fatalf("GetEnv: %v", err)
		}
		if env2 == nil {
			t.Fatal("GetEnv returned nil")
		}
	})
}
