# Field Access via JNI

This guide covers reading Java object fields and static class fields from Go using the `jni` library.

## Instance Field Access

Use `GetFieldID` to look up a field, then a typed getter to read its value:

```go
vm.Do(func(env *jni.Env) error {
    cls, err := env.FindClass("android/graphics/Point")
    if err != nil {
        return err
    }

    // Look up field IDs by name and JNI type signature
    xField, err := env.GetFieldID(cls, "x", "I")  // int field
    if err != nil {
        return err
    }
    yField, err := env.GetFieldID(cls, "y", "I")
    if err != nil {
        return err
    }

    // Read the field values from an instance
    x := env.GetIntField(pointObj, xField)
    y := env.GetIntField(pointObj, yField)
    fmt.Printf("Point: (%d, %d)\n", x, y)
    return nil
})
```

### Available Field Getters

| Method | Return Type | JNI Signature |
|--------|------------|---------------|
| `env.GetBooleanField(obj, fid)` | `uint8` (0 or 1) | `"Z"` |
| `env.GetByteField(obj, fid)` | `int8` | `"B"` |
| `env.GetCharField(obj, fid)` | `uint16` | `"C"` |
| `env.GetShortField(obj, fid)` | `int16` | `"S"` |
| `env.GetIntField(obj, fid)` | `int32` | `"I"` |
| `env.GetLongField(obj, fid)` | `int64` | `"J"` |
| `env.GetFloatField(obj, fid)` | `float32` | `"F"` |
| `env.GetDoubleField(obj, fid)` | `float64` | `"D"` |
| `env.GetObjectField(obj, fid)` | `*jni.Object` | `"L<class>;"` |

### Reading String Fields

String fields are object fields. Cast the result to `*jni.String` for conversion:

```go
nameField, err := env.GetFieldID(cls, "name", "Ljava/lang/String;")
if err != nil {
    return err
}
nameObj := env.GetObjectField(obj, nameField)
name := env.GoString((*jni.String)(unsafe.Pointer(nameObj)))
```

## Static Field Access

Static fields use `GetStaticFieldID` and `GetStatic*Field`:

```go
// Example: reading android.os.Build static fields
vm.Do(func(env *jni.Env) error {
    cls, err := env.FindClass("android/os/Build")
    if err != nil {
        return err
    }

    // Read a static String field
    fid, err := env.GetStaticFieldID(cls, "MODEL", "Ljava/lang/String;")
    if err != nil {
        return err
    }
    obj := env.GetStaticObjectField(cls, fid)
    model := env.GoString((*jni.String)(unsafe.Pointer(obj)))
    fmt.Println("Model:", model)

    return nil
})
```

### Reading Static Int Fields

```go
// Example: reading android.os.Build.VERSION.SDK_INT
vm.Do(func(env *jni.Env) error {
    cls, err := env.FindClass("android/os/Build$VERSION")
    if err != nil {
        return err
    }
    sdkFid, err := env.GetStaticFieldID(cls, "SDK_INT", "I")
    if err != nil {
        return err
    }
    sdkInt := env.GetStaticIntField(cls, sdkFid)
    fmt.Printf("API level: %d\n", sdkInt)
    return nil
})
```

## Complete Example: Reading Build Info

The `os/build` package demonstrates reading multiple static fields into a Go struct:

```go
import "github.com/AndroidGoLab/jni/os/build"

// GetBuildInfo reads static fields from android.os.Build.
info, err := build.GetBuildInfo(vm)
// info.Device, info.Model, info.Product, info.Manufacturer,
// info.Brand, info.Board, info.Hardware

// GetVersionInfo reads android.os.Build.VERSION fields.
ver, err := build.GetVersionInfo(vm)
// ver.SDKInt (int32), ver.Release, ver.Codename, ver.Incremental
```

Under the hood, `GetBuildInfo` uses this pattern for each field:

```go
func readStaticStringField(env *jni.Env, cls *jni.Class, name string) (string, error) {
    fid, err := env.GetStaticFieldID(cls, name, "Ljava/lang/String;")
    if err != nil {
        return "", fmt.Errorf("get field %s: %w", name, err)
    }
    obj := env.GetStaticObjectField(cls, fid)
    return env.GoString((*jni.String)(unsafe.Pointer(obj))), nil
}
```

## JNI Type Signatures Reference

| Java Type | JNI Signature |
|-----------|--------------|
| `boolean` | `Z` |
| `byte` | `B` |
| `char` | `C` |
| `short` | `S` |
| `int` | `I` |
| `long` | `J` |
| `float` | `F` |
| `double` | `D` |
| `void` | `V` |
| `String` | `Ljava/lang/String;` |
| `Object` | `Ljava/lang/Object;` |
| `int[]` | `[I` |
| `String[]` | `[Ljava/lang/String;` |

### Method Signature Format

Method signatures follow the pattern `(param-types)return-type`:

| Java Method | JNI Signature |
|-------------|--------------|
| `int getX()` | `()I` |
| `void setName(String s)` | `(Ljava/lang/String;)V` |
| `boolean equals(Object o)` | `(Ljava/lang/Object;)Z` |
| `String[] getNames(int count)` | `(I)[Ljava/lang/String;` |
| `void requestLocationUpdates(String, long, float, LocationListener, Looper)` | `(Ljava/lang/String;JFLandroid/location/LocationListener;Landroid/os/Looper;)V` |
