EXAMPLE_NAME := camera_display
EXAMPLE_PERMISSIONS := android.permission.CAMERA
EXAMPLE_NEEDS_PROXY := true
EXAMPLE_EXTRA_JAVA := $(wildcard ../common/java/center/dx/jni/generated/*.java)
# ANativeWindow_toSurface requires API level 26+.
MIN_SDK := 26

include ../apk.mk
