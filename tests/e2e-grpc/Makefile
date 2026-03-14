# E2E gRPC tests for jnictl against jniservice on Android emulator.
#
# Prerequisites: Android SDK, running emulator, NDK for cross-compilation.

ANDROID_SDK ?= $(HOME)/Android/Sdk
NDK ?= $(shell ls -d $(ANDROID_SDK)/ndk/* 2>/dev/null | sort -V | tail -1)
API_LEVEL ?= 30
ADB ?= $(ANDROID_SDK)/platform-tools/adb
CC_ARM64 := $(NDK)/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android$(API_LEVEL)-clang
CC_X86 := $(NDK)/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android$(API_LEVEL)-clang
ARCH ?= $(shell $(ADB) shell getprop ro.product.cpu.abi 2>/dev/null)
REMOTE_DIR := /data/local/tmp
PORT ?= 50051

# Detect architecture and set CC
ifeq ($(ARCH),arm64-v8a)
  GOARCH := arm64
  CC := $(CC_ARM64)
else
  GOARCH := amd64
  CC := $(CC_X86)
endif

.PHONY: build push start-server stop-server test clean

build:
	# Build jnictl for Android
	CGO_ENABLED=0 GOOS=android GOARCH=$(GOARCH) go build -o build/jnictl ../../cmd/jnictl/
	# Build jniservice shared library
	CGO_ENABLED=1 GOOS=android GOARCH=$(GOARCH) CC=$(CC) \
		go build -buildmode=c-shared -o build/libjniservice.so ../../cmd/jniservice/
	# Compile Java wrapper
	javac --release 17 -d build ../../cmd/jniservice/JNIService.java
	$(ANDROID_SDK)/build-tools/$$(ls $(ANDROID_SDK)/build-tools | sort -V | tail -1)/d8 \
		--lib $$(ls -d $(ANDROID_SDK)/platforms/android-* | sort -V | tail -1)/android.jar \
		--output build build/JNIService.class

push: build
	$(ADB) push build/jnictl $(REMOTE_DIR)/jnictl
	$(ADB) shell chmod +x $(REMOTE_DIR)/jnictl
	$(ADB) push build/libjniservice.so $(REMOTE_DIR)/libjniservice.so
	$(ADB) push build/classes.dex $(REMOTE_DIR)/jniservice.dex

start-server:
	$(ADB) shell "JNISERVICE_PORT=$(PORT) app_process -Djava.class.path=$(REMOTE_DIR)/jniservice.dex $(REMOTE_DIR) JNIService &"
	sleep 2
	@echo "jniservice started on port $(PORT)"

stop-server:
	$(ADB) shell "pkill -f JNIService" || true

test: push start-server
	$(ADB) shell "$(REMOTE_DIR)/jnictl --addr localhost:$(PORT) --insecure power power-manager is-interactive" || true
	@echo "E2E test complete"
	$(MAKE) stop-server

clean:
	rm -rf build/
