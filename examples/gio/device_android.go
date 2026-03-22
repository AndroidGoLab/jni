//go:build android

package main

import (
	"fmt"

	"gioui.org/app"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/os/build"
)

func deviceInfo() string {
	vm := jni.VMFromUintptr(app.JavaVM())

	info, err := build.GetBuildInfo(vm)
	if err != nil {
		return fmt.Sprintf("build error: %v", err)
	}
	ver, err := build.GetVersionInfo(vm)
	if err != nil {
		return fmt.Sprintf("version error: %v", err)
	}

	return fmt.Sprintf(
		"Manufacturer: %s\nModel: %s\nAndroid: %s (API %d)",
		info.Manufacturer, info.Model, ver.Release, ver.SDKInt,
	)
}
