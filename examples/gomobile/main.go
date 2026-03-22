//go:build android

// Command gomobile demonstrates bridging gomobile's RunOnJVM into this
// library's typed JNI wrappers to read Android device information.
//
// Build with:
//
//	gomobile build -target=android ./examples/gomobile
package main

import (
	"fmt"
	"log"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"

	"github.com/AndroidGoLab/jni"
	"github.com/AndroidGoLab/jni/os/build"
)

func main() {
	app.Main(func(a app.App) {
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				if e.Crosses(lifecycle.StageAlive) == lifecycle.CrossOn {
					if err := printBuildInfo(); err != nil {
						log.Printf("error: %v", err)
					}
				}
				if e.Crosses(lifecycle.StageAlive) == lifecycle.CrossOff {
					return
				}
			}
		}
	})
}

// printBuildInfo uses RunOnJVM to bridge into this library's typed
// wrappers and prints device information to logcat.
func printBuildInfo() error {
	return app.RunOnJVM(func(vm, env, ctx uintptr) error {
		jniVM := jni.VMFromUintptr(vm)

		info, err := build.GetBuildInfo(jniVM)
		if err != nil {
			return fmt.Errorf("GetBuildInfo: %w", err)
		}
		log.Printf("Build info: %+v", info)

		ver, err := build.GetVersionInfo(jniVM)
		if err != nil {
			return fmt.Errorf("GetVersionInfo: %w", err)
		}
		log.Printf("Version info: %+v", ver)

		return nil
	})
}
