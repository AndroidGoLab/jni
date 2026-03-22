GOGIO          := $(HOME)/go/bin/gogio
ANDROID_SERIAL ?= $(ANDROID_SERIAL)

APK      := gio.apk
APPID    := center.dx.jni.examples.gio
ACTIVITY := org.gioui.GioActivity

# Use the Vulkan backend (-tags noopengl) because Gio v0.9.0's OpenGL ES
# path hits EGL_BAD_SURFACE on recent Android versions (tested on Pixel 8a,
# Android 16 / API 36).
.PHONY: build install run clean

build: $(APK)

$(APK):
	$(GOGIO) -target android -appid $(APPID) \
		-minsdk 24 -targetsdk 33 \
		-tags noopengl \
		-o $@ .

install: $(APK)
	adb install -r $(APK)

run: install
	adb shell am start -n $(APPID)/$(ACTIVITY)

clean:
	rm -f $(APK)
