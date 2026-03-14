#include <jni.h>

static JavaVM* gVM = NULL;

// Go entry points (defined in main.go of each example).
extern void goRun(JavaVM* vm);
extern const char* goGetOutput(void);

JNIEXPORT jint JNI_OnLoad(JavaVM* vm, void* reserved) {
    (void)reserved;
    gVM = vm;
    return JNI_VERSION_1_6;
}

// Called from ExampleActivity.onCreate — runs the example.
JNIEXPORT void JNICALL
Java_go_jni_example_ExampleActivity_nativeRun(JNIEnv* env, jobject thiz) {
    (void)env;
    (void)thiz;
    goRun(gVM);
}

// Called from ExampleActivity.onCreate — returns captured output.
JNIEXPORT jstring JNICALL
Java_go_jni_example_ExampleActivity_nativeGetOutput(JNIEnv* env, jobject thiz) {
    (void)thiz;
    const char* out = goGetOutput();
    return (*env)->NewStringUTF(env, out ? out : "");
}
