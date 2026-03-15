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
package main

/*
#include <jni.h>
*/
import "C"
import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/grpc/server"
	"github.com/xaionaro-go/jni/grpc/server/acl"
	"github.com/xaionaro-go/jni/grpc/server/certauth"
	jnirawserver "github.com/xaionaro-go/jni/grpc/server/jni_raw"
	"github.com/xaionaro-go/jni/handlestore"
	authpb "github.com/xaionaro-go/jni/proto/auth"
	jnirawpb "github.com/xaionaro-go/jni/proto/jni_raw"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// globalVM, globalHandles, and globalJNIServer are set during runServer so that
// setAppContext (called later from Java) can store the APK's Application context
// and ClassLoader.
var (
	globalVM        *jni.VM
	globalHandles   *handlestore.HandleStore
	globalJNIServer *jnirawserver.Server
)

// Java_center_dx_jni_jniservice_JNIServiceForeground_setAppContext is the JNI-mangled name
// for center.dx.jni.jniservice.JNIServiceForeground.setAppContext(Object).
//
//export Java_center_dx_jni_jniservice_JNIServiceForeground_setAppContext
func Java_center_dx_jni_jniservice_JNIServiceForeground_setAppContext(cenv *C.JNIEnv, cls C.jclass, ctx C.jobject) {
	// Called from Java after System.loadLibrary in APK mode.
	// Stores the APK's Application context (with the correct ClassLoader
	// and permissions) in the HandleStore.
	if globalVM == nil || globalHandles == nil {
		fmt.Fprintf(os.Stderr, "jniservice: setAppContext called before runServer\n")
		return
	}
	globalVM.Do(func(goEnv *jni.Env) error {
		// ctx is C.jobject from this package's CGO. jni.ObjectFromRef takes
		// capi.Object (= C.jobject from capi's CGO). Both are C pointer types
		// so unsafe conversion is safe.
		obj := jni.ObjectFromRef(*(*jni.CAPIObject)(unsafe.Pointer(&ctx)))
		handle := globalHandles.Put(goEnv, obj)
		fmt.Fprintf(os.Stderr, "jniservice: APK app context stored (handle=%d)\n", handle)

		// Also store the APK's ClassLoader so FindClass can fall back to it
		// for APK-specific classes (JNI FindClass from native threads uses
		// BootClassLoader which can't see APK classes).
		ctxCls, err := goEnv.FindClass("android/content/Context")
		if err != nil {
			fmt.Fprintf(os.Stderr, "jniservice: WARNING: can't find Context class: %v\n", err)
			return nil
		}
		getClMID, err := goEnv.GetMethodID(ctxCls, "getClassLoader", "()Ljava/lang/ClassLoader;")
		if err != nil {
			fmt.Fprintf(os.Stderr, "jniservice: WARNING: can't find getClassLoader: %v\n", err)
			return nil
		}
		clObj, err := goEnv.CallObjectMethod(obj, getClMID)
		if err != nil || clObj == nil || clObj.Ref() == 0 {
			fmt.Fprintf(os.Stderr, "jniservice: WARNING: getClassLoader failed: %v\n", err)
			return nil
		}
		clHandle := globalHandles.Put(goEnv, clObj)
		if globalJNIServer != nil {
			globalJNIServer.AppClassLoader = clHandle
		}
		fmt.Fprintf(os.Stderr, "jniservice: APK ClassLoader stored (handle=%d)\n", clHandle)

		return nil
	})
}

//export runServer
func runServer(cvm *C.JavaVM) {
	vm := jni.VMFromPtr(unsafe.Pointer(cvm))

	// In APK mode, Go's os.Stderr goes to /dev/null. Redirect to a file
	// so we can debug startup failures.
	if logFile, err := os.OpenFile("/data/data/center.dx.jni.jniservice/cache/jniservice.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
		os.Stderr = logFile
	}

	listenAddr := os.Getenv("JNISERVICE_LISTEN")
	if listenAddr == "" {
		listenAddr = "127.0.0.1"
	}

	port := os.Getenv("JNISERVICE_PORT")
	if port == "" {
		port = "50051"
	}

	handles := handlestore.New()

	globalVM = vm
	globalHandles = handles

	// Initialize Android system context (Looper + ActivityThread).
	// This makes the Context handle available for Android API calls.
	initAndroidContext(vm, handles)

	// Determine data directory. Try the APK's files dir first (writable by the
	// app process), fall back to /data/local/tmp (writable by shell user in
	// app_process mode).
	dataDir := os.Getenv("JNISERVICE_DATA_DIR")
	if dataDir == "" {
		candidates := []string{
			"/data/data/center.dx.jni.jniservice/files/jniservice",
			"/data/local/tmp/jniservice",
		}
		for _, dir := range candidates {
			if err := os.MkdirAll(dir, 0700); err == nil {
				dataDir = dir
				break
			}
		}
		if dataDir == "" {
			fmt.Fprintf(os.Stderr, "jniservice: no writable data directory found\n")
			os.Exit(1)
		}
	} else {
		if err := os.MkdirAll(dataDir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "jniservice: create data dir %s: %v\n", dataDir, err)
			os.Exit(1)
		}
	}
	fmt.Fprintf(os.Stderr, "jniservice: data dir: %s\n", dataDir)

	// mTLS is always enabled. The only unauthenticated RPC is Register
	// (client enrollment via CSR). All other RPCs require a valid client
	// certificate and per-method ACL grants.
	ca, err := certauth.LoadOrCreateCA(dataDir + "/ca")
	if err != nil {
		fmt.Fprintf(os.Stderr, "jniservice: load/create CA: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "jniservice: CA loaded\n")

	aclStore, err := acl.OpenStore(dataDir + "/acl.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "jniservice: open ACL store: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "jniservice: ACL store opened\n")

	serverTLS, err := generateServerTLS(ca)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jniservice: generate server TLS: %v\n", err)
		os.Exit(1)
	}

	caPool := x509.NewCertPool()
	caPool.AddCert(ca.Cert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLS},
		// VerifyClientCertIfGiven allows Register to work without a cert.
		ClientAuth: tls.VerifyClientCertIfGiven,
		ClientCAs:  caPool,
		MinVersion: tls.VersionTLS12,
	}

	auth := server.ACLAuth{Store: aclStore}
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewTLS(tlsConfig)),
		grpc.ChainUnaryInterceptor(server.UnaryAuthInterceptor(auth)),
		grpc.ChainStreamInterceptor(server.StreamAuthInterceptor(auth)),
	}
	fmt.Fprintf(os.Stderr, "jniservice: mTLS enabled\n")

	grpcServer := grpc.NewServer(opts...)

	// Register handle store + any available Android API services.
	server.RegisterAll(grpcServer, vm, handles)
	fmt.Fprintf(os.Stderr, "jniservice: registered handlestore\n")

	// Register AuthService (always available — Register is the unauthenticated
	// entry point for client enrollment).
	authpb.RegisterAuthServiceServer(grpcServer, &server.AuthServiceServer{
		CA:    ca,
		Store: aclStore,
	})
	fmt.Fprintf(os.Stderr, "jniservice: registered auth service\n")

	// Register the raw JNI service for low-level JNI access.
	globalJNIServer = &jnirawserver.Server{
		VM:      vm,
		Handles: handles,
	}
	jnirawpb.RegisterJNIServiceServer(grpcServer, globalJNIServer)
	fmt.Fprintf(os.Stderr, "jniservice: registered jni_raw service\n")

	svcInfo := grpcServer.GetServiceInfo()
	for name := range svcInfo {
		fmt.Fprintf(os.Stderr, "jniservice: registered service: %s\n", name)
	}

	// Enable gRPC server reflection for debugging.
	// reflection.Register(grpcServer) // uncomment if reflection package is available

	addr := net.JoinHostPort(listenAddr, port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jniservice: listen %s: %v\n", addr, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "jniservice: listening on %s\n", lis.Addr())

	// Serve in a goroutine so JNI_OnLoad can return.
	// The JVM keeps the process alive; the server goroutine handles RPCs.
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Fprintf(os.Stderr, "jniservice: serve: %v\n", err)
		}
	}()
}

// generateServerTLS creates an ephemeral server TLS certificate signed
// by the CA. The key is kept in memory only (not written to disk).
// The certificate has ExtKeyUsageServerAuth so it can be used for TLS
// server authentication.
func generateServerTLS(ca *certauth.CA) (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generating server key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generating serial: %w", err)
	}

	now := time.Now()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "jniservice"},
		NotBefore:    now,
		NotAfter:     now.Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		// SANs so clients can verify the server cert when connecting via
		// localhost, 127.0.0.1, or 0.0.0.0 (common for adb port-forward).
		DNSNames:    []string{"localhost", "jniservice"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv4(0, 0, 0, 0), net.IPv6loopback},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, ca.Cert, &key.PublicKey, ca.Key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("signing server certificate: %w", err)
	}

	serverCert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("parsing server certificate: %w", err)
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
		Leaf:        serverCert,
	}, nil
}

// initAndroidContext obtains an Android Context and stores it in the HandleStore.
//
// Two modes:
//   - APK mode: an ActivityThread already exists (the app framework created it).
//     Use currentActivityThread().getApplication() to get the app's Context.
//   - app_process mode: no ActivityThread exists. Create one via Looper.prepare()
//     + ActivityThread.systemMain(), then use getSystemContext().
func initAndroidContext(vm *jni.VM, handles *handlestore.HandleStore) {
	// Pin to OS thread so Looper/ActivityThread calls run on the same thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var contextHandle int64
	err := vm.Do(func(env *jni.Env) error {
		atCls, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread class: %w", err)
		}

		// Check if an ActivityThread already exists (APK mode).
		currentATMID, err := env.GetStaticMethodID(atCls, "currentActivityThread", "()Landroid/app/ActivityThread;")
		if err != nil {
			return fmt.Errorf("get currentActivityThread: %w", err)
		}
		atObj, err := env.CallStaticObjectMethod(atCls, currentATMID)
		if err != nil {
			return fmt.Errorf("currentActivityThread: %w", err)
		}

		switch {
		case atObj != nil && atObj.Ref() != 0:
			// APK mode: ActivityThread exists. Get the Application context.
			fmt.Fprintf(os.Stderr, "jniservice: APK mode — using existing ActivityThread\n")

			getAppMID, err := env.GetStaticMethodID(atCls, "currentApplication", "()Landroid/app/Application;")
			if err != nil {
				return fmt.Errorf("get currentApplication: %w", err)
			}
			appObj, err := env.CallStaticObjectMethod(atCls, getAppMID)
			if err != nil {
				return fmt.Errorf("currentApplication: %w", err)
			}
			if appObj == nil || appObj.Ref() == 0 {
				return fmt.Errorf("currentApplication returned null")
			}
			contextHandle = handles.Put(env, appObj)

		default:
			// app_process mode: no ActivityThread. Create one.
			fmt.Fprintf(os.Stderr, "jniservice: app_process mode — creating ActivityThread\n")

			looperCls, err := env.FindClass("android/os/Looper")
			if err != nil {
				return fmt.Errorf("find Looper class: %w", err)
			}
			prepareMID, err := env.GetStaticMethodID(looperCls, "prepare", "()V")
			if err != nil {
				return fmt.Errorf("get Looper.prepare: %w", err)
			}
			if err := env.CallStaticVoidMethod(looperCls, prepareMID); err != nil {
				return fmt.Errorf("Looper.prepare: %w", err)
			}

			systemMainMID, err := env.GetStaticMethodID(atCls, "systemMain", "()Landroid/app/ActivityThread;")
			if err != nil {
				return fmt.Errorf("get systemMain: %w", err)
			}
			atObj, err := env.CallStaticObjectMethod(atCls, systemMainMID)
			if err != nil {
				return fmt.Errorf("systemMain: %w", err)
			}

			getCtxMID, err := env.GetMethodID(atCls, "getSystemContext", "()Landroid/app/ContextImpl;")
			if err != nil {
				return fmt.Errorf("get getSystemContext: %w", err)
			}
			ctxObj, err := env.CallObjectMethod(atObj, getCtxMID)
			if err != nil {
				return fmt.Errorf("getSystemContext: %w", err)
			}
			contextHandle = handles.Put(env, ctxObj)
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "jniservice: WARNING: context init failed: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "jniservice: android context initialized (handle=%d)\n", contextHandle)
}

func main() {} // Required for c-shared build mode.
