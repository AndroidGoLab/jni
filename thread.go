package jni

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/AndroidGoLab/jni/capi"
)

// Do pins the calling goroutine to an OS thread, attaches it to the JVM
// (if not already attached), executes fn with a valid *Env, and detaches
// if it was the one that attached.
//
// This is the primary entry point for all JNI operations. It ensures:
//  1. The goroutine stays on the same OS thread for the duration of fn.
//  2. The OS thread has a valid JNIEnv (attaching if necessary).
//  3. The thread is detached after fn returns (only if Do attached it).
//
// It is safe to nest Do calls. An inner Do finds the thread already
// attached via GetEnv and neither re-attaches nor detaches. Go
// reference-counts LockOSThread/UnlockOSThread, so nesting is safe.
func (vm *VM) Do(fn func(env *Env) error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	env, attached, err := vm.getOrAttachCurrentThread()
	if err != nil {
		return err
	}
	if attached {
		defer func() { _ = vm.detachCurrentThread() }()
	}
	return fn(env)
}

// GetEnv returns the JNIEnv for the current thread if it is already
// attached to the JVM. Returns ErrDetached if the thread is not attached.
func (vm *VM) GetEnv(version capi.Jint) (*Env, error) {
	var envPtr unsafe.Pointer
	rc := capi.GetEnv(vm.ptr, &envPtr, version)
	if rc == capi.JNI_OK {
		return &Env{ptr: (*capi.Env)(envPtr)}, nil
	}
	if rc == capi.JNI_EDETACHED {
		return nil, ErrDetached
	}
	return nil, Error(rc)
}

// AttachCurrentThread attaches the current OS thread to the JVM and
// returns a valid JNIEnv pointer.
func (vm *VM) AttachCurrentThread() (*Env, error) {
	var envPtr *capi.Env
	rc := capi.AttachCurrentThread(vm.ptr, &envPtr, unsafe.Pointer(nil))
	if rc != capi.JNI_OK {
		return nil, Error(rc)
	}
	return &Env{ptr: envPtr}, nil
}

// detachCurrentThread detaches the current OS thread from the JVM.
func (vm *VM) detachCurrentThread() error {
	rc := capi.DetachCurrentThread(vm.ptr)
	if rc != capi.JNI_OK {
		return Error(rc)
	}
	return nil
}

// getOrAttachCurrentThread tries GetEnv first; if the current thread is
// not attached (JNI_EDETACHED), calls AttachCurrentThread.
// Returns (env, wasAttachedByUs, error).
func (vm *VM) getOrAttachCurrentThread() (*Env, bool, error) {
	env, err := vm.GetEnv(JNI_VERSION_1_6)
	if err == nil {
		return env, false, nil
	}
	if err != ErrDetached {
		return nil, false, fmt.Errorf("jni: GetEnv: %w", err)
	}
	env, err = vm.AttachCurrentThread()
	if err != nil {
		return nil, false, fmt.Errorf("jni: AttachCurrentThread: %w", err)
	}
	return env, true, nil
}
