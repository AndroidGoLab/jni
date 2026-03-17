package jni

import (
	"errors"
	"testing"

	"github.com/AndroidGoLab/jni/capi"
)

// --- proxy registry (pure Go) ---

func TestRegisterLookupUnregisterProxy(t *testing.T) {
	called := false
	h := ProxyHandler(func(env *Env, method string, args []*Object) (*Object, error) {
		called = true
		return nil, nil
	})

	id := registerProxy(h)
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, ok := lookupProxy(id)
	if !ok {
		t.Fatal("lookupProxy returned false for registered handler")
	}
	if got == nil {
		t.Fatal("handler is nil")
	}

	// Invoke to verify it is the right handler.
	_, _ = got(nil, "", nil)
	if !called {
		t.Fatal("handler was not invoked")
	}

	unregisterProxy(id)

	_, ok = lookupProxy(id)
	if ok {
		t.Fatal("lookupProxy returned true after unregister")
	}
}

func TestLookupProxyNotFound(t *testing.T) {
	_, ok := lookupProxy(-99999)
	if ok {
		t.Fatal("expected false for non-existent handler")
	}
}

// --- cstringLiteral ---

func TestCstringLiteral(t *testing.T) {
	ptr := cstringLiteral("hello")
	if ptr == nil {
		t.Fatal("cstringLiteral returned nil")
	}
}

func TestCstringLiteralEmpty(t *testing.T) {
	ptr := cstringLiteral("")
	if ptr == nil {
		t.Fatal("cstringLiteral returned nil for empty string")
	}
}

// --- extractGoString ---

func TestExtractGoString(t *testing.T) {
	withEnv(t, func(env *Env) {
		str, err := env.NewStringUTF("hello world")
		if err != nil {
			t.Fatalf("NewStringUTF: %v", err)
		}
		result := extractGoString(env.ptr, capi.String(str.ref))
		if result != "hello world" {
			t.Errorf("extractGoString = %q, want %q", result, "hello world")
		}
		env.DeleteLocalRef(&str.Object)
	})
}

func TestExtractGoStringNull(t *testing.T) {
	withEnv(t, func(env *Env) {
		result := extractGoString(env.ptr, 0)
		if result != "" {
			t.Errorf("extractGoString(null) = %q, want empty", result)
		}
	})
}

func TestExtractGoStringEmpty(t *testing.T) {
	withEnv(t, func(env *Env) {
		str, err := env.NewStringUTF("")
		if err != nil {
			t.Fatalf("NewStringUTF: %v", err)
		}
		result := extractGoString(env.ptr, capi.String(str.ref))
		if result != "" {
			t.Errorf("extractGoString(empty) = %q, want empty", result)
		}
		env.DeleteLocalRef(&str.Object)
	})
}

// --- ensureProxyInit ---

func TestEnsureProxyInit(t *testing.T) {
	withEnv(t, func(env *Env) {
		err := ensureProxyInit(env)
		if err != nil {
			t.Fatalf("ensureProxyInit: %v", err)
		}
		// Calling again should succeed (idempotent via sync.Once).
		err = ensureProxyInit(env)
		if err != nil {
			t.Fatalf("ensureProxyInit (second call): %v", err)
		}
	})
}

// --- throwGoError ---

func TestThrowGoError(t *testing.T) {
	withEnv(t, func(env *Env) {
		throwGoError(env, errors.New("test error from Go"))
		if !env.ExceptionCheck() {
			t.Fatal("expected pending exception after throwGoError")
		}
		env.ExceptionClear()
	})
}

// --- dispatchProxyInvocation ---

func TestDispatchProxyInvocationNilArgs(t *testing.T) {
	withEnv(t, func(env *Env) {
		if err := ensureProxyInit(env); err != nil {
			t.Fatalf("ensureProxyInit: %v", err)
		}

		var receivedMethod string
		var receivedArgs []*Object
		h := ProxyHandler(func(e *Env, method string, args []*Object) (*Object, error) {
			receivedMethod = method
			receivedArgs = args
			return nil, nil
		})
		id := registerProxy(h)
		defer unregisterProxy(id)

		methodStr, err := env.NewStringUTF("testMethod")
		if err != nil {
			t.Fatalf("NewStringUTF: %v", err)
		}
		defer env.DeleteLocalRef(&methodStr.Object)

		result := dispatchProxyInvocation(env.ptr, id, capi.String(methodStr.ref), 0, 0)
		if result != 0 {
			t.Errorf("expected null result, got %v", result)
		}
		if receivedMethod != "testMethod" {
			t.Errorf("method = %q, want %q", receivedMethod, "testMethod")
		}
		if receivedArgs != nil {
			t.Errorf("args = %v, want nil", receivedArgs)
		}
	})
}

func TestDispatchProxyInvocationWithArgs(t *testing.T) {
	withEnv(t, func(env *Env) {
		if err := ensureProxyInit(env); err != nil {
			t.Fatalf("ensureProxyInit: %v", err)
		}

		var receivedArgs []*Object
		h := ProxyHandler(func(e *Env, method string, args []*Object) (*Object, error) {
			receivedArgs = args
			return nil, nil
		})
		id := registerProxy(h)
		defer unregisterProxy(id)

		methodStr, err := env.NewStringUTF("withArgs")
		if err != nil {
			t.Fatalf("NewStringUTF: %v", err)
		}
		defer env.DeleteLocalRef(&methodStr.Object)

		objCls, err := env.FindClass("java/lang/Object")
		if err != nil {
			t.Fatalf("FindClass: %v", err)
		}
		defer env.DeleteLocalRef(&objCls.Object)

		argStr, err := env.NewStringUTF("argValue")
		if err != nil {
			t.Fatalf("NewStringUTF: %v", err)
		}
		defer env.DeleteLocalRef(&argStr.Object)

		argsArray, err := env.NewObjectArray(1, objCls, &argStr.Object)
		if err != nil {
			t.Fatalf("NewObjectArray: %v", err)
		}
		defer env.DeleteLocalRef(&argsArray.Object)

		result := dispatchProxyInvocation(
			env.ptr, id,
			capi.String(methodStr.ref),
			capi.ObjectArray(argsArray.ref),
			0,
		)
		if result != 0 {
			t.Errorf("expected null result, got %v", result)
		}
		if len(receivedArgs) != 1 {
			t.Fatalf("len(args) = %d, want 1", len(receivedArgs))
		}
		if receivedArgs[0] == nil {
			t.Fatal("args[0] is nil")
		}
	})
}

func TestDispatchProxyInvocationReturnsObject(t *testing.T) {
	withEnv(t, func(env *Env) {
		if err := ensureProxyInit(env); err != nil {
			t.Fatalf("ensureProxyInit: %v", err)
		}

		h := ProxyHandler(func(e *Env, method string, args []*Object) (*Object, error) {
			str, err := e.NewStringUTF("result")
			if err != nil {
				return nil, err
			}
			return &str.Object, nil
		})
		id := registerProxy(h)
		defer unregisterProxy(id)

		methodStr, err := env.NewStringUTF("getResult")
		if err != nil {
			t.Fatalf("NewStringUTF: %v", err)
		}
		defer env.DeleteLocalRef(&methodStr.Object)

		result := dispatchProxyInvocation(env.ptr, id, capi.String(methodStr.ref), 0, 0)
		if result == 0 {
			t.Fatal("expected non-null result")
		}
		capi.DeleteLocalRef(env.ptr, result)
	})
}

func TestDispatchProxyInvocationHandlerError(t *testing.T) {
	withEnv(t, func(env *Env) {
		if err := ensureProxyInit(env); err != nil {
			t.Fatalf("ensureProxyInit: %v", err)
		}

		h := ProxyHandler(func(e *Env, method string, args []*Object) (*Object, error) {
			return nil, errors.New("handler failed")
		})
		id := registerProxy(h)
		defer unregisterProxy(id)

		methodStr, err := env.NewStringUTF("failMethod")
		if err != nil {
			t.Fatalf("NewStringUTF: %v", err)
		}
		defer env.DeleteLocalRef(&methodStr.Object)

		result := dispatchProxyInvocation(env.ptr, id, capi.String(methodStr.ref), 0, 0)
		if result != 0 {
			t.Error("expected null result on error")
		}
		if !env.ExceptionCheck() {
			t.Fatal("expected pending exception after handler error")
		}
		env.ExceptionClear()
	})
}

func TestDispatchProxyInvocationUnregisteredHandler(t *testing.T) {
	withEnv(t, func(env *Env) {
		if err := ensureProxyInit(env); err != nil {
			t.Fatalf("ensureProxyInit: %v", err)
		}

		methodStr, err := env.NewStringUTF("noHandler")
		if err != nil {
			t.Fatalf("NewStringUTF: %v", err)
		}
		defer env.DeleteLocalRef(&methodStr.Object)

		result := dispatchProxyInvocation(env.ptr, -777, capi.String(methodStr.ref), 0, 0)
		if result != 0 {
			t.Error("expected null result for unregistered handler")
		}
	})
}

// --- NewProxy (integration) ---

func TestNewProxy(t *testing.T) {
	withEnv(t, func(env *Env) {
		runnableCls, err := env.FindClass("java/lang/Runnable")
		if err != nil {
			t.Fatalf("FindClass(Runnable): %v", err)
		}
		defer env.DeleteLocalRef(&runnableCls.Object)

		proxy, cleanup, err := env.NewProxy(
			[]*Class{runnableCls},
			func(e *Env, method string, args []*Object) (*Object, error) {
				return nil, nil
			},
		)
		if err != nil {
			t.Fatalf("NewProxy: %v", err)
		}
		defer cleanup()
		defer env.DeleteLocalRef(proxy)

		if proxy == nil {
			t.Fatal("proxy is nil")
		}
		if !env.IsInstanceOf(proxy, runnableCls) {
			t.Error("proxy does not implement Runnable")
		}
	})
}

// TestNewProxyCallbackDispatches verifies that invoking a method on a Java
// proxy actually dispatches to the Go handler via the native bridge.
func TestNewProxyCallbackDispatches(t *testing.T) {
	withEnv(t, func(env *Env) {
		runnableCls, err := env.FindClass("java/lang/Runnable")
		if err != nil {
			t.Fatalf("FindClass(Runnable): %v", err)
		}
		defer env.DeleteLocalRef(&runnableCls.Object)

		var called bool
		var receivedMethod string
		proxy, cleanup, err := env.NewProxy(
			[]*Class{runnableCls},
			func(e *Env, method string, args []*Object) (*Object, error) {
				called = true
				receivedMethod = method
				return nil, nil
			},
		)
		if err != nil {
			t.Fatalf("NewProxy: %v", err)
		}
		defer cleanup()
		defer env.DeleteLocalRef(proxy)

		// Invoke Runnable.run() on the proxy from Java.
		runMID, err := env.GetMethodID(runnableCls, "run", "()V")
		if err != nil {
			t.Fatalf("GetMethodID(run): %v", err)
		}
		if err := env.CallVoidMethod(proxy, runMID); err != nil {
			t.Fatalf("CallVoidMethod(run): %v", err)
		}

		if !called {
			t.Fatal("Go handler was not called when Java proxy method was invoked")
		}
		if receivedMethod != "run" {
			t.Errorf("received method = %q, want %q", receivedMethod, "run")
		}
	})
}

// TestNewProxyObjectMethods verifies that hashCode/equals/toString on a proxy
// are handled correctly (returning boxed primitives, not null) to avoid
// NullPointerException when Java unboxes the result.
func TestNewProxyObjectMethods(t *testing.T) {
	withEnv(t, func(env *Env) {
		runnableCls, err := env.FindClass("java/lang/Runnable")
		if err != nil {
			t.Fatalf("FindClass(Runnable): %v", err)
		}
		defer env.DeleteLocalRef(&runnableCls.Object)

		proxy, cleanup, err := env.NewProxy(
			[]*Class{runnableCls},
			func(e *Env, method string, args []*Object) (*Object, error) {
				return nil, nil
			},
		)
		if err != nil {
			t.Fatalf("NewProxy: %v", err)
		}
		defer cleanup()
		defer env.DeleteLocalRef(proxy)

		// Call hashCode() — must return a boxed Integer, not null.
		objCls, err := env.FindClass("java/lang/Object")
		if err != nil {
			t.Fatalf("FindClass(Object): %v", err)
		}
		defer env.DeleteLocalRef(&objCls.Object)

		hashMID, err := env.GetMethodID(objCls, "hashCode", "()I")
		if err != nil {
			t.Fatalf("GetMethodID(hashCode): %v", err)
		}
		hash, err := env.CallIntMethod(proxy, hashMID)
		if err != nil {
			t.Fatalf("hashCode() threw exception: %v", err)
		}
		_ = hash // just verify no NPE

		// Call toString() — must return a non-null String.
		toStringMID, err := env.GetMethodID(objCls, "toString", "()Ljava/lang/String;")
		if err != nil {
			t.Fatalf("GetMethodID(toString): %v", err)
		}
		strObj, err := env.CallObjectMethod(proxy, toStringMID)
		if err != nil {
			t.Fatalf("toString() threw exception: %v", err)
		}
		if strObj == nil || strObj.Ref() == 0 {
			t.Fatal("toString() returned null")
		}
		env.DeleteLocalRef(strObj)

		// Call equals(self) — must return true, not NPE.
		equalsMID, err := env.GetMethodID(objCls, "equals", "(Ljava/lang/Object;)Z")
		if err != nil {
			t.Fatalf("GetMethodID(equals): %v", err)
		}
		eq, err := env.CallBooleanMethod(proxy, equalsMID, ObjectValue(proxy))
		if err != nil {
			t.Fatalf("equals() threw exception: %v", err)
		}
		if eq == 0 {
			t.Error("equals(self) returned false, want true")
		}
	})
}

func TestNewProxyNoInterfaces(t *testing.T) {
	withEnv(t, func(env *Env) {
		_, _, err := env.NewProxy(nil, func(e *Env, method string, args []*Object) (*Object, error) {
			return nil, nil
		})
		if err == nil {
			t.Fatal("expected error for empty interface list")
		}
	})
}

func TestNewProxyMultipleInterfaces(t *testing.T) {
	withEnv(t, func(env *Env) {
		runnableCls, err := env.FindClass("java/lang/Runnable")
		if err != nil {
			t.Fatalf("FindClass(Runnable): %v", err)
		}
		defer env.DeleteLocalRef(&runnableCls.Object)

		callableCls, err := env.FindClass("java/util/concurrent/Callable")
		if err != nil {
			t.Fatalf("FindClass(Callable): %v", err)
		}
		defer env.DeleteLocalRef(&callableCls.Object)

		proxy, cleanup, err := env.NewProxy(
			[]*Class{runnableCls, callableCls},
			func(e *Env, method string, args []*Object) (*Object, error) {
				return nil, nil
			},
		)
		if err != nil {
			t.Fatalf("NewProxy: %v", err)
		}
		defer cleanup()
		defer env.DeleteLocalRef(proxy)

		if !env.IsInstanceOf(proxy, runnableCls) {
			t.Error("proxy does not implement Runnable")
		}
		if !env.IsInstanceOf(proxy, callableCls) {
			t.Error("proxy does not implement Callable")
		}
	})
}
