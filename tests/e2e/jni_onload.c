#include <jni.h>
#include <stdio.h>

// Forward declaration for Go test entry point.
extern void runE2ETests(JavaVM* vm);

// JNI_OnLoad is called when the shared library is loaded by System.load().
JNIEXPORT jint JNI_OnLoad(JavaVM* vm, void* reserved) {
    fprintf(stderr, "JNI_OnLoad called\n");
    runE2ETests(vm);
    return JNI_VERSION_1_6;
}
