# Gio + JNI Typed Wrappers

This example demonstrates using this library's typed JNI wrappers from
within a [Gio](https://gioui.org) UI application. It reads Android
device information (manufacturer, model, Android version) via the
`os/build` package and displays it as text labels in a Gio layout.

## What it shows

- Bridging Gio's `app.JavaVM()` into this library via `jni.VMFromUintptr`
- Calling typed wrappers (`build.GetBuildInfo`, `build.GetVersionInfo`)
  from a Gio event loop
- Rendering JNI results in a Gio material-themed UI

## Build

Install `gogio` (the Gio build tool):

```sh
go install gioui.org/cmd/gogio@latest
```

Build the APK:

```sh
gogio -target android -appid center.dx.jni.examples.gio .
```

Install on a connected device:

```sh
adb install -r gio.apk
```
