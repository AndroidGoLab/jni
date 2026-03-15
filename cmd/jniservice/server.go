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
//	JNISERVICE_MTLS   Set to "1" to enable mTLS with ACL-based auth
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
	mtlsEnabled := os.Getenv("JNISERVICE_MTLS") == "1"

	handles := handlestore.New()

	// Initialize Android system context (Looper + ActivityThread).
	// This makes the Context handle available for Android API calls.
	initAndroidContext(vm, handles)

	// Build the auth interceptor chain and server options.
	var (
		auth    server.Authorizer
		opts    []grpc.ServerOption
		ca      *certauth.CA
		aclStore *acl.Store
	)

	switch {
	case mtlsEnabled:
		var err error
		ca, err = certauth.LoadOrCreateCA("/data/local/tmp/jniservice-ca/")
		if err != nil {
			fmt.Fprintf(os.Stderr, "jniservice: load/create CA: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "jniservice: CA loaded\n")

		aclStore, err = acl.OpenStore("/data/local/tmp/jniservice.db")
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

		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
		auth = server.ACLAuth{Store: aclStore}
		fmt.Fprintf(os.Stderr, "jniservice: mTLS enabled\n")

	case token != "":
		auth = server.TokenAuth{Token: token}

	default:
		auth = server.NoAuth{}
	}

	opts = append(opts,
		grpc.ChainUnaryInterceptor(server.UnaryAuthInterceptor(auth)),
		grpc.ChainStreamInterceptor(server.StreamAuthInterceptor(auth)),
	)

	grpcServer := grpc.NewServer(opts...)

	// Register handle store + any available Android API services.
	server.RegisterAll(grpcServer, vm, handles)
	fmt.Fprintf(os.Stderr, "jniservice: registered handlestore\n")

	// Register AuthService when mTLS is enabled.
	if mtlsEnabled {
		authpb.RegisterAuthServiceServer(grpcServer, &server.AuthServiceServer{
			CA:    ca,
			Store: aclStore,
		})
		fmt.Fprintf(os.Stderr, "jniservice: registered auth service\n")
	}

	// Register the raw JNI service for low-level JNI access.
	jnirawpb.RegisterJNIServiceServer(grpcServer, &jnirawserver.Server{
		VM:      vm,
		Handles: handles,
	})
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

// initAndroidContext initializes the Android system context by calling
// Looper.prepare() and ActivityThread.systemMain() on a pinned OS thread.
// The resulting Context handle is stored in the HandleStore and printed
// to stderr for CLI tools to reference.
func initAndroidContext(vm *jni.VM, handles *handlestore.HandleStore) {
	// Pin to OS thread so Looper.prepare() and systemMain() run on the same thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var contextHandle int64
	err := vm.Do(func(env *jni.Env) error {
		// Looper.prepare() — required before creating an ActivityThread.
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

		// ActivityThread.systemMain() — creates a full ActivityThread with system context.
		atCls, err := env.FindClass("android/app/ActivityThread")
		if err != nil {
			return fmt.Errorf("find ActivityThread class: %w", err)
		}
		systemMainMID, err := env.GetStaticMethodID(atCls, "systemMain", "()Landroid/app/ActivityThread;")
		if err != nil {
			return fmt.Errorf("get systemMain: %w", err)
		}
		atObj, err := env.CallStaticObjectMethod(atCls, systemMainMID)
		if err != nil {
			return fmt.Errorf("systemMain: %w", err)
		}

		// ActivityThread.getSystemContext() → ContextImpl (which IS a Context).
		getCtxMID, err := env.GetMethodID(atCls, "getSystemContext", "()Landroid/app/ContextImpl;")
		if err != nil {
			return fmt.Errorf("get getSystemContext: %w", err)
		}
		ctxObj, err := env.CallObjectMethod(atObj, getCtxMID)
		if err != nil {
			return fmt.Errorf("getSystemContext: %w", err)
		}

		contextHandle = handles.Put(env, ctxObj)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "jniservice: WARNING: context init failed: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "jniservice: android context initialized (handle=%d)\n", contextHandle)
}

func main() {} // Required for c-shared build mode.
