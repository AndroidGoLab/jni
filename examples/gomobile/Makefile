GOMOBILE       := $(HOME)/go/bin/gomobile
ANDROID_NDK_HOME ?= $(HOME)/Android/Sdk/ndk/28.0.13004108
BUILD_TOOLS    := $(HOME)/Android/Sdk/build-tools/35.0.0
ANDROID_JAR    := $(HOME)/Android/Sdk/platforms/android-35/android.jar
ANDROID_SERIAL ?= $(ANDROID_SERIAL)

# gomobile hardcodes targetSdkVersion=16 in the manifest, which modern
# Android rejects (minimum 24).  The build target patches the APK manifest
# to set minSdk=24 / targetSdk=24, then re-signs with the debug key.
APK       := gomobile.apk
RAW_APK   := gomobile-raw.apk
PACKAGE   := org.golang.todo.gomobile
ACTIVITY  := org.golang.app.GoNativeActivity

.PHONY: build install run clean

build: $(APK)

$(RAW_APK):
	ANDROID_NDK_HOME=$(ANDROID_NDK_HOME) $(GOMOBILE) build \
		-target=android/arm64 -androidapi 24 -o $@ .

$(APK): $(RAW_APK)
	$(eval WORK := $(shell mktemp -d))
	# Extract the original APK (classes.dex + native libs).
	unzip -q -o $(RAW_APK) -d $(WORK)/extracted
	# Write a corrected AndroidManifest.xml.
	printf '%s\n' \
	  '<?xml version="1.0" encoding="utf-8"?>' \
	  '<manifest xmlns:android="http://schemas.android.com/apk/res/android"' \
	  '    package="$(PACKAGE)"' \
	  '    android:versionCode="1"' \
	  '    android:versionName="1.0">' \
	  '  <uses-sdk android:minSdkVersion="24" android:targetSdkVersion="24" />' \
	  '  <application android:label="Gomobile" android:debuggable="true">' \
	  '    <activity' \
	  '        android:label="Gomobile"' \
	  '        android:name="$(ACTIVITY)"' \
	  '        android:configChanges="orientation|keyboardHidden">' \
	  '      <meta-data android:name="android.app.lib_name" android:value="gomobile" />' \
	  '      <intent-filter>' \
	  '        <action android:name="android.intent.action.MAIN" />' \
	  '        <category android:name="android.intent.category.LAUNCHER" />' \
	  '      </intent-filter>' \
	  '    </activity>' \
	  '  </application>' \
	  '</manifest>' > $(WORK)/AndroidManifest.xml
	# Create an APK with the corrected manifest.
	$(BUILD_TOOLS)/aapt package -f \
		-M $(WORK)/AndroidManifest.xml \
		-I $(ANDROID_JAR) \
		-F $(WORK)/base.apk
	# Add classes.dex and native libraries.
	cd $(WORK)/extracted && zip -q $(WORK)/base.apk \
		classes.dex lib/arm64-v8a/libgomobile.so
	# Align and sign.
	$(BUILD_TOOLS)/zipalign -f 4 $(WORK)/base.apk $(WORK)/aligned.apk
	$(BUILD_TOOLS)/apksigner sign \
		--ks $(HOME)/.android/debug.keystore \
		--ks-pass pass:android \
		--key-pass pass:android \
		--ks-key-alias androiddebugkey \
		--out $@ $(WORK)/aligned.apk
	rm -rf $(WORK)
	@echo "=== APK ready: $@ ==="
	@$(BUILD_TOOLS)/aapt dump badging $@ 2>/dev/null | grep -E 'sdk|package'

install: $(APK)
	adb install -r $(APK)

run: install
	adb shell am start -n $(PACKAGE)/$(ACTIVITY)

clean:
	rm -f $(APK) $(RAW_APK)
