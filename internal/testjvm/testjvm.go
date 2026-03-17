// Package testjvm creates and manages a JVM instance for use in tests.
package testjvm

// #cgo CFLAGS: -I/usr/lib/jvm/java-25-openjdk-amd64/include -I/usr/lib/jvm/java-25-openjdk-amd64/include/linux
// #cgo LDFLAGS: -L/usr/lib/jvm/java-25-openjdk-amd64/lib/server -ljvm
// #include <jni.h>
// #include <stdlib.h>
//
// static int createJVM(JavaVM **vm, JNIEnv **env, const char *classpath) {
//     JavaVMInitArgs args;
//     JavaVMOption opts[1];
//     args.version = JNI_VERSION_1_8;
//     args.ignoreUnrecognized = JNI_TRUE;
//     if (classpath != NULL) {
//         opts[0].optionString = (char*)classpath;
//         opts[0].extraInfo = NULL;
//         args.nOptions = 1;
//         args.options = opts;
//     } else {
//         args.nOptions = 0;
//         args.options = NULL;
//     }
//     return JNI_CreateJavaVM(vm, (void**)env, &args);
// }
import "C"

import (
	"fmt"
	"unsafe"
)

// Create creates a new JVM and returns raw pointers to the JavaVM and JNIEnv.
// An optional classpath can be provided (e.g., "/path/to/classes").
func Create(classpath ...string) (vmPtr, envPtr unsafe.Pointer, err error) {
	var cVM *C.JavaVM
	var cEnv *C.JNIEnv
	var rc C.int
	if len(classpath) > 0 && classpath[0] != "" {
		cp := C.CString("-Djava.class.path=" + classpath[0])
		defer C.free(unsafe.Pointer(cp))
		rc = C.createJVM(&cVM, &cEnv, cp)
	} else {
		rc = C.createJVM(&cVM, &cEnv, nil)
	}
	if rc != 0 {
		return nil, nil, fmt.Errorf("JNI_CreateJavaVM failed: %d", rc)
	}
	return unsafe.Pointer(cVM), unsafe.Pointer(cEnv), nil
}
