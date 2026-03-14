//go:build android

// JNI gRPC proxy server for Android devices.
//
// This binary is compiled as a c-shared library and loaded by the
// JNIService Java class via app_process. When the shared library is
// loaded, JNI_OnLoad (in jni_onload.c) calls runServer, which obtains
// the Android system Context via ActivityThread reflection and starts
// a gRPC server exposing all Android API services and the raw JNI surface.
//
// Configuration is via environment variables:
//
//	JNISERVICE_PORT   TCP port (default "50051")
//	JNISERVICE_LISTEN Listen address (default "0.0.0.0")
//	JNISERVICE_TOKEN  Bearer token for auth (empty = no auth)
package main

/*
#include <jni.h>
*/
import "C"
import (
	"fmt"
	"net"
	"os"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/app"
	"github.com/xaionaro-go/jni/grpc/server"
	jnirawserver "github.com/xaionaro-go/jni/grpc/server/jni_raw"
	"github.com/xaionaro-go/jni/handlestore"
	jnirawpb "github.com/xaionaro-go/jni/proto/jni_raw"
	"google.golang.org/grpc"
)

//export runServer
func runServer(cvm *C.JavaVM) {
	vm := jni.VMFromPtr(unsafe.Pointer(cvm))

	listenAddr := os.Getenv("JNISERVICE_LISTEN")
	if listenAddr == "" {
		listenAddr = "0.0.0.0"
	}

	port := os.Getenv("JNISERVICE_PORT")
	if port == "" {
		port = "50051"
	}

	token := os.Getenv("JNISERVICE_TOKEN")

	// Obtain the Android system Context via ActivityThread, the same
	// technique used by the E2E test suite to run inside app_process.
	appCtx, err := getSystemContext(vm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jniservice: get system context: %v\n", err)
		os.Exit(1)
	}

	handles := handlestore.New()

	// Build the auth interceptor chain.
	var auth server.Authorizer
	switch {
	case token != "":
		auth = server.TokenAuth{Token: token}
	default:
		auth = server.NoAuth{}
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(server.UnaryAuthInterceptor(auth)),
		grpc.ChainStreamInterceptor(server.StreamAuthInterceptor(auth)),
	)

	// Register all generated Android API service servers.
	server.RegisterAll(grpcServer, appCtx, handles)

	// Register the raw JNI service for low-level JNI access.
	jnirawpb.RegisterJNIServiceServer(grpcServer, &jnirawserver.Server{
		VM:      vm,
		Handles: handles,
	})

	addr := net.JoinHostPort(listenAddr, port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jniservice: listen %s: %v\n", addr, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "jniservice: listening on %s\n", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "jniservice: serve: %v\n", err)
		os.Exit(1)
	}
}

// getSystemContext obtains an Android Context via ActivityThread reflection.
// This is the standard technique for running inside app_process without a
// full Activity/Application lifecycle. It mirrors the pattern used in
// tests/e2e/e2e.go.
func getSystemContext(vm *jni.VM) (*app.Context, error) {
	var ctx app.Context
	ctx.VM = vm

	err := vm.Do(func(env *jni.Env) error {
		if err := app.Init(env); err != nil {
			return err
		}

		atClass, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread: %w", err)
		}

		// Try currentActivityThread() first (returns existing thread or null).
		currentMid, err := env.GetStaticMethodID(atClass, "currentActivityThread", "()Landroid/app/ActivityThread;")
		if err != nil {
			return fmt.Errorf("get currentActivityThread: %w", err)
		}
		atObj, _ := env.CallStaticObjectMethod(atClass, currentMid)

		if atObj == nil || atObj.Ref() == 0 {
			// Prepare Looper before creating ActivityThread.
			looperClass, err := env.FindClass("android/os/Looper")
			if err != nil {
				return fmt.Errorf("find Looper: %w", err)
			}
			prepMid, err := env.GetStaticMethodID(looperClass, "prepareMainLooper", "()V")
			if err != nil {
				return fmt.Errorf("get prepareMainLooper: %w", err)
			}
			// Ignore error: the main looper may already be prepared.
			_ = env.CallStaticVoidMethod(looperClass, prepMid)

			// Create ActivityThread via systemMain().
			sysMid, err := env.GetStaticMethodID(atClass, "systemMain", "()Landroid/app/ActivityThread;")
			if err != nil {
				return fmt.Errorf("get systemMain: %w", err)
			}
			atObj, err = env.CallStaticObjectMethod(atClass, sysMid)
			if err != nil {
				return fmt.Errorf("call systemMain: %w", err)
			}
		}

		getCtxMid, err := env.GetMethodID(atClass, "getSystemContext", "()Landroid/app/ContextImpl;")
		if err != nil {
			return fmt.Errorf("get getSystemContext: %w", err)
		}
		sysCtxObj, err := env.CallObjectMethod(atObj, getCtxMid)
		if err != nil {
			return fmt.Errorf("call getSystemContext: %w", err)
		}

		// Create a package context for com.android.shell (matches shell uid 2000
		// and has location/network permissions).
		ctxClass, err := env.FindClass("android/content/Context")
		if err != nil {
			return fmt.Errorf("find Context class: %w", err)
		}
		createPkgCtxMid, err := env.GetMethodID(ctxClass, "createPackageContext", "(Ljava/lang/String;I)Landroid/content/Context;")
		if err != nil {
			return fmt.Errorf("get createPackageContext: %w", err)
		}
		pkgName, err := env.NewStringUTF("com.android.shell")
		if err != nil {
			return fmt.Errorf("new string: %w", err)
		}
		shellCtxObj, err := env.CallObjectMethod(sysCtxObj, createPkgCtxMid,
			jni.ObjectValue(&pkgName.Object), jni.IntValue(0))
		if err != nil {
			// Fall back to system context if createPackageContext fails.
			ctx.Obj = env.NewGlobalRef(sysCtxObj)
			return nil
		}

		ctx.Obj = env.NewGlobalRef(shellCtxObj)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}

func main() {} // Required for c-shared build mode.
