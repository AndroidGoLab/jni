#include <jni.h>
#include <stdio.h>

// Forward declaration for the Go server entry point.
extern void runServer(JavaVM* vm);

// JNI_OnLoad is called when the shared library is loaded by System.load().
// It delegates to the Go runServer function which starts the gRPC server
// and blocks until termination.
JNIEXPORT jint JNI_OnLoad(JavaVM* vm, void* reserved) {
    (void)reserved;
    fprintf(stderr, "jniservice: JNI_OnLoad called\n");
    runServer(vm);
    return JNI_VERSION_1_6;
}
