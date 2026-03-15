package jni

/*
#include <jni.h>
#include <string.h>
#include <stdio.h>

extern jobject goProxyDispatch(JNIEnv *env, jobject thiz, jobject proxy,
                               jobject method, jobjectArray args);

static inline jint proxy_register_natives(JNIEnv *env, jclass cls) {
    JNINativeMethod m = {
        "invoke",
        "(Ljava/lang/Object;Ljava/lang/reflect/Method;[Ljava/lang/Object;)Ljava/lang/Object;",
        (void*)goProxyDispatch
    };
    return (*env)->RegisterNatives(env, cls, &m, 1);
}

extern jobject goProxyDispatchInner(JNIEnv *env, jlong handlerID,
                                    jstring methodName, jobjectArray args,
                                    jobject method);

extern jobject goAbstractDispatch(JNIEnv *env, jclass cls, jlong handlerID,
                                  jstring methodName, jobjectArray args);

static inline jint abstract_register_natives(JNIEnv *env, jclass cls) {
    JNINativeMethod m = {
        "invoke",
        "(JLjava/lang/String;[Ljava/lang/Object;)Ljava/lang/Object;",
        (void*)goAbstractDispatch
    };
    return (*env)->RegisterNatives(env, cls, &m, 1);
}

// proxy_invoke is the full native dispatch: extract fields in C, call Go.
// It handles standard Object methods (hashCode, equals, toString) in C
// to avoid NullPointerException when Java unboxes null to a primitive.
static inline jobject proxy_invoke(JNIEnv *env, jobject thiz, jobject proxy,
                                   jobject method, jobjectArray args,
                                   jfieldID handlerIDField, jmethodID getNameMID) {
    jlong handlerID = (*env)->GetLongField(env, thiz, handlerIDField);
    jstring name = (jstring)(*env)->CallObjectMethod(env, method, getNameMID);

    const char *nameStr = (*env)->GetStringUTFChars(env, name, NULL);
    if (nameStr == NULL) {
        if (name != NULL) (*env)->DeleteLocalRef(env, name);
        return NULL;
    }

    // hashCode() -> Integer: use System.identityHashCode(proxy)
    if (strcmp(nameStr, "hashCode") == 0) {
        (*env)->ReleaseStringUTFChars(env, name, nameStr);
        (*env)->DeleteLocalRef(env, name);
        jclass sysCls = (*env)->FindClass(env, "java/lang/System");
        jmethodID ihcMid = (*env)->GetStaticMethodID(
            env, sysCls, "identityHashCode", "(Ljava/lang/Object;)I");
        jint hash = (*env)->CallStaticIntMethod(env, sysCls, ihcMid, proxy);
        (*env)->DeleteLocalRef(env, sysCls);
        jclass intCls = (*env)->FindClass(env, "java/lang/Integer");
        jmethodID voMid = (*env)->GetStaticMethodID(
            env, intCls, "valueOf", "(I)Ljava/lang/Integer;");
        jobject boxed = (*env)->CallStaticObjectMethod(env, intCls, voMid, hash);
        (*env)->DeleteLocalRef(env, intCls);
        return boxed;
    }

    // equals(Object) -> Boolean: identity comparison
    if (strcmp(nameStr, "equals") == 0) {
        (*env)->ReleaseStringUTFChars(env, name, nameStr);
        (*env)->DeleteLocalRef(env, name);
        jobject other = (args != NULL)
            ? (*env)->GetObjectArrayElement(env, args, 0) : NULL;
        jboolean eq = (*env)->IsSameObject(env, proxy, other);
        jclass boolCls = (*env)->FindClass(env, "java/lang/Boolean");
        jmethodID voMid = (*env)->GetStaticMethodID(
            env, boolCls, "valueOf", "(Z)Ljava/lang/Boolean;");
        jobject boxed = (*env)->CallStaticObjectMethod(env, boolCls, voMid, eq);
        (*env)->DeleteLocalRef(env, boolCls);
        return boxed;
    }

    // toString() -> String
    if (strcmp(nameStr, "toString") == 0) {
        (*env)->ReleaseStringUTFChars(env, name, nameStr);
        (*env)->DeleteLocalRef(env, name);
        char buf[64];
        snprintf(buf, sizeof(buf), "GoProxy@%lld", (long long)handlerID);
        return (*env)->NewStringUTF(env, buf);
    }

    (*env)->ReleaseStringUTFChars(env, name, nameStr);

    // Delegate all other methods to Go, passing the Method object
    // so handlers can inspect return type (e.g. to detect void).
    jobject result = goProxyDispatchInner(env, handlerID, name, args, method);
    (*env)->DeleteLocalRef(env, name);
    return result;
}
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni/capi"
)

func init() {
	proxyNativeRegistrar = registerProxyNativesImpl
	proxyAbstractRegistrar = registerAbstractNativesImpl
}

func registerProxyNativesImpl(envPtr *capi.Env, cls capi.Class) error {
	// Both capi.Env and C.JNIEnv are the same underlying C struct.
	// Both capi.Class and C.jclass are the same underlying C type.
	// Use reinterpret casts via pointers to avoid vet warnings.
	cenv := *(*(*C.JNIEnv))(unsafe.Pointer(&envPtr))
	ccls := *(*C.jclass)(unsafe.Pointer(&cls))

	rc := C.proxy_register_natives(cenv, ccls)
	if rc != 0 {
		return fmt.Errorf("jni: RegisterNatives for GoInvocationHandler.invoke failed (rc=%d)", rc)
	}
	return nil
}

func registerAbstractNativesImpl(envPtr *capi.Env, cls capi.Class) error {
	cenv := *(*(*C.JNIEnv))(unsafe.Pointer(&envPtr))
	ccls := *(*C.jclass)(unsafe.Pointer(&cls))

	rc := C.abstract_register_natives(cenv, ccls)
	if rc != 0 {
		return fmt.Errorf("jni: RegisterNatives for GoAbstractDispatch.invoke failed (rc=%d)", rc)
	}
	return nil
}

//export goProxyDispatch
func goProxyDispatch(
	cenv *C.JNIEnv,
	thiz C.jobject,
	proxy C.jobject,
	method C.jobject,
	args C.jobjectArray,
) C.jobject {
	// Do all JNI field/method access in C to avoid cross-package CGo type casts.
	cfid := *(*C.jfieldID)(unsafe.Pointer(&fidHandlerID))
	cmid := *(*C.jmethodID)(unsafe.Pointer(&midMethodGetName))

	return C.proxy_invoke(cenv, thiz, proxy, method, args, cfid, cmid)
}

//export goProxyDispatchInner
func goProxyDispatchInner(
	cenv *C.JNIEnv,
	handlerID C.jlong,
	methodName C.jstring,
	args C.jobjectArray,
	method C.jobject,
) C.jobject {
	envPtr := *(*(*capi.Env))(unsafe.Pointer(&cenv))

	capiMethodName := *(*capi.String)(unsafe.Pointer(&methodName))
	capiArgs := *(*capi.ObjectArray)(unsafe.Pointer(&args))
	capiMethod := *(*capi.Object)(unsafe.Pointer(&method))

	result := dispatchProxyInvocation(envPtr, int64(handlerID), capiMethodName, capiArgs, capiMethod)

	return *(*C.jobject)(unsafe.Pointer(&result))
}

//export goAbstractDispatch
func goAbstractDispatch(
	cenv *C.JNIEnv,
	cls C.jclass,
	handlerID C.jlong,
	methodName C.jstring,
	args C.jobjectArray,
) C.jobject {
	envPtr := *(*(*capi.Env))(unsafe.Pointer(&cenv))

	capiMethodName := *(*capi.String)(unsafe.Pointer(&methodName))
	capiArgs := *(*capi.ObjectArray)(unsafe.Pointer(&args))

	// Abstract adapters don't have a Method object; pass 0 to fall back
	// to the basic handler in dispatchProxyInvocation.
	result := dispatchProxyInvocation(envPtr, int64(handlerID), capiMethodName, capiArgs, 0)

	return *(*C.jobject)(unsafe.Pointer(&result))
}
