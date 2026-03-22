//go:build !android

package main

func deviceInfo() string {
	return "JNI wrappers require Android.\nBuild with: gogio -target android ."
}
