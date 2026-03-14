# E2E test harness for JNI Go bindings on Android.
# Requires: Android SDK (with build-tools, platform-tools, NDK), running emulator.

ANDROID_SDK ?= $(HOME)/android-sdk
NDK_VERSION ?= $(shell ls $(ANDROID_SDK)/ndk 2>/dev/null | sort -V | tail -1)
NDK         ?= $(ANDROID_SDK)/ndk/$(NDK_VERSION)
API_LEVEL   ?= 30
BUILD_TOOLS ?= $(shell ls $(ANDROID_SDK)/build-tools 2>/dev/null | sort -V | tail -1)
PLATFORM    ?= $(shell ls -d $(ANDROID_SDK)/platforms/android-* 2>/dev/null | sort -V | tail -1)

ADB     := $(ANDROID_SDK)/platform-tools/adb
D8      := $(ANDROID_SDK)/build-tools/$(BUILD_TOOLS)/d8
CC_AMD64 := $(NDK)/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android$(API_LEVEL)-clang
CC_ARM64 := $(NDK)/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android$(API_LEVEL)-clang

# Detect emulator architecture.
ARCH ?= $(shell $(ADB) shell getprop ro.product.cpu.abi 2>/dev/null)

ifeq ($(ARCH),arm64-v8a)
  GOARCH := arm64
  CC     := $(CC_ARM64)
else ifeq ($(ARCH),x86_64)
  GOARCH := amd64
  CC     := $(CC_AMD64)
else
  GOARCH := amd64
  CC     := $(CC_AMD64)
endif

HANDLER_JAVA := $(shell find ../../internal/testjvm/testdata -name 'GoInvocationHandler.java' 2>/dev/null | head -1)
BUILD_DIR    := build

.PHONY: all build push run clean

all: run

build: $(BUILD_DIR)/libe2etest.so $(BUILD_DIR)/classes.dex

$(BUILD_DIR)/classes.dex: TestRunner.java $(HANDLER_JAVA)
	@mkdir -p $(BUILD_DIR)/com/github/xaionaro_go/jni/internal
	@cp $(HANDLER_JAVA) $(BUILD_DIR)/com/github/xaionaro_go/jni/internal/
	javac --release 17 -d $(BUILD_DIR) TestRunner.java $(BUILD_DIR)/com/github/xaionaro_go/jni/internal/GoInvocationHandler.java
	$(D8) --lib $(PLATFORM)/android.jar --output $(BUILD_DIR) \
		$(BUILD_DIR)/TestRunner.class $(BUILD_DIR)/com/github/xaionaro_go/jni/internal/GoInvocationHandler.class

$(BUILD_DIR)/libe2etest.so: e2e.go jni_onload.c
	cd ../.. && CGO_ENABLED=1 GOOS=android GOARCH=$(GOARCH) CC=$(CC) \
		go build -buildmode=c-shared -o tests/e2e/$(BUILD_DIR)/libe2etest.so ./tests/e2e/

push: build
	$(ADB) push $(BUILD_DIR)/classes.dex /data/local/tmp/test.dex
	$(ADB) push $(BUILD_DIR)/libe2etest.so /data/local/tmp/libe2etest.so

run: push
	$(ADB) shell 'app_process -Djava.class.path=/data/local/tmp/test.dex /data/local/tmp TestRunner' 2>&1 | tee /dev/stderr | grep -q E2E_PASS

clean:
	rm -rf $(BUILD_DIR)
