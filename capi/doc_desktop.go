//go:build !android && cgo

// Desktop builds require CGO_CFLAGS and CGO_LDFLAGS pointing to a JDK:
//
//	export CGO_CFLAGS="-I$JAVA_HOME/include -I$JAVA_HOME/include/linux"
//	export CGO_LDFLAGS="-L$JAVA_HOME/lib/server -ljvm"
package capi

// #include <jni.h>
import "C"
