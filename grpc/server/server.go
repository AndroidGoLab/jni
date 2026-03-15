package server

import (
	"github.com/xaionaro-go/jni"
	"github.com/xaionaro-go/jni/handlestore"
	handlepb "github.com/xaionaro-go/jni/proto/handlestore"
	"google.golang.org/grpc"
)

// RegisterAll registers all available gRPC service servers.
// Currently only registers the HandleStore service. Android API service
// servers will be added once javagen produces matching Go wrapper packages.
func RegisterAll(s grpc.ServiceRegistrar, vm *jni.VM, handles *handlestore.HandleStore) {
	handlepb.RegisterHandleStoreServiceServer(s, &handlestore.Server{VM: vm, Handles: handles})
}
