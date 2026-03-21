#include <jni.h>
#include <android/native_activity.h>

// Go entry point.
extern void goOnCreate(JavaVM* vm, jobject activity);

// Called by Android when the NativeActivity is created.
// activity->clazz is a global ref to the Activity instance.
void ANativeActivity_onCreate(ANativeActivity* activity, void* savedState, size_t savedStateSize) {
    (void)savedState;
    (void)savedStateSize;
    goOnCreate(activity->vm, activity->clazz);
}
