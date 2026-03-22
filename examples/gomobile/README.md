# gomobile integration example

Demonstrates bridging gomobile's `app.RunOnJVM` into this library's typed
JNI wrappers to read Android device build information without any Java code.

## What it does

On app startup, the example calls `app.RunOnJVM` to obtain the JVM pointer,
converts it with `jni.VMFromUintptr`, and uses the `os/build` package to read
`android.os.Build` and `android.os.Build.VERSION` fields. Results are printed
to logcat.

## Prerequisites

Install gomobile:

```sh
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init
```

## Build

From the repository root:

```sh
gomobile build -target=android ./examples/gomobile
```

This produces a `gomobile.apk` file.

## Install and run

```sh
adb install gomobile.apk
adb shell am start -n org.golang.app.gomobile/.GoNativeActivity
```

## View output

```sh
adb logcat -s GoLog:*
```
