# Shared APK build rules for JNI examples.
#
# Usage: set EXAMPLE_NAME and optionally EXAMPLE_PERMISSIONS, then include.
#
#   EXAMPLE_NAME        := location
#   EXAMPLE_PERMISSIONS := android.permission.ACCESS_FINE_LOCATION
#   include ../apk.mk
#
# Targets:
#   make build   — produce build/$(EXAMPLE_NAME).apk (multi-arch: arm64 + x86_64)
#   make install — adb install
#   make run     — install + launch
#   make clean   — remove build artifacts

# ---------------------------------------------------------------------------
# SDK / NDK discovery
# ---------------------------------------------------------------------------
ANDROID_SDK ?= $(shell \
	if [ -n "$$ANDROID_HOME" ]; then echo "$$ANDROID_HOME"; \
	elif [ -d "$$HOME/Android/Sdk" ]; then echo "$$HOME/Android/Sdk"; \
	elif [ -d "$$HOME/android-sdk" ]; then echo "$$HOME/android-sdk"; \
	fi)

NDK_VERSION  ?= $(shell ls $(ANDROID_SDK)/ndk 2>/dev/null | sort -V | tail -1)
NDK          ?= $(ANDROID_SDK)/ndk/$(NDK_VERSION)
BUILD_TOOLS  ?= $(shell ls $(ANDROID_SDK)/build-tools 2>/dev/null | sort -V | tail -1)
PLATFORM     ?= $(shell ls -d $(ANDROID_SDK)/platforms/android-* 2>/dev/null | sort -V | tail -1)
MIN_SDK      ?= 24
TARGET_SDK   ?= $(shell basename $(PLATFORM) | sed 's/android-//')

ADB       := $(ANDROID_SDK)/platform-tools/adb
AAPT2     := $(ANDROID_SDK)/build-tools/$(BUILD_TOOLS)/aapt2
D8        := $(ANDROID_SDK)/build-tools/$(BUILD_TOOLS)/d8
ZIPALIGN  := $(ANDROID_SDK)/build-tools/$(BUILD_TOOLS)/zipalign
APKSIGNER := $(ANDROID_SDK)/build-tools/$(BUILD_TOOLS)/apksigner

# Helpers for string manipulation.
space := $(subst ,, )
comma := ,

NDK_TOOLCHAIN := $(NDK)/toolchains/llvm/prebuilt/linux-x86_64/bin

CC_ARM64 := $(NDK_TOOLCHAIN)/aarch64-linux-android$(MIN_SDK)-clang
CC_AMD64 := $(NDK_TOOLCHAIN)/x86_64-linux-android$(MIN_SDK)-clang

# ---------------------------------------------------------------------------
# Paths
# ---------------------------------------------------------------------------
BUILD        := build
SHARED_DIR   := ../apk
HANDLER_DIR  := ../../internal/testjvm/testdata
PACKAGE_NAME ?= com.github.xaionaro_go.jni.examples.$(EXAMPLE_NAME)

.PHONY: all build install run clean

all: build

build: $(BUILD)/$(EXAMPLE_NAME).apk

# ---------------------------------------------------------------------------
# Debug keystore (generated once per build dir)
# ---------------------------------------------------------------------------
$(BUILD)/debug.keystore:
	@mkdir -p $(BUILD)
	keytool -genkeypair -keystore $@ -storepass android -alias debug \
		-keyalg RSA -keysize 2048 -validity 10000 \
		-dname "CN=Debug" -noprompt 2>/dev/null

# ---------------------------------------------------------------------------
# Copy shared jni_onload.c (CGo needs it in the Go package directory)
# ---------------------------------------------------------------------------
jni_onload.c: $(SHARED_DIR)/jni_onload.c
	cp $< $@

# ---------------------------------------------------------------------------
# AndroidManifest.xml (generated from EXAMPLE_PERMISSIONS)
# ---------------------------------------------------------------------------
$(BUILD)/AndroidManifest.xml:
	@mkdir -p $(BUILD)
	@printf '<?xml version="1.0" encoding="utf-8"?>\n' > $@
	@printf '<manifest xmlns:android="http://schemas.android.com/apk/res/android"\n' >> $@
	@printf '    package="%s">\n' '$(PACKAGE_NAME)' >> $@
	@$(foreach perm,$(EXAMPLE_PERMISSIONS), \
		printf '    <uses-permission android:name="%s" />\n' '$(perm)' >> $@;)
	@printf '    <application android:label="%s">\n' '$(EXAMPLE_NAME)' >> $@
	@PERM_CSV="$(subst $(space),$(comma),$(EXAMPLE_PERMISSIONS))"; \
		if [ -n "$$PERM_CSV" ]; then \
			printf '        <meta-data android:name="example.permissions" android:value="%s" />\n' "$$PERM_CSV" >> $@; \
		fi
	@printf '        <activity android:name="com.github.xaionaro_go.jni.example.ExampleActivity"\n' >> $@
	@printf '                  android:exported="true">\n' >> $@
	@printf '            <intent-filter>\n' >> $@
	@printf '                <action android:name="android.intent.action.MAIN" />\n' >> $@
	@printf '                <category android:name="android.intent.category.LAUNCHER" />\n' >> $@
	@printf '            </intent-filter>\n' >> $@
	@printf '        </activity>\n' >> $@
	@printf '    </application>\n' >> $@
	@printf '</manifest>\n' >> $@

# ---------------------------------------------------------------------------
# Java → DEX
# ---------------------------------------------------------------------------
$(BUILD)/classes.dex: $(SHARED_DIR)/ExampleActivity.java $(HANDLER_DIR)/com/github/xaionaro_go/jni/internal/GoInvocationHandler.java
	@mkdir -p $(BUILD)/java
	javac --release 17 -classpath $(PLATFORM)/android.jar \
		-d $(BUILD)/java \
		$(SHARED_DIR)/ExampleActivity.java \
		$(HANDLER_DIR)/com/github/xaionaro_go/jni/internal/GoInvocationHandler.java
	$(D8) --lib $(PLATFORM)/android.jar --output $(BUILD) \
		$$(find $(BUILD)/java -name '*.class')

# ---------------------------------------------------------------------------
# Go → c-shared libraries (arm64 + x86_64)
# ---------------------------------------------------------------------------
$(BUILD)/lib/arm64-v8a/libexample.so: main.go jni_onload.c
	@mkdir -p $(dir $@)
	cd ../.. && CGO_ENABLED=1 GOOS=android GOARCH=arm64 CC=$(CC_ARM64) \
		go build -buildmode=c-shared \
		-o examples/$(EXAMPLE_NAME)/$@ \
		./examples/$(EXAMPLE_NAME)/
	@rm -f $(@:.so=.h)

$(BUILD)/lib/x86_64/libexample.so: main.go jni_onload.c
	@mkdir -p $(dir $@)
	cd ../.. && CGO_ENABLED=1 GOOS=android GOARCH=amd64 CC=$(CC_AMD64) \
		go build -buildmode=c-shared \
		-o examples/$(EXAMPLE_NAME)/$@ \
		./examples/$(EXAMPLE_NAME)/
	@rm -f $(@:.so=.h)

# ---------------------------------------------------------------------------
# Package, align, sign → APK (multi-arch)
# ---------------------------------------------------------------------------
$(BUILD)/$(EXAMPLE_NAME).apk: $(BUILD)/AndroidManifest.xml $(BUILD)/classes.dex $(BUILD)/lib/arm64-v8a/libexample.so $(BUILD)/lib/x86_64/libexample.so $(BUILD)/debug.keystore
	$(AAPT2) link --manifest $(BUILD)/AndroidManifest.xml \
		-I $(PLATFORM)/android.jar \
		--min-sdk-version $(MIN_SDK) \
		--target-sdk-version $(TARGET_SDK) \
		-o $(BUILD)/base.apk
	cd $(BUILD) && zip -j base.apk classes.dex
	cd $(BUILD) && zip -r base.apk lib/
	$(ZIPALIGN) -f 4 $(BUILD)/base.apk $(BUILD)/aligned.apk
	$(APKSIGNER) sign --ks $(BUILD)/debug.keystore --ks-pass pass:android \
		--out $@ $(BUILD)/aligned.apk
	@rm -f $(BUILD)/base.apk $(BUILD)/aligned.apk
	@echo "Built: $@"

# ---------------------------------------------------------------------------
# Install & run
# ---------------------------------------------------------------------------
install: $(BUILD)/$(EXAMPLE_NAME).apk
	$(ADB) install -r $<

run: install
	@$(foreach perm,$(EXAMPLE_PERMISSIONS), \
		$(ADB) shell pm grant $(PACKAGE_NAME) $(perm) 2>/dev/null || true;)
	$(ADB) shell am start -n $(PACKAGE_NAME)/com.github.xaionaro_go.jni.example.ExampleActivity

# ---------------------------------------------------------------------------
# Clean
# ---------------------------------------------------------------------------
clean:
	rm -rf $(BUILD) jni_onload.c
