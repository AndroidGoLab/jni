# Getting Started with go-jni

This guide covers the core patterns for using the `jni` library to call Android Java APIs from Go.

## JNI Lifecycle

Every JNI operation must run on a thread attached to the JVM. The `VM.Do()` method handles this automatically:

```go
import "github.com/AndroidGoLab/jni"

err := vm.Do(func(env *jni.Env) error {
    // All JNI calls happen inside this closure.
    // The current goroutine is locked to an OS thread and attached to the JVM.
    // Local references are valid only within this scope.
    cls, err := env.FindClass("android/os/Build")
    if err != nil {
        return fmt.Errorf("find class: %w", err)
    }
    // ... use cls ...
    return nil
})
```

`VM.Do()` locks the goroutine to an OS thread, attaches it to the JVM if needed, runs the closure, and detaches on return.

## Obtaining a VM

In a NativeActivity, the VM pointer comes from the activity struct:

```go
/*
#include <android/native_activity.h>
*/
import "C"

//export ANativeActivity_onCreate
func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {
    vm := jni.VMFromUintptr(uintptr(activity.vm))
    activityObj := jni.ObjectFromUintptr(uintptr(activity.clazz))
    // ... use vm and activityObj ...
}
```

From **Gio**: `vm := jni.VMFromUintptr(uintptr(app.JavaVM()))`.

From **gomobile**: use the VM pointer passed to `RunOnJVM`.

## App Context

Most Android APIs need a `Context`. Create one from a JNI object reference:

```go
import "github.com/AndroidGoLab/jni/app"

// From an existing global reference (e.g., a NativeActivity object):
ctx, err := app.ContextFromObject(vm, globalRef)
if err != nil {
    return err
}
defer ctx.Close()
```

In NativeActivity examples, the shared `ui` package provides a helper:

```go
import "github.com/AndroidGoLab/jni/examples/common/ui"

ctx, err := ui.GetAppContext(vm)
if err != nil {
    return err
}
defer ctx.Close()
```

## System Services (Manager Pattern)

All 53 Android API packages follow the same pattern:

```go
import "github.com/AndroidGoLab/jni/os/battery"

// 1. Create manager from context
mgr, err := battery.NewManager(ctx)
if err != nil {
    return err
}
defer mgr.Close()

// 2. Call typed methods
charging, err := mgr.IsCharging()
capacity, err := mgr.GetIntProperty(int32(battery.BatteryPropertyCapacity))
```

Every manager:
- Takes an `*app.Context` in its constructor
- Stores a `*jni.GlobalRef` to the Java service object
- Provides typed Go methods that wrap JNI calls
- Must be `Close()`d to release the global reference

## Reference Management

JNI has two reference types:

- **Local references**: Valid only within a single `VM.Do()` call. Automatically freed when the closure returns.
- **Global references**: Valid across `VM.Do()` calls. Must be explicitly freed with `env.DeleteGlobalRef()`.

When you need an object to outlive a `VM.Do()` scope, convert it:

```go
var globalObj *jni.Object
vm.Do(func(env *jni.Env) error {
    localObj, err := env.CallObjectMethod(someObj, someMethod)
    if err != nil {
        return err
    }
    // Convert local ref to global ref
    globalObj = env.NewGlobalRef(localObj)
    env.DeleteLocalRef(localObj)
    return nil
})
// globalObj is valid here, outside VM.Do()

// Later, when done:
vm.Do(func(env *jni.Env) error {
    env.DeleteGlobalRef(globalObj)
    return nil
})
```

## String Conversions

Go strings to JNI and back:

```go
vm.Do(func(env *jni.Env) error {
    // Go string -> JNI String
    jStr, err := env.NewStringUTF("hello")
    if err != nil {
        return err
    }
    defer env.DeleteLocalRef(&jStr.Object)

    // JNI String -> Go string
    goStr := env.GoString(jStr)
    fmt.Println(goStr) // "hello"
    return nil
})
```

## Error Handling

Java exceptions are automatically converted to Go errors:

```go
vm.Do(func(env *jni.Env) error {
    result, err := env.CallObjectMethod(obj, method)
    if err != nil {
        // err contains the Java exception message
        return fmt.Errorf("call failed: %w", err)
    }
    return nil
})
```

## JNI Value Types

When calling methods with arguments, wrap values using typed constructors:

```go
// Primitive types
jni.IntValue(42)
jni.LongValue(int64(1000))
jni.FloatValue(float32(3.14))
jni.BooleanValue(true)

// Object types (including strings)
jStr, _ := env.NewStringUTF("hello")
jni.ObjectValue(&jStr.Object)
```

## Complete Example: Battery Status

```go
import (
    "github.com/AndroidGoLab/jni"
    "github.com/AndroidGoLab/jni/app"
    "github.com/AndroidGoLab/jni/os/battery"
)

func getBatteryInfo(vm *jni.VM, activityRef *jni.GlobalRef) error {
    ctx, err := app.ContextFromObject(vm, activityRef)
    if err != nil {
        return err
    }
    defer ctx.Close()

    mgr, err := battery.NewManager(ctx)
    if err != nil {
        return err
    }
    defer mgr.Close()

    charging, _ := mgr.IsCharging()
    capacity, _ := mgr.GetIntProperty(int32(battery.BatteryPropertyCapacity))
    fmt.Printf("Battery: %d%%, charging: %v\n", capacity, charging)
    return nil
}
```
